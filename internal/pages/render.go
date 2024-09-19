package pages

import (
	"html/template"
	"net/http"
)

// RenderTemplate renders an HTML template and injects the provided data.
func RenderTemplate(w http.ResponseWriter, templatePath string, data TemplateData) {
	t, err := template.ParseFiles(templatePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
