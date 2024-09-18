package user

import (
	"net/http"
	"pact/internal/models"
	"pact/internal/render"
)

// ShowRegistrationForm renders the user registration form.
func ShowRegistrationForm(w http.ResponseWriter, r *http.Request) {
	// Create TemplateData
	data := models.TemplateData{
		Page: models.Page{
			Title:   "User Registration",
			Heading: "Register for an Account",
		},
	}

	// Render the template
	render.RenderTemplate(w, "internal/templates/user/registration.html", data)
}
