package config

import (
	"os"
)

// Config holds all configuration for the application
type Config struct {
	Environment string
	Port        string
	Redis       RedisConfig
	Storage     StorageConfig
	AI          AIConfig
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	CDNPath string
	S3Path  string
	R2Path  string
}

// AIConfig holds AI service configuration
type AIConfig struct {
	BaseURL string
	Timeout int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PORT", "8080"),
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0, // Default to DB 0
		},
		Storage: StorageConfig{
			CDNPath: getEnv("CDN_PATH", "./storage/cdn"),
			S3Path:  getEnv("S3_PATH", "./storage/s3"),
			R2Path:  getEnv("R2_PATH", "./storage/r2"),
		},
		AI: AIConfig{
			BaseURL: getEnv("AI_SERVICE_URL", "http://localhost:8000"),
			Timeout: 30, // 30 seconds timeout
		},
	}

	return cfg, nil
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

