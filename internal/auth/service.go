package auth

import (
	"fmt"
	"pact/internal/db"
	"pact/internal/user"

	"golang.org/x/crypto/bcrypt"
)

func validateUsernamePassword(email string, password string) (user.User, error) {
	db := db.GetDB()
	var user user.User

	query := "SELECT user.Id, email, password_hash, role FROM users WHERE email = $1"

	err := db.QueryRow(query, email).Scan(&user.UserId, &user.Email, &user.Password, &user.Role)
	if err != nil {
		return user, fmt.Errorf("error, something went wrong finding a user's email & password using the email and password provided: %w", err)
	}

	// compare provided password to the password from the db
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return user, fmt.Errorf("error, this email exists, but this password did not match your current password: %w", err)
	}

	return user, nil
}
