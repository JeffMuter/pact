package auth

import (
	"database/sql"
	"fmt"
	"pact/internal/db"

	"golang.org/x/crypto/bcrypt"
)

func validateUsernamePassword(email string, password string) (bool, error) {
	// open db connection
	db := db.GetDB()

	var hashedPassword string
	query := "SELECT password FROM users WHERE email = $1"
	err := db.QueryRow(query, email).Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, err
		}
		return false, err
	}

	// compare provided password to the password from the db
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false, fmt.Errorf("invalid password")
	}

	return true, nil
}
