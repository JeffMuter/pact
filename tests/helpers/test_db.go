package helpers

import (
	"database/sql"
	"os"
	"pact/database"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
)

// SetupTestDB creates an in-memory SQLite database with the schema loaded.
// Returns the database connection and a cleanup function that should be deferred.
func SetupTestDB(t *testing.T) (*sql.DB, *database.Queries, func()) {
	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	require.NoError(t, err, "failed to open in-memory database")

	// Verify connection
	err = db.Ping()
	require.NoError(t, err, "failed to ping in-memory database")

	// Load and execute schema
	schema, err := os.ReadFile("../../database/schema.sql")
	require.NoError(t, err, "failed to read schema.sql")

	_, err = db.Exec(string(schema))
	require.NoError(t, err, "failed to execute schema")

	// Create queries instance
	queries := database.New(db)
	require.NotNil(t, queries, "failed to create queries instance")

	// Return cleanup function
	cleanup := func() {
		db.Close()
	}

	return db, queries, cleanup
}

// SetupTestDBWithPath is like SetupTestDB but allows specifying the schema path
// for tests at different directory depths.
func SetupTestDBWithPath(t *testing.T, schemaPath string) (*sql.DB, *database.Queries, func()) {
	// Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	require.NoError(t, err, "failed to open in-memory database")

	// Verify connection
	err = db.Ping()
	require.NoError(t, err, "failed to ping in-memory database")

	// Load and execute schema
	schema, err := os.ReadFile(schemaPath)
	require.NoError(t, err, "failed to read schema.sql from "+schemaPath)

	_, err = db.Exec(string(schema))
	require.NoError(t, err, "failed to execute schema")

	// Create queries instance
	queries := database.New(db)
	require.NotNil(t, queries, "failed to create queries instance")

	// Return cleanup function
	cleanup := func() {
		db.Close()
	}

	return db, queries, cleanup
}

// SetupTestDBWithGlobalQueries creates a test database and sets it as the global database
// for tests that use database.GetQueries() (like middleware tests).
func SetupTestDBWithGlobalQueries(t *testing.T, schemaPath string) (*sql.DB, *database.Queries, func()) {
	db, queries, dbCleanup := SetupTestDBWithPath(t, schemaPath)

	// Set as global database for middleware/handlers that use database.GetQueries()
	database.SetTestDB(db)

	cleanup := func() {
		dbCleanup()
	}

	return db, queries, cleanup
}
