package pages

import (
	"fmt"
	"net/http"
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

func RenderLayoutTemplate(w http.ResponseWriter, name string, data interface{}) {
	fmt.Println("Rendering layout template:", name)
	tmpl, ok := tmplConstruct.layouts[name]
	if !ok {
		http.Error(w, fmt.Sprintf("The template %s does not exist.", name), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	err := tmpl.ExecuteTemplate(w, "defaultLayout", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func RenderTemplateFraction(w http.ResponseWriter, name string, data interface{}) {
	fmt.Println("Rendering fraction:", name)
	tmpl, ok := tmplConstruct.fractions[name]
	if !ok {
		http.Error(w, fmt.Sprintf("The template fraction %s does not exist.", name), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	err := tmpl.ExecuteTemplate(w, name, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
