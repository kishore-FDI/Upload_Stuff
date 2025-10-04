package main

import (
	"fmt"
	"log"
	"mediapipeline/internal/api"
	"mediapipeline/internal/api/router"
	"mediapipeline/internal/config"
	"mediapipeline/internal/db"
	"mediapipeline/internal/middleware"

	"net/http"
)

func main() {
	// Loading Config
	cfg := config.InitConfig()
	if cfg.JWTSecret == "" {
		log.Println("JWTSecret is empty!")
		panic("")
	} else {
		log.Println(config.GetConfig().JWTSecret)
	}

	// Initializing the databases
	db.InitRedis()
	db.InitSQLite()

	// setting up the routes
	mux := router.NewRouter()
	api.SetUpRoutes(mux)

	// middleware for logging
	server := middleware.Logging(mux)

	// Starting the server
	fmt.Println("Server starting on :", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, server))

}
