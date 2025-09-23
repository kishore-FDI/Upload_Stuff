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

// LinkFileToAPIKey inserts a mapping between a business api_key and a file_id
func LinkFileToAPIKey(apiKey string, fileID string) error {
	_, err := SQLDB.Exec("INSERT OR IGNORE INTO business_files (api_key, file_id) VALUES (?, ?)", apiKey, fileID)
	return err
}

// ListFileIDsByAPIKey returns file ids associated with an api_key
func ListFileIDsByAPIKey(apiKey string) ([]string, error) {
	rows, err := SQLDB.Query("SELECT file_id FROM business_files WHERE api_key = ? ORDER BY created_at DESC", apiKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}
