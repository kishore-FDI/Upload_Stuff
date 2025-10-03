package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"mediapipeline/internal/config"
	"mediapipeline/internal/db"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" || req.Email == "" {
		http.Error(w, "username, password, and email are required", http.StatusBadRequest)
		return
	}

	// Check if username/email already exists
	var exists int
	err := db.SQLDB.QueryRow("SELECT 1 FROM business WHERE name=? OR email=?", req.Username, req.Email).Scan(&exists)
	if err != sql.ErrNoRows && err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if exists == 1 {
		http.Error(w, "Username or email already exists", http.StatusConflict)
		return
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	_, err = db.SQLDB.Exec(
		"INSERT INTO business (name, email, password_hash) VALUES (?, ?, ?)",
		req.Username, req.Email, hashedPassword,
	)
	if err != nil {
		http.Error(w, "Database error inserting user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("User created successfully"))
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "username and password are required", http.StatusBadRequest)
		return
	}

	// Fetch business ID and password hash
	var passwordHash string
	var businessID int
	err := db.SQLDB.QueryRow("SELECT id, password_hash FROM business WHERE name=?", req.Username).
		Scan(&businessID, &passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !checkPasswordHash(req.Password, passwordHash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate access and refresh tokens
	accessToken, err := generateToken(businessID, req.Username, time.Minute*15)
	if err != nil {
		http.Error(w, "Error generating access token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := generateToken(businessID, req.Username, time.Hour*24*7)
	if err != nil {
		http.Error(w, "Error generating refresh token", http.StatusInternalServerError)
		return
	}

	// Store refresh token in DB
	_, err = db.SQLDB.Exec(
		"INSERT INTO refresh_tokens (username, token, expires_at) VALUES (?, ?, ?)",
		req.Username, refreshToken, time.Now().Add(time.Hour*24*7),
	)
	if err != nil {
		http.Error(w, "Database error saving refresh token", http.StatusInternalServerError)
		return
	}

	resp := tokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func RefreshToken(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	cfg := config.GetConfig()
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["username"] == nil || claims["business_id"] == nil {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	username := claims["username"].(string)
	businessID := int(claims["business_id"].(float64)) // JSON numbers parsed as float64

	// Check DB for refresh token validity
	var dbToken string
	err = db.SQLDB.QueryRow("SELECT token FROM refresh_tokens WHERE username=? AND token=?", username, req.RefreshToken).
		Scan(&dbToken)
	if err != nil {
		http.Error(w, "Refresh token not found", http.StatusUnauthorized)
		return
	}

	// Generate a new access token
	newAccessToken, err := generateToken(businessID, username, time.Minute*15)
	if err != nil {
		http.Error(w, "Error generating access token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"access_token": newAccessToken,
	})
}
