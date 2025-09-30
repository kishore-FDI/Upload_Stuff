package main

import (
	"fmt"
	"log"
	"mediapipeline/internal/api"
	"mediapipeline/internal/config"
	"mediapipeline/internal/db"

	"net/http"
)
func main(){
	cfg,err:=config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initializing the databases
	db.InitRedis(cfg)
	db.InitSQLite()

	// setting up the routes
	mux := http.NewServeMux()
	api.SetUpRoutes(mux)
	
	fmt.Println("Server starting on :",cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))

}