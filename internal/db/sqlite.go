package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var SQLDB *sql.DB

func InitSQLite() {
	var err error
	SQLDB, err = sql.Open("sqlite3", "./mediapipeline.db")
	if err != nil {
		log.Fatalf("Failed to open SQLite DB: %v", err)
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS business (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		api_key TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := SQLDB.Exec(createTable); err != nil {
		log.Fatalf("Failed to create business table: %v", err)
	}

	log.Println("SQLite initialized and business table ready")
}

