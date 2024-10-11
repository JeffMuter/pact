package main

import (
	"log"
	"net/http"
	"pact/internal/db"
	"pact/internal/pages"
	"pact/internal/router"
)

func main() {
	err := db.OpenDatabase()
	if err != nil {
		log.Fatalf("db connections failed...: %v", err)
	}

	// Initialize templates
	err = pages.InitTemplates()
	if err != nil {
		log.Fatalf("Failed to initialize templates: %v", err)
	}

	// Setup route
	r := router.Router()

	// Start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
