package middleware

import (
	"context"
	"net/http"
	"strings"

	"mediapipeline/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// Define a private type for context keys
type contextKey string

const businessIDKey contextKey = "business_id"

type Claims struct {
	BusinessID int    `json:"business_id"`
	Username   string `json:"username"`
	jwt.RegisteredClaims
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.GetConfig().JWTSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Inject businessID into context using the typed key
		ctx := context.WithValue(r.Context(), businessIDKey, claims.BusinessID)
		next(w, r.WithContext(ctx))
	}
}

// Helper function to retrieve businessID from context safely
func GetBusinessID(ctx context.Context) (int, bool) {
	businessID, ok := ctx.Value(businessIDKey).(int)
	return businessID, ok
}
