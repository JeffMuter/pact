package pages

import (
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"text/template"
)

type TemplateRenderer struct {
	templates   map[string]*template.Template
	templateDir string
	baseLayout  string
	partials    []string
}

func NewTemplateRenderer(templateDir, baseLayout string, partials ...string) *TemplateRenderer {
	return &TemplateRenderer{
		templates:   make(map[string]*template.Template),
		templateDir: templateDir,
		baseLayout:  baseLayout,
		partials:    partials,
	}
}

func (tr *TemplateRenderer) LoadTemplates() error {
	return filepath.Walk(tr.templateDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".html" {
			return nil
		}

		templateName := filepath.Base(path)
		if templateName == tr.baseLayout {
			return nil
		}

		files := append([]string{filepath.Join(tr.templateDir, tr.baseLayout)}, path)
		for _, partial := range tr.partials {
			files = append(files, filepath.Join(tr.templateDir, partial))
		}

		tmpl, err := template.ParseFiles(files...)
		if err != nil {
			return fmt.Errorf("error parsing template %s: %v", templateName, err)
		}

		tr.templates[templateName] = tmpl
		return nil
	})
}

func (tr *TemplateRenderer) RenderTemplate(w http.ResponseWriter, templateName string, data interface{}) error {
	tmpl, ok := tr.templates[templateName]
	if !ok {
		return fmt.Errorf("template %s not found", templateName)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, "base", data)
}
