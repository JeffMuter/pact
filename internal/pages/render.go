package pages

import (
	"fmt"
	"log"
	"net/http"
	"pact/internal/auth"
	"path/filepath"
	"strings"
	"text/template"
)

type TemplateConstruct struct {
	layouts   map[string]*template.Template
	fractions map[string]*template.Template
}

var tmplConstruct *TemplateConstruct

func InitTemplates() error {
	tmplConstruct = &TemplateConstruct{
		layouts:   make(map[string]*template.Template),
		fractions: make(map[string]*template.Template),
	}

	// Parse the default layout first
	defaultLayout, err := template.ParseFiles("./internal/templates/contentTemplates/defaultLayout.html")
	if err != nil {
		return fmt.Errorf("error parsing default layout: %w", err)
	}

	// Parse all fraction templates
	fractions, err := filepath.Glob("./internal/templates/fractions/*.html")
	if err != nil {
		return fmt.Errorf("error globbing fractions: %w", err)
	}

	// Parse fractions and add them to both layouts and fractions maps
	for _, fraction := range fractions {
		name := strings.TrimSuffix(filepath.Base(fraction), ".html")

		// Parse the fraction template
		tmpl, err := template.Must(defaultLayout.Clone()).ParseFiles(fraction)
		if err != nil {
			return fmt.Errorf("error parsing fraction %s: %w", name, err)
		}

		// Add to fractions map
		tmplConstruct.fractions[name] = tmpl

		// Add to layouts map (each fraction can also be rendered as a full page)
		tmplConstruct.layouts[name] = tmpl
	}

	// Parse additional layout templates
	layouts, err := filepath.Glob("./internal/templates/contentTemplates/*.html")
	if err != nil {
		return fmt.Errorf("error globbing layouts: %w", err)
	}

	for _, layout := range layouts {
		name := strings.TrimSuffix(filepath.Base(layout), ".html")
		if name == "defaultLayout" {
			continue // Skip the default layout as we've already parsed it
		}

		// Clone the default layout and add the specific layout template
		tmpl, err := template.Must(defaultLayout.Clone()).ParseFiles(layout)
		if err != nil {
			return fmt.Errorf("error parsing layout %s: %w", name, err)
		}

		// Add all fractions to this layout
		for _, fraction := range fractions {
			_, err := tmpl.ParseFiles(fraction)
			if err != nil {
				return fmt.Errorf("error adding fraction to layout %s: %w", name, err)
			}
		}

		tmplConstruct.layouts[name] = tmpl
	}

	fmt.Printf("layouts: %v\n", tmplConstruct.layouts)
	fmt.Printf("fractions: %v\n", tmplConstruct.fractions)
	return nil
}

func RenderLayoutTemplate(w http.ResponseWriter, r *http.Request, templateName string, data TemplateData) {
	// Retrieve and validate authStatus from context
	authStatus, err := auth.GetAuthStatusFromContext(r.Context())
	if err != nil {
		log.Printf("Error getting authStatus: %v", err)
		// Default to guest if there's an error
		authStatus = "guest"
	}
	data.Data["AuthStatus"] = authStatus

	log.Printf("authStatus set in template data: %v", authStatus)

	// Retrieve the template
	tmpl, ok := tmplConstruct.layouts[templateName]
	if !ok {
		log.Printf("Template not found: %s", templateName)
		http.Error(w, fmt.Sprintf("The template %s does not exist.", templateName), http.StatusInternalServerError)
		return
	}

	// Set content type and execute template
	w.Header().Set("Content-Type", "text/html")
	err = tmpl.ExecuteTemplate(w, "defaultLayout", data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}

func RenderTemplateFraction(w http.ResponseWriter, templateName string, data TemplateData) {
	fmt.Println("Rendering fraction:", templateName)
	tmpl, ok := tmplConstruct.fractions[templateName]
	if !ok {
		http.Error(w, fmt.Sprintf("The template fraction %s does not exist.", templateName), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	err := tmpl.ExecuteTemplate(w, templateName, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
