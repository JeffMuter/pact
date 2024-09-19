package pages

import (
	"html/template"
	"log"
	"net/http"
)

var templates *template.Template

// InitTemplates() parses all templates to a single global template, but all the definitions, layouts, and blocks are all accessible. Allowing us to initialize them
// on app run, but not have to initialize a new set of templates for every request.
func InitTemplates() {
	var err error
	// Parse all templates in the templates directory with a `.html` extension
	templates, err = template.ParseGlob("./templates/*.html")
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}
}

// RenderTemplate renders an HTML template and injects the provided data.
func RenderTemplate(w http.ResponseWriter, templName string, data TemplateData) {

	w.Header().Set("Content-Type", "text/html; charset=utf-8") // necessary or else the template will load as plain text

	err := templates.ExecuteTemplate(w, templName, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
