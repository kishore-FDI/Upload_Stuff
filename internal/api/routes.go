package api

import (
	"fmt"
	"log"
	"mediapipeline/internal/api/auth"
	apikeys "mediapipeline/internal/api/keys"
	"mediapipeline/internal/api/router"
	"mediapipeline/internal/middleware"
	"net/http"
	// "mediapipeline/internal/config"
)

func ServerHandler(w http.ResponseWriter, r *http.Request) error {
	fmt.Fprintf(w, "Hello, World!")
	return nil
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

func SetUpRoutes(routes *router.Router) {
	// Health check
	routes.HandleFunc("/health", ErrorHandler(ServerHandler))

	apiV1 := routes.Group("/api/v1")

	// Auth routes
	authRoutes := apiV1.Group("/auth")
	{
		authRoutes.HandleFunc("/signup", auth.SignUp)
		authRoutes.HandleFunc("/signin", auth.SignIn)
		authRoutes.HandleFunc("/refresh", auth.RefreshToken)
	}

	apiKeyRoutes := apiV1.Group("/apikeys")
	apiKeyRoutes.Use(middleware.AuthMiddleware)
	{
		apiKeyRoutes.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
			apikeys.CreateAPIKey(w, r)
		})
		apiKeyRoutes.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
			apikeys.ListAPIKeys(w, r)
		})
		apiKeyRoutes.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {

			apikeys.DeleteAPIKey(w, r)
		})
	}
}
