package api

import (
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


