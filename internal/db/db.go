package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func OpenDatabase() error {
	var err error
	db, err = sql.Open("sqlite3", "./database/database.db")
	if err != nil {
		return fmt.Errorf("error unable to connect to db: %w", err)
	}

	err = db.Ping()
	if err != nil {
		fmt.Println("unable to ping db...", err)
		return fmt.Errorf("error pinging db: %w", err)
	}
	return nil
}

func GetDB() *sql.DB {
	return db
}
