package auth

import (
	"fmt"
	"pact/internal/db"

	"golang.org/x/crypto/bcrypt"
)

func validateUsernamePassword(email string, password string) error {
	db := db.GetDB()

	var hashedPassword, foundEmail string
	query := "SELECT email, password_hash FROM users WHERE email = $1"

	err := db.QueryRow(query, email).Scan(&foundEmail, &hashedPassword)
	if err != nil {
		return fmt.Errorf("error, something went wrong finding a user's email & password using the email and password provided: %w", err)
	}

	if email != foundEmail {
		return fmt.Errorf("error, given email(e: %s) does not match any existing emails in storage: %w", email, err)
	}
	// compare provided password to the password from the db
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return fmt.Errorf("error, this email exists, but this password did not match your current password: %w", err)
	}

	return nil
}
