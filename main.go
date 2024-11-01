package main

import (
	"fmt"
	"log"
	"net/http"
	"pact/database"
	"pact/internal/pages"
	"pact/internal/router"
)

func main() {
	// open database
	err := database.OpenDatabase()
	if err != nil {
		log.Fatalf("db connections failed...: %v", err)
	}

	fmt.Println("db opened...")

	// Initialize templates in local memory
	err = pages.InitTemplates()
	if err != nil {
		log.Fatalf("Failed to initialize templates: %v", err)
	}

	fmt.Println("templates initialized...")

	// Setup route
	r := router.Router()

	// Start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
