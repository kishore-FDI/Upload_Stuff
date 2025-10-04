package apikeys

import (
	"crypto/rand"
	"encoding/hex"
)
func generateRandomAPIKey() (string, error) {
	bytes := make([]byte, 32) // 256-bit key
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
