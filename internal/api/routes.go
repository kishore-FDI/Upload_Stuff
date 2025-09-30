package api

import (
	"log"
	"net/http"
	// "mediapipeline/internal/config"
)

func errorHandler(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("Panic recovered: %v", err)
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()

        // Execute the handler and check for errors
        if err := f(w, r); err != nil {
            log.Printf("Handler error: %v", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    }
}