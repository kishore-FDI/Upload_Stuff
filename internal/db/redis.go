package db

import (
	"context"
	"log"
	"mediapipeline/internal/config"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client
var Ctx = context.Background()

func InitRedis(cfg *config.Config) {
	addr := cfg.Redis.Host+":"+cfg.Redis.Port
	pass := cfg.Redis.Password
	Rdb = redis.NewClient(&redis.Options{
		Addr:addr,
		Password:pass,
		DB:0,
	})
	
	if err := Rdb.Ping(Ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis")
}