package main

import (
	"log"
	"os"

	"mediapipeline/internal/api"
	"mediapipeline/internal/config"
	"mediapipeline/internal/db"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Init Redis and SQLite
	db.InitRedis()
	db.InitSQLite()
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Start background storage migrator (SQLite-based policy)
	api.StartStorageMigrator(cfg, api.StorageMigrationPolicy{
		PromoteMinAccesses: 5,  // >= 5 accesses in last 24h -> S3
		DemoteIdleHours:    24, // idle for 24h -> R2
	})

	r := gin.Default()

	api.SetupRoutes(r, cfg)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Media Pipeline API server on port okay %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
