// TO-DO Review the tables later
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
			name TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := SQLDB.Exec(createTable); err != nil {
		log.Fatalf("Failed to create business table: %v", err)
	}

	createRefreshTokens := `
		CREATE TABLE IF NOT EXISTS refresh_tokens (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			business_id INTEGER NOT NULL,
			token TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (business_id) REFERENCES business(id)
		);
	`
	if _, err := SQLDB.Exec(createRefreshTokens); err != nil {
		log.Fatalf("Failed to create refresh_tokens table: %v", err)
	}
	createAPIKeys := `
		CREATE TABLE IF NOT EXISTS api_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			business_id INTEGER NOT NULL,
			key TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (business_id) REFERENCES business(id)
		);
	`
	if _, err := SQLDB.Exec(createAPIKeys); err != nil {
		log.Fatalf("Failed to create api_keys table: %v", err)
	}
	createFiles := `
	CREATE TABLE IF NOT EXISTS files (
		id TEXT PRIMARY KEY,
		business_id INTEGER,
		filename TEXT,
		size_bytes INTEGER,
		storage_tier TEXT NOT NULL DEFAULT 'S3',
		path TEXT NOT NULL,
		access_count INTEGER NOT NULL DEFAULT 0,
		last_accessed_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME,
		FOREIGN KEY (business_id) REFERENCES business(id)
	);
	`
	if _, err := SQLDB.Exec(createFiles); err != nil {
		log.Fatalf("Failed to create files table: %v", err)
	}

	createAccess := `
	CREATE TABLE IF NOT EXISTS file_access_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_id TEXT NOT NULL,
		accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		user_agent TEXT,
		ip TEXT,
		FOREIGN KEY (file_id) REFERENCES files(id)
	);
	`
	if _, err := SQLDB.Exec(createAccess); err != nil {
		log.Fatalf("Failed to create file_access_log table: %v", err)
	}

	createAPIKeyFiles := `
	CREATE TABLE IF NOT EXISTS business_files (
		api_key TEXT NOT NULL,
		file_id TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (api_key, file_id),
		FOREIGN KEY (api_key) REFERENCES api_keys(key),
		FOREIGN KEY (file_id) REFERENCES files(id)
	);
	`
	if _, err := SQLDB.Exec(createAPIKeyFiles); err != nil {
		log.Fatalf("Failed to create business_files table: %v", err)
	}

	log.Println("SQLite initialized and tables ready")
}
