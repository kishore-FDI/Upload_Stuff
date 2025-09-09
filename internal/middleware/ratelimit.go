package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

type RateLimitStrategy interface {
	Key(c *gin.Context) (string, error)
}

type BusinessRateLimit struct{}

func (s BusinessRateLimit) Key(c *gin.Context) (string, error) {
	ip := c.ClientIP()
	return "rate:" + ip + ":" + c.Request.Method + ":" + c.FullPath(), nil
}

type UserRateLimit struct{}

func (s UserRateLimit) Key(c *gin.Context) (string, error) {
	username := c.GetHeader("X-Username")
	apiKey := c.GetHeader("X-API-KEY")
	if apiKey == "" {
		return "", errors.New("missing X-API-KEY header")
	}
	return "api_key:" + apiKey + ":rate:" + username + ":" + c.Request.Method + ":" + c.FullPath(), nil
}

type TokenRateLimit struct{}

func (s TokenRateLimit) Key(c *gin.Context) (string, error) {
	token := c.GetHeader("X-Upload-Token")
	if token == "" {
		return "", errors.New("missing X-Upload-Token header")
	}
	return "token:" + token + ":rate:" + c.Request.Method + ":" + c.FullPath(), nil
}
func RateLimiter(rdb *redis.Client, limit int, window time.Duration, strat RateLimitStrategy) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		key, err := strat.Key(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "rate limiter error," + err.Error()})
			return
		}
		if count == 1 {
			rdb.Expire(ctx, key, window)
		}
		if count > int64(limit) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}

// TokenValidator validates upload tokens and ensures single-use
func TokenValidator(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Upload-Token")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing X-Upload-Token header"})
			return
		}

		// Check if token exists and is valid
		tokenKey := "upload_token:" + token
		tokenData, err := rdb.HGetAll(c.Request.Context(), tokenKey).Result()
		if err != nil || len(tokenData) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid upload token"})
			return
		}

		// Check token status - only allow "issued" tokens for new uploads
		if status, ok := tokenData["status"]; !ok || status != "issued" {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "token already used or expired"})
			return
		}

		// Store token data in context for use in handlers
		c.Set("token_data", tokenData)
		c.Set("business_id", tokenData["business_id"])
		c.Next()
	}
}

