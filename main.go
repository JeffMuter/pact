package main

import (
	"fmt"
	"log"
	"net/http"
	"pact/database"
	"pact/internal/buckets"
	"pact/internal/pages"
	"pact/internal/router"
	"pact/internal/storage"
	"pact/internal/stripe"
	"time"
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

	err = storage.Init()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Wire dependencies
	buckets.SetRenderFunc(pages.RenderTemplateFraction)
	stripe.SetDB(database.GetQueries())

	go func() {
		for {
			buckets.ProcessDueRepeatingTasks()
			time.Sleep(15 * time.Minute)
		}
	}()

	// Setup router
	r := router.Router()

	// Start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
