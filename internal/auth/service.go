package auth

import (
	"context"
	"database/sql"
	"fmt"
	"pact/database"

	"golang.org/x/crypto/bcrypt"
)

func ValidateUsernamePassword(email string, password string) (database.User, error) {
	fmt.Println(email + " " + password)
	var user database.User

	queries := database.GetQueries()
	ctx := context.Background()

	user, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("error getting user by email: no rows returned")
			return user, fmt.Errorf("invalid email or password")
		}
		fmt.Printf("get user by email query != nil: %v\n", err)
		// Unexpected database error
		return user, fmt.Errorf("an error occurred during authentication")
	}

	// Compare provided password to the password from the db
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		fmt.Println("error: invalid email")
		// Password doesn't match, but we don't want to reveal this information
		return user, fmt.Errorf("invalid email or password")
	}

	return user, nil
}
func GetAuthStatusFromContext(ctx context.Context) (string, error) {
	authStatusValue := ctx.Value("authStatus")
	if authStatusValue == nil {
		return "guest", fmt.Errorf("authStatus not found in context")
	}

	authStatus, ok := authStatusValue.(string)
	if !ok {
		return "guest", fmt.Errorf("authStatus is not a string: %v", authStatusValue)
	}

	// Validate authStatus value
	switch authStatus {
	case "guest", "registered", "member":
		return authStatus, nil
	default:
		return "guest", fmt.Errorf("invalid authStatus value: %s", authStatus)
	}
}

func LogoutHandler() error {

}
