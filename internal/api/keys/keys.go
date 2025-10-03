package apikeys

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"mediapipeline/internal/db"
	"mediapipeline/internal/middleware"
)

// --- Handlers ---

// CreateAPIKey generates a new API key for the authenticated business
func CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	businessID, ok := middleware.GetBusinessID(r.Context())

	fmt.Println(businessID)

	if !ok {
		http.Error(w, "Failed to find the ID of the Registered Business", http.StatusInternalServerError)
	}

	apiKey, err := generateRandomAPIKey()
	if err != nil {
		http.Error(w, "Failed to generate API key", http.StatusInternalServerError)
		return
	}

	_, err = db.SQLDB.Exec(
		"INSERT INTO api_keys (business_id, key, created_at) VALUES (?, ?, ?)",
		businessID, apiKey, time.Now(),
	)
	if err != nil {
		http.Error(w, "Database error saving API key", http.StatusInternalServerError)
		return
	}

	resp := map[string]string{
		"api_key": apiKey,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListAPIKeys returns all API keys for the authenticated business
func ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	businessID, ok := middleware.GetBusinessID(r.Context())

	if !ok {
		http.Error(w, "Failed to find the ID of the Registered Business", http.StatusInternalServerError)
	}

	rows, err := db.SQLDB.Query(
		"SELECT id, key, created_at FROM api_keys WHERE business_id=?",
		businessID,
	)
	if err != nil {
		http.Error(w, "Database error fetching API keys", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var keys []map[string]interface{}
	for rows.Next() {
		var id int
		var key string
		var createdAt string
		if err := rows.Scan(&id, &key, &createdAt); err != nil {
			http.Error(w, "Error scanning API keys", http.StatusInternalServerError)
			return
		}
		keys = append(keys, map[string]interface{}{
			"id":         id,
			"api_key":    key,
			"created_at": createdAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

// DeleteAPIKey deletes a specific API key for the authenticated business
func DeleteAPIKey(w http.ResponseWriter, r *http.Request, keyID int) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	businessID, ok := middleware.GetBusinessID(r.Context())

	if !ok {
		http.Error(w, "Failed to find the ID of the Registered Business", http.StatusInternalServerError)
	}

	res, err := db.SQLDB.Exec(
		"DELETE FROM api_keys WHERE id=? AND business_id=?",
		keyID, businessID,
	)
	if err != nil {
		http.Error(w, "Database error deleting API key", http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "API key not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
