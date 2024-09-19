package pages

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

var templates *template.Template

// InitTemplates() parses all templates to a single global template, but all the definitions, layouts, and blocks are all accessible. Allowing us to initialize them
// on app run, but not have to initialize a new set of templates for every request.
func InitTemplates() {
	var err error

	// Print the current working directory for debugging
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}
	log.Printf("Current working directory: %s", dir)

	templates, err = template.ParseGlob("./internal/templates/*.html")
	if err != nil {
		fmt.Printf("Error parsing templates: %v", err)
	}
}

// RenderTemplate renders an HTML template and injects the provided data.
func RenderTemplate(w http.ResponseWriter, templName string, data TemplateData) {
	fmt.Println("made it in render.templname: " + templName)
	w.Header().Set("Content-Type", "text/html; charset=utf-8") // necessary or else the template will load as plain text

	fmt.Println(templates.Templates())

	err := templates.ExecuteTemplate(w, templName, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
