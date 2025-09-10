package api

import (
	"fmt"
	"mediapipeline/internal/db"
	"mediapipeline/internal/middleware"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RegisterBusinessRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

func SetupBusinessRoutes(r *gin.RouterGroup) {
	business := r.Group("/business")
	business.Use(middleware.RateLimiter(db.RDB, 1, time.Minute, middleware.BusinessRateLimit{}))
	{
		business.POST("/register", registerBusinessHandler)
	}
}

func registerBusinessHandler(c *gin.Context) {
	var req RegisterBusinessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	business, err := db.CreateBusiness(req.Name, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create business: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Business registered successfully",
		"api_key": business.APIKey,
	})
}

func listBusinessUploadsHandler(c *gin.Context) {
	apiKey := c.GetHeader("X-API-KEY")
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing X-API-KEY header"})
		return
	}

	business, err := db.GetBusinessByAPIKey(apiKey)
	if err != nil || business == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
		return
	}

	username := c.Query("username")
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
		if uploadData["business_id"] != fmt.Sprintf("%d", business.ID) {
			continue
		}
		if username != "" && uploadData["username"] != username {
			continue
		}
		token := key[7:]
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
