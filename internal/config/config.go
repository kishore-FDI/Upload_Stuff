package config

import (
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Port        string
	Redis       RedisConfig
	Storage     StorageConfig
	AI          AIConfig
	JWTSecret   string
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
	S3Path string
	R2Path string
}

// getEnv reads environment variable or fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// Singleton variables
var (
	cfg  *Config
	once sync.Once
)

// Load returns the singleton config instance
func InitConfig() *Config {
	once.Do(func() {
		// Load .env file if it exists
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found or error reading it, using system env/fallbacks")
		}

		cfg = &Config{
			Environment: getEnv("ENVIRONMENT", "development"),
			Port:        getEnv("PORT", "8080"),
			Redis: RedisConfig{
				Host:     getEnv("REDIS_HOST", "localhost"),
				Port:     getEnv("REDIS_PORT", "6379"),
				Password: getEnv("REDIS_PASSWORD", ""),
				DB:       0, // Default DB
			},
			Storage: StorageConfig{
				S3Path: getEnv("S3_PATH", "./storage/s3"),
				R2Path: getEnv("R2_PATH", "./storage/r2"),
			},
			AI: AIConfig{
				BaseURL: getEnv("AI_SERVICE_URL", "http://localhost:8000"),
			},
			JWTSecret: getEnv("JWT_SECRET", "supersecretkey"),
		}
	})
	return cfg
}

func GetConfig() *Config {
	if cfg == nil {
		log.Println("Warning: Config not initialized, calling InitConfig automatically")
		InitConfig()
	}
	return cfg
}
