package api

import (
	"database/sql"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"mediapipeline/internal/config"
	"mediapipeline/internal/db"

	"github.com/redis/go-redis/v9"
)

// StartStorageMigrator launches an event-driven worker that promotes/demotes files based on recent access
func StartStorageMigrator(cfg *config.Config, policy StorageMigrationPolicy) {
	// Ensure directories exist
	if cfg.Storage.S3Path == "" {
		cfg.Storage.S3Path = "./storage/s3"
	}
	if cfg.Storage.R2Path == "" {
		cfg.Storage.R2Path = "./storage/r2"
	}
	_ = os.MkdirAll(cfg.Storage.S3Path, 0o755)
	_ = os.MkdirAll(cfg.Storage.R2Path, 0o755)

	go func() {
		ensureConsumerGroup("migration_candidates", "migrators")

		for {
			// Read new messages for this consumer
			streams, err := db.RDB.XReadGroup(db.Ctx, &redis.XReadGroupArgs{
				Group:    "migrators",
				Consumer: hostnameOr("worker-1"),
				Streams:  []string{"migration_candidates", ">"},
				Count:    64,
				Block:    5 * time.Second,
				NoAck:    false,
			}).Result()
			if err == redis.Nil {
				continue
			}
			if err != nil {
				log.Printf("migrator: XReadGroup error: %v", err)
				time.Sleep(2 * time.Second)
				continue
			}

			for _, s := range streams {
				for _, msg := range s.Messages {
					fileID, _ := msg.Values["file_id"].(string)
					if fileID == "" {
						// ack and skip malformed
						_, _ = db.RDB.XAck(db.Ctx, s.Stream, "migrators", msg.ID).Result()
						continue
					}
					processCandidate(cfg, policy, fileID)
					_, _ = db.RDB.XAck(db.Ctx, s.Stream, "migrators", msg.ID).Result()
				}
			}
		}
	}()
}

func processCandidate(cfg *config.Config, policy StorageMigrationPolicy, fileID string) {
	if db.SQLDB == nil {
		return
	}

	var (
		filename     string
		path         string
		storageTier  string
		lastAccessed sql.NullTime
	)
	row := db.SQLDB.QueryRow("SELECT filename, path, storage_tier, last_accessed_at FROM files WHERE id = ?", fileID)
	if err := row.Scan(&filename, &path, &storageTier, &lastAccessed); err != nil {
		if err != sql.ErrNoRows {
			log.Printf("migrator: db error for %s: %v", fileID, err)
		}
		return
	}

	now := time.Now()
	// Demotion check
	if policy.DemoteIdleHours > 0 && storageTier == "S3" {
		cutoff := now.Add(-time.Duration(policy.DemoteIdleHours) * time.Hour)
		idle := !lastAccessed.Valid || lastAccessed.Time.Before(cutoff)
		if idle {
			dst := filepath.Join(cfg.Storage.R2Path, filename)
			if moveFile(path, dst) {
				_, _ = db.SQLDB.Exec("UPDATE files SET storage_tier = 'R2', path = ?, updated_at = ? WHERE id = ?", dst, now, fileID)
				return
			}
			log.Printf("migrator: failed to demote %s", fileID)
		}
	}

	// Promotion check
	if policy.PromoteMinAccesses > 0 && storageTier == "R2" {
		key := "file:" + fileID + ":access_count"
		cntStr, err := db.RDB.Get(db.Ctx, key).Result()
		if err == redis.Nil {
			// no recent activity
			return
		}
		if err != nil {
			log.Printf("migrator: redis get error for %s: %v", fileID, err)
			return
		}
		cnt, _ := strconv.ParseInt(cntStr, 10, 64)
		if int(cnt) >= policy.PromoteMinAccesses {
			dst := filepath.Join(cfg.Storage.S3Path, filename)
			if moveFile(path, dst) {
				_, _ = db.SQLDB.Exec("UPDATE files SET storage_tier = 'S3', path = ?, updated_at = ? WHERE id = ?", dst, now, fileID)
				return
			}
			log.Printf("migrator: failed to promote %s", fileID)
		}
	}
}

func ensureConsumerGroup(stream, group string) {
	// Create group and stream if not exists
	if err := db.RDB.XGroupCreateMkStream(db.Ctx, stream, group, "$").Err(); err != nil {
		if err.Error() != "BUSYGROUP Consumer Group name already exists" {
			log.Printf("migrator: XGroupCreateMkStream error: %v", err)
		}
	}
}

func hostnameOr(fallback string) string {
	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}
	return fallback
}

func moveFile(src, dst string) bool {
	if src == dst {
		return true
	}
	if src == "" || dst == "" {
		return false
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return false
	}
	if err := os.Rename(src, dst); err == nil {
		return true
	}
	in, err := os.Open(src)
	if err != nil {
		return false
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return false
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return false
	}
	_ = os.Remove(src)
	return true
}
