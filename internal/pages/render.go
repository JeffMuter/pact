package pages

import (
	"log"
	"net/http"
	"path/filepath"
	"text/template"
)

// RenderTemplate renders an HTML template and injects the provided data.
func RenderTemplate(w http.ResponseWriter, layoutTempl string, contentTempl string, data TemplateData) error {
	// Parse the layout and block templates
	files := []string{
		filepath.Join("templates", layoutTempl),  // e.g., "base.html"
		filepath.Join("templates", contentTempl), // e.g., "home.html"
	}

	templ, err := template.ParseFiles(files...)
	if err != nil {
		return err
	}

	//	w.Header().Set("Content-Type", "text/html; charset=utf-8") // necessary or else the template will load as plain text

	// Always render the layout.html as the base template
	err = templ.ExecuteTemplate(w, "defaultLayout", data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Error rendering layout: %v", err)
		return err
	}
	return nil
}
