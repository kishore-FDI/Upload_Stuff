package api

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// NotifyExternalUploadCompleted sends the file ID to an external HTTP endpoint (non-blocking)
func NotifyExternalUploadCompleted(id string) {
	url := os.Getenv("EXTERNAL_HTTP_URL")
	if url == "" {
		// Not configured; skip
		return
	}
	payload := map[string]string{"file_id": id}
	body, _ := json.Marshal(payload)

	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("external notify prepare failed: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("external notify failed: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("external notify non-2xx: %s", resp.Status)
	}
}
