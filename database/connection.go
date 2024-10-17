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
	db, err = sql.Open("sqlite3", "./database.db")
	if err != nil {
		return fmt.Errorf("error unable to connect to db: %w", err)
	}
	fmt.Println("Database connection opened")

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
