package render

import (
	"html/template"
	"net/http"
	"pact/handlers"
	"pact/internal/models"
	"pact/middleware"
)

// RenderTemplate renders an HTML template and injects the provided data.
func RenderTemplate(w http.ResponseWriter, templatePath string, data models.TemplateData) {
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

func Router() *http.ServeMux {

	mux := http.NewServeMux()

	// Register handlers
	mux.Handle("/", middleware.AuthMiddleware(http.HandlerFunc(handlers.ServeHomePage)))
	mux.HandleFunc("/post/", handlers.ServePostPage)

	mux.HandleFunc("GET /login", handlers.ServeLoginPage)
	mux.HandleFunc("POST /login", handlers.LoginFormHandler)

	mux.HandleFunc("GET /registeruser", handlers.ServeRegistrationPage)
	mux.HandleFunc("POST /registeruser", handlers.RegisterHandler)

	return mux

}
