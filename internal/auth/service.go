package auth

import (
	"context"
	"database/sql"
	"fmt"
	"pact/database"

	"golang.org/x/crypto/bcrypt"
)

func validateUsernamePassword(email string, password string) (database.User, error) {
	var user database.User

	queries := database.GetQueries()
	ctx := context.Background()

	user, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, fmt.Errorf("invalid email or password")
		}
		// Unexpected database error
		return user, fmt.Errorf("an error occurred during authentication")
	}

	// Compare provided password to the password from the db
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		// Password doesn't match, but we don't want to reveal this information
		return user, fmt.Errorf("invalid email or password")
	}

	return user, nil
}
