package pages

import (
	"fmt"
	"net/http"
	"path/filepath"
	"text/template"
)

var templates map[string]*template.Template

func InitTemplates() error {
	templates = make(map[string]*template.Template)

	layouts, err := filepath.Glob("./internal/templates/contentTemplates/*.html")
	if err != nil {
		return err
	}
	fmt.Println(len(layouts))
	fmt.Println(layouts)

	includes, err := filepath.Glob("./internal/templates/fractions/*.html")
	if err != nil {
		return err
	}
	fmt.Println(len(includes))
	fmt.Println(includes)

	// Generate our templates map from our layouts/ and includes/ directories
	for _, layout := range layouts {
		includes = append(includes, layout)
		name := filepath.Base(layout)
		templates[name] = template.Must(template.ParseFiles(includes...))
		fmt.Printf("template added: %s. Len of map: %d\n", name, len(templates))
	}
	return nil
}

func RenderTemplate(w http.ResponseWriter, name string, data interface{}) {
	tmpl, ok := templates[name+".html"]
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
