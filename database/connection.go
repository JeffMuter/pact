package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var queries *Queries

func OpenDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", "./database/database.db?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("error unable to connect to db: %w", err)
	}
	fmt.Println("Database connection opened")

	// Configure connection pool for SQLite (single writer optimal)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("error pinging db: %w", err)
	}
	fmt.Println("Database pinged successfully")

	queries = New(db)
	if queries == nil {
		return fmt.Errorf("failed to initialize queries")
	} else {
		fmt.Println("Queries initialized")
	}

	return nil
}

func GetDB() *sql.DB {
	return db
}

func GetQueries() *Queries {
	if queries == nil {
		fmt.Println("if you see this, there's a problem in getqueries...")
	}
	return queries
}

// SetTestDB sets the global database and queries for testing purposes only.
// This should only be called from test code.
func SetTestDB(testDB *sql.DB) {
	db = testDB
	queries = New(testDB)
}
