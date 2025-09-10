package api

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"mediapipeline/internal/config"
	"mediapipeline/internal/db"
	"mediapipeline/internal/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, cfg *config.Config) {
	r.Use(corsMiddleware())
	r.GET("/health", healthCheck)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Media Pipeline API",
				"version": "1.0.0",
				"service": "Content Moderation Pipeline",
			})
		})

		tusHandler, err := initTusHandler(cfg)
		if err != nil {
			log.Fatalf("failed to initialize tus handler: %v", err)
		}

		uploads := v1.Group("/uploads")
		{
			uploads.POST("/", gin.WrapF(tusHandler.PostFile))
			uploads.HEAD("/:id", gin.WrapF(tusHandler.HeadFile))
			uploads.PATCH("/:id", gin.WrapF(tusHandler.PatchFile))
			uploads.GET("/:id", gin.WrapF(tusHandler.GetFile))
			uploads.DELETE("/:id", gin.WrapF(tusHandler.DelFile))
		}

		uploadsMeta := v1.Group("/uploads/meta")
		uploadsMeta.POST("/", middleware.RateLimiter(db.RDB, 10, time.Minute, middleware.UserRateLimit{}), uploadHandler)
		uploadsMeta.PUT("/:token", resumeUploadHandler)
		uploadsMeta.GET("/:token/status", statusHandler)

		business := v1.Group("/business")
		business.Use(middleware.RateLimiter(db.RDB, 10, time.Minute, middleware.BusinessRateLimit{}))
		{
			business.GET("/uploads", listBusinessUploadsHandler)
		}

		storage := v1.Group("/storage")
		storage.Use(middleware.RateLimiter(db.RDB, 10, time.Minute, middleware.UserRateLimit{}))
		{
			storage.GET("/:id", downloadHandler)
			storage.DELETE("/:id", deleteHandler)
		}

		ws := v1.Group("/ws")
		{
			ws.GET("/:id", wsHandler)
		}

		moderation := v1.Group("/moderation")
		moderation.Use(middleware.RateLimiter(db.RDB, 10, time.Minute, middleware.UserRateLimit{}))
		{
			moderation.POST("/check", moderationHandler)
			moderation.GET("/:id/result", resultHandler)
		}

		SetupBusinessRoutes(v1)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-KEY, X-Upload-Token, Tus-Resumable, Upload-Length, Upload-Metadata, Upload-Offset, Upload-Concat")
		c.Header("Access-Control-Expose-Headers", "Location, Upload-Offset")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "Media Pipeline API",
	})
}

func statusHandler(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing token parameter"})
		return
	}

	// Check if it's a TUS upload ID (starts with upload:)
	uploadKey := "upload:" + token
	uploadData, err := db.RDB.HGetAll(db.Ctx, uploadKey).Result()
	if err != nil || len(uploadData) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
		return
	}

	// Parse progress data
	status := uploadData["status"]
	progress := 0.0
	if p, ok := uploadData["progress"]; ok && p != "" {
		if parsed, err := strconv.ParseFloat(p, 64); err == nil {
			progress = parsed
		}
	}

	offset := int64(0)
	if o, ok := uploadData["offset"]; ok && o != "" {
		if parsed, err := strconv.ParseInt(o, 10, 64); err == nil {
			offset = parsed
		}
	}

	size := int64(0)
	if s, ok := uploadData["size"]; ok && s != "" {
		if parsed, err := strconv.ParseInt(s, 10, 64); err == nil {
			size = parsed
		}
	}

	response := gin.H{
		"token":    token,
		"status":   status,
		"progress": progress,
		"offset":   offset,
		"size":     size,
	}

	// Add timestamps if available
	if createdAt, ok := uploadData["created_at"]; ok {
		response["created_at"] = createdAt
	}
	if updatedAt, ok := uploadData["updated_at"]; ok {
		response["updated_at"] = updatedAt
	}
	if completedAt, ok := uploadData["completed_at"]; ok {
		response["completed_at"] = completedAt
	}

	c.JSON(http.StatusOK, response)
}

func deleteHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file ID parameter"})
		return
	}

	// Check if file exists
	filePath := "./uploads_data/" + id
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot access file"})
		}
		return
	}

	// Try to get filename from TUS info
	info, err := readTusInfo(id)
	filename := id
	if err == nil {
		if fn, ok := info.MetaData["filename"]; ok && fn != "" {
			filename = fn
			realPath := filepath.Join("./uploads_data", filename)
			if _, err := os.Stat(realPath); err == nil {
				filePath = realPath
			}
		}
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete file"})
		return
	}

	// Delete the .info file if it exists
	infoPath := filepath.Join("./uploads_data", id+".info")
	if _, err := os.Stat(infoPath); err == nil {
		os.Remove(infoPath)
	}

	// Remove from Redis
	uploadKey := "upload:" + id
	db.RDB.Del(db.Ctx, uploadKey)

	c.JSON(http.StatusOK, gin.H{
		"message":  "file deleted successfully",
		"file_id":  id,
		"filename": filename,
	})
}

func moderationHandler(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Moderation handler not implemented yet"})
}

func resultHandler(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Result handler not implemented yet"})
}
