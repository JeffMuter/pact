package main

import (
	"log"
	"net/http"
	"pact/internal/pages"
)

type TemplateData struct {
	Title   string
	Content interface{}
}

func main() {
	tr := pages.NewTemplateRenderer(
		"./templates",
		"layout.html",
		"navbar.html",
		"footer.html",
	)

	if err := tr.LoadTemplates(); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		data := TemplateData{
			Title: "Login",
			Content: map[string]string{
				"Heading": "Login",
			},
		}
		if err := tr.RenderTemplate(w, "login.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		data := TemplateData{
			Title: "Register",
			Content: map[string]string{
				"Heading": "Registration Page",
			},
		}
		if err := tr.RenderTemplate(w, "register.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
