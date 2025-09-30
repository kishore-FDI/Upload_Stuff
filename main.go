package main

import (
	"fmt"
	"log"
	"mediapipeline/internal/config"
	"mediapipeline/internal/db"
)
func main(){
	cfg,err:=config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	// Initializing the databases
	db.InitRedis(cfg)
	db.InitSQLite()


	fmt.Print()
}