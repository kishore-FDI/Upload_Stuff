package db

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type Business struct {
	ID        int
	Name      string
	Email     string
	APIKey    string
	CreatedAt string
}

// Generate a random API key
func GenerateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate api key: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// Insert a new business
func CreateBusiness(name, email string) (*Business, error) {
	apiKey, err := GenerateAPIKey()
	if err != nil {
		return nil, err
	}
	res, err := SQLDB.Exec("INSERT INTO business (name, email, api_key) VALUES (?, ?, ?)", name, email, apiKey)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &Business{ID: int(id), Name: name, Email: email, APIKey: apiKey}, nil
}

// GetBusinessByAPIKey fetches a business by its API key
func GetBusinessByAPIKey(apiKey string) (*Business, error) {
    row := SQLDB.QueryRow("SELECT id, name, email, api_key, created_at FROM business WHERE api_key = ?", apiKey)
    b := &Business{}
    if err := row.Scan(&b.ID, &b.Name, &b.Email, &b.APIKey, &b.CreatedAt); err != nil {
        return nil, err
    }
    return b, nil
}
