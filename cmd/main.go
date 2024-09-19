package main

import (
	"fmt"
	"log"
	"net/http"
	"pact/internal/db"
	"pact/internal/router"
)

func main() {
	err := db.OpenDatabase()
	if err != nil {
		log.Fatalf("db connections failed...: %v", err)
	}
	mux := router.Router()
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Printf("error with listen&serve: %v", err)
	}
}
