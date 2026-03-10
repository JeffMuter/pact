package pages

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pact/internal/auth"
	"path/filepath"
	"strings"
	"text/template"
)

var funcMap = template.FuncMap{
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("dict requires an even number of arguments")
		}
		dict := make(map[string]interface{})
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	},
	"slice": func(values ...interface{}) []interface{} {
		return values
	},
	"printf": fmt.Sprintf,
	"parseJSON": func(jsonStr string) []string {
		var result []string
		if jsonStr == "" || jsonStr == "null" {
			return result
		}
		err := json.Unmarshal([]byte(jsonStr), &result)
		if err != nil {
			log.Printf("Error parsing JSON: %v", err)
			return result
		}
		return result
	},
}

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



	// Parse all fraction templates into one master template
	fractions, err := filepath.Glob("./internal/templates/fractions/*.html")
	if err != nil {
		return fmt.Errorf("error globbing fractions: %w", err)
	}

	// Create a master template with all fractions
	masterFractions := template.New("master").Funcs(funcMap)
	for _, fraction := range fractions {
		_, err := masterFractions.ParseFiles(fraction)
		if err != nil {
			return fmt.Errorf("error parsing fraction %s into master: %w", fraction, err)
		}
	}

	// For each fraction, create a template that includes all fractions
	for _, fraction := range fractions {
		name := strings.TrimSuffix(filepath.Base(fraction), ".html")

		// Clone the master fractions (which has everything)
		tmpl, err := masterFractions.Clone()
		if err != nil {
			return fmt.Errorf("error cloning master for fraction %s: %w", name, err)
		}

		// Debug: print available templates
		if name == "description" {
			fmt.Printf("Templates available for description: %d\n", len(tmpl.Templates()))
			for _, t := range tmpl.Templates() {
				fmt.Printf("  - %s\n", t.Name())
			}
		}

		// Store in fractions map
		tmplConstruct.fractions[name] = tmpl
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

		// Create a new template with defaultLayout
		tmpl := template.New("layoutTemplate").Funcs(funcMap)
		
		// Parse defaultLayout first
		_, err := tmpl.ParseFiles("./internal/templates/contentTemplates/defaultLayout.html")
		if err != nil {
			return fmt.Errorf("error parsing defaultLayout: %w", err)
		}

		// Parse the specific layout file (which defines "content" block)
		_, err = tmpl.ParseFiles(layout)
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

func RenderLayoutTemplate(w http.ResponseWriter, r *http.Request, templateName string, data interface{}) {
	// Retrieve and validate authStatus from context
	authStatus, err := auth.GetAuthStatusFromContext(r.Context())
	if err != nil {
		log.Printf("Error getting authStatus: %v", err)
		// Default to guest if there's an error
		authStatus = "guest"
	}
	if td, ok := data.(TemplateData); ok {
		td.Data["AuthStatus"] = authStatus
		data = td
	} else if bd, ok := data.(interface{ GetData() map[string]any }); ok {
		bd.GetData()["AuthStatus"] = authStatus
	}

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

func RenderTemplateFraction(w http.ResponseWriter, templateName string, data interface{}) {
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
