package user

import (
	"context"
	"fmt"
	"net/http"
	"pact/database"
	"pact/internal/pages"
)

// ShowRegistrationForm renders the user registration form.
func ShowRegistrationForm(w http.ResponseWriter, r *http.Request) {
	// Create TemplateData
	data := pages.TemplateData{
		Data: map[string]string{
			"Title": "Registration",
		},
	}

	// Render the template
	pages.RenderLayoutTemplate(w, r, "registerForm", data)
}

func GetUserByEmail(email string) (*database.User, error) {
	queries := database.GetQueries()
	ctx := context.Background()

	user, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		return &user, fmt.Errorf("error scanning row from query to get a user from an email(e: %s): %w", email, err)
	}

	return &user, nil
}
