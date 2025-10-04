package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

// Custom claims
type Claims struct {
	BusinessID int    `json:"business_id"`
	Username   string `json:"username"`
	jwt.RegisteredClaims
}
