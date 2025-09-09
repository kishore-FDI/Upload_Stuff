package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"mediapipeline/internal/config"
	"mediapipeline/internal/db"
	"mediapipeline/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/tus/tusd/pkg/filestore"
	tusd "github.com/tus/tusd/pkg/handler"
	"github.com/tus/tusd/pkg/memorylocker"
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

		// Initialize TUS handler
		tusHandler, err := initTusHandler(cfg)
		if err != nil {
			log.Fatalf("failed to initialize tus handler: %v", err)
		}

		// TUS upload endpoints
		uploads := v1.Group("/uploads")
		{
			// TUS protocol endpoints
			uploads.POST("/", gin.WrapF(tusHandler.PostFile))
			uploads.HEAD("/:id", gin.WrapF(tusHandler.HeadFile))
			uploads.PATCH("/:id", gin.WrapF(tusHandler.PatchFile))
			uploads.GET("/:id", gin.WrapF(tusHandler.GetFile))
			uploads.DELETE("/:id", gin.WrapF(tusHandler.DelFile))
		}

		// Metadata/token management endpoints
		uploadsMeta := v1.Group("/uploads/meta")
		uploadsMeta.POST("/", middleware.RateLimiter(db.RDB, 10, time.Minute, middleware.UserRateLimit{}), uploadHandler)
		uploadsMeta.PUT("/:token", resumeUploadHandler)
		uploadsMeta.GET("/:token/status", statusHandler)

		// Business endpoints for managing uploads
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

func uploadHandler(c *gin.Context) {
	// Get business API key and username from headers
	apiKey := c.GetHeader("X-API-KEY")
	username := c.GetHeader("X-Username")
	
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing X-API-KEY header"})
		return
	}
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing X-Username header"})
		return
	}

	// Validate business
	business, err := db.GetBusinessByAPIKey(apiKey)
	if err != nil || business == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
		return
	}

	// Generate upload token
	token, err := db.GenerateAPIKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Store upload data directly with token as key
	uploadKey := "upload:" + token
	fields := map[string]interface{}{
		"business_id": business.ID,
		"username":    username,
		"status":      "in_progress",
		"created_at":  time.Now().UTC().Format(time.RFC3339),
	}
	if err := db.RDB.HSet(db.Ctx, uploadKey, fields).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to persist upload"})
		return
	}

	// Store token metadata
	tokenKey := "upload_token:" + token
	tokenFields := map[string]interface{}{
		"business_id": business.ID,
		"username":    username,
		"status":      "used",
		"created_at":  time.Now().UTC().Format(time.RFC3339),
		"used_at":     time.Now().UTC().Format(time.RFC3339),
	}
	if err := db.RDB.HSet(db.Ctx, tokenKey, tokenFields).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store token"})
		return
	}

	// Set token expiration (15 minutes)
	if err := db.RDB.Expire(db.Ctx, tokenKey, 15*time.Minute).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set token expiration"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token":      token,
		"expires_in": 900,
	})
}

func resumeUploadHandler(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing token parameter"})
		return
	}

	// Check if upload exists
	uploadKey := "upload:" + token
	uploadData, err := db.RDB.HGetAll(db.Ctx, uploadKey).Result()
	if err != nil || len(uploadData) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
		return
	}

	// Check if upload is in progress
	if status, ok := uploadData["status"]; !ok || status != "in_progress" {
		c.JSON(http.StatusConflict, gin.H{"error": "upload not in progress"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"status":     uploadData["status"],
		"created_at": uploadData["created_at"],
	})
}

func statusHandler(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Status handler not implemented yet"})
}

func downloadHandler(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Download handler not implemented yet"})
}

func deleteHandler(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Delete handler not implemented yet"})
}

func moderationHandler(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Moderation handler not implemented yet"})
}

func resultHandler(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Result handler not implemented yet"})
}

func listBusinessUploadsHandler(c *gin.Context) {
	// Get business API key from header
	apiKey := c.GetHeader("X-API-KEY")
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing X-API-KEY header"})
		return
	}

	// Validate business
	business, err := db.GetBusinessByAPIKey(apiKey)
	if err != nil || business == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
		return
	}

	// Get optional username filter
	username := c.Query("username")

	// Scan for uploads belonging to this business
	pattern := "upload:*"
	keys, err := db.RDB.Keys(db.Ctx, pattern).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan uploads"})
		return
	}

	var uploads []gin.H
	for _, key := range keys {
		uploadData, err := db.RDB.HGetAll(db.Ctx, key).Result()
		if err != nil || len(uploadData) == 0 {
			continue
		}

		// Filter by business
		if uploadData["business_id"] != fmt.Sprintf("%d", business.ID) {
			continue
		}

		// Filter by username if provided
		if username != "" && uploadData["username"] != username {
			continue
		}

		token := key[7:] // Remove "upload:" prefix
		uploads = append(uploads, gin.H{
			"token":      token,
			"username":   uploadData["username"],
			"status":     uploadData["status"],
			"created_at": uploadData["created_at"],
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"business_id": business.ID,
		"uploads":     uploads,
		"count":       len(uploads),
	})
}

// initTusHandler initializes tusd with a disk filestore and hooks for validation
func initTusHandler(_ *config.Config) (*tusd.UnroutedHandler, error) {
	// ensure upload directory exists
	uploadDir := "./uploads_data"
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create upload dir: %w", err)
	}

	// filestore: writes to local disk. Replace with S3 store for production.
	store := filestore.New(uploadDir)
	locker := memorylocker.New()

	composer := tusd.NewStoreComposer()
	store.UseIn(composer)
	locker.UseIn(composer)

	config := tusd.Config{
		StoreComposer:           composer,
		BasePath:                "/api/v1/uploads/",
		DisableDownload:         false,
		DisableTermination:      false,
		NotifyCreatedUploads:    true,
		NotifyCompleteUploads:   true,
		NotifyUploadProgress:    true, // enable progress notifications
		RespectForwardedHeaders: true,
	}

	// Validate API key and username before creating uploads.
	config.PreUploadCreateCallback = func(hook tusd.HookEvent) error {
		h := hook.HTTPRequest
		apiKey := h.Header.Get("X-API-KEY")
		username := h.Header.Get("X-Username")
		if apiKey == "" || username == "" {
			return tusd.NewHTTPError(fmt.Errorf("missing auth headers"), http.StatusBadRequest)
		}

		business, err := db.GetBusinessByAPIKey(apiKey)
		if err != nil || business == nil {
			return tusd.NewHTTPError(fmt.Errorf("invalid api key"), http.StatusUnauthorized)
		}

		if hook.Upload.MetaData == nil {
			hook.Upload.MetaData = make(map[string]string)
		}
		hook.Upload.MetaData["business_id"] = fmt.Sprintf("%d", business.ID)
		hook.Upload.MetaData["username"] = username

		return nil
	}

	config.PreFinishResponseCallback = func(hook tusd.HookEvent) error {
		id := hook.Upload.ID
		meta := hook.Upload.MetaData
		uploadKey := "upload:" + id
		fields := map[string]interface{}{
			"business_id": meta["business_id"],
			"username":    meta["username"],
			"status":      "uploaded",
			"size":        hook.Upload.Size,
			"created_at":  time.Now().UTC().Format(time.RFC3339),
		}
		_ = db.RDB.HSet(db.Ctx, uploadKey, fields)
		_ = db.RDB.Expire(db.Ctx, uploadKey, 24*time.Hour)
		return nil
	}

	h, err := tusd.NewUnroutedHandler(config)
	if err != nil {
		return nil, err
	}

	// Start goroutine for logging progress with a CLI progress bar
	go func() {
		for {
			select {
			case info := <-h.CreatedUploads:
				log.Printf("Upload %s created (size: %d)", info.Upload.ID, info.Upload.Size)
			case info := <-h.UploadProgress:
				if info.Upload.Size > 0 {
					percent := float64(info.Upload.Offset) / float64(info.Upload.Size) * 100
					barWidth := 50
					filled := int(percent / 100 * float64(barWidth))
					bar := fmt.Sprintf("\r[%s%s] %.2f%%",
						strings.Repeat("=", filled),
						strings.Repeat(" ", barWidth-filled),
						percent,
					)
					fmt.Print(bar)
				}
			case info := <-h.CompleteUploads:
				fmt.Printf("\r[==================================================] 100.00%%\n")
				log.Printf("Upload %s completed (%d bytes)", info.Upload.ID, info.Upload.Size)
			}
		}
	}()

	return h, nil
}
