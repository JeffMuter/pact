package user

import (
	"fmt"
	"net/http"
	"pact/internal/db"
	"pact/internal/pages"
)

// ShowRegistrationForm renders the user registration form.
func ShowRegistrationForm(w http.ResponseWriter, r *http.Request) {
	// Create TemplateData
	data := pages.TemplateData{
		Page: pages.Page{
			Title:   "User Registration",
			Heading: "Register for an Account",
		},
	}

	// Render the template
	pages.RenderTemplate(w, "internal/templates/user/registration.html", data)
}

func GetUserByEmail(email string) (*User, error) {
	user := makeUser()
	db := db.GetDB()
	query := `SELECT * FROM users WHERE id = $1`
	err := db.QueryRow(query, email).Scan(
		&user.UserId,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Role,
		&user.Created_at,
		&user.Updated_at,
	)
	if err != nil {
		return user, fmt.Errorf("error scanning row from query to get a user from an email(e: %s): %w", email, err)
	}

	return user, nil
}
