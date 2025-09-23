package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"mediapipeline/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func downloadHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file ID parameter"})
		return
	}

	// Lookup file in DB
	var (
		filename    string
		path        string
		storageTier string
	)
	row := db.SQLDB.QueryRow("SELECT filename, path, storage_tier FROM files WHERE id = ?", id)
	err := row.Scan(&filename, &path, &storageTier)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	// Fallbacks if missing
	if filename == "" {
		filename = id
	}
	if path == "" {
		// Backward compatibility: try legacy locations
		legacy := filepath.Join("./uploads_data", filename)
		if _, statErr := os.Stat(legacy); statErr == nil {
			path = legacy
		} else {
			legacy = filepath.Join("./uploads_data", id)
			path = legacy
		}
	}

	// Ensure file exists
	if _, statErr := os.Stat(path); statErr != nil {
		if os.IsNotExist(statErr) {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not available in storage"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot access file"})
		return
	}

	// Log access and update counters
	ua := c.Request.UserAgent()
	ip := c.ClientIP()
	_, _ = db.SQLDB.Exec("INSERT INTO file_access_log (file_id, user_agent, ip) VALUES (?, ?, ?)", id, ua, ip)
	_, _ = db.SQLDB.Exec("UPDATE files SET access_count = access_count + 1, last_accessed_at = ? , updated_at = ? WHERE id = ?", time.Now(), time.Now(), id)

	// Event-driven signals: Redis rolling 24h counter with TTL and publish to migration stream
	// Increment access count key and set TTL only if not already set (ExpireNX)
	key := "file:" + id + ":access_count"
	_ = db.RDB.Incr(db.Ctx, key).Err()
	_ = db.RDB.ExpireNX(db.Ctx, key, 24*time.Hour).Err()
	// Publish candidate to Redis Stream with trimming to keep it bounded
	_ = db.RDB.XAdd(db.Ctx, &redis.XAddArgs{
		Stream: "migration_candidates",
		MaxLen: 10000,
		Approx: true,
		ID:     "*",
		Values: map[string]interface{}{"file_id": id, "ts": time.Now().UnixMilli()},
	}).Err()

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.File(path)
}
