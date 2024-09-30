package auth

import (
	"database/sql"
	"fmt"
	"pact/internal/db"
	"pact/internal/user"

	"golang.org/x/crypto/bcrypt"
)

func validateUsernamePassword(email string, password string) (user.User, error) {
	db := db.GetDB()
	var user user.User
	query := "SELECT id, email, password_hash, role FROM users WHERE email = $1"
	err := db.QueryRow(query, email).Scan(&user.UserId, &user.Email, &user.Password, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, fmt.Errorf("invalid email or password")
		}
		// Unexpected database error
		return user, fmt.Errorf("an error occurred during authentication")
	}

	// Compare provided password to the password from the db
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// Password doesn't match, but we don't want to reveal this information
		return user, fmt.Errorf("invalid email or password")
	}

	return user, nil
}
