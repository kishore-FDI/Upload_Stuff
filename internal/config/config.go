package config

import (
	"os"
)

type Config struct {
	Environment string
	Port        string
	Redis       RedisConfig
	Storage     StorageConfig
	AI          AIConfig
}

type AIConfig struct {
	BaseURL string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type StorageConfig struct {
	S3Path  string
	R2Path  string
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func Load()(*Config,error){
	config := &Config{
		Environment: getEnv("ENVIRONMENT","development"),
		Port:        getEnv("PORT","8080"),
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0, // Default to DB 0
		},
		Storage: StorageConfig{
			S3Path:  getEnv("S3_PATH", "./storage/s3"),
			R2Path:  getEnv("R2_PATH", "./storage/r2"),
		},
		AI: AIConfig{
			BaseURL: getEnv("AI_SERVICE_URL", "http://localhost:8000"),
		},
	}
	return config,nil
}

