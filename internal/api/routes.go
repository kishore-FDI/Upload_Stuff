package api

import (
	"fmt"
	"log"
	"net/http"
	// "mediapipeline/internal/config"
)

func ServerHandler(w http.ResponseWriter,r *http.Request)error{
	fmt.Fprintf(w, "Hello, World!")
	return nil
}

func SetUpRoutes(routes *http.ServeMux){
	routes.Handle("/health",ErrorHandler(ServerHandler))
}

func ErrorHandler(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("Panic recovered: %v", err)
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()

        if err := f(w, r); err != nil {
            log.Printf("Handler error: %v", err)
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
    }
}