package api

import (
	"mediapipeline/internal/db"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func uploadHandler(c *gin.Context) {
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

	business, err := db.GetBusinessByAPIKey(apiKey)
	if err != nil || business == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
		return
	}

	token, err := db.GenerateAPIKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

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

	uploadKey := "upload:" + token
	uploadData, err := db.RDB.HGetAll(db.Ctx, uploadKey).Result()
	if err != nil || len(uploadData) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "upload not found"})
		return
	}

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
