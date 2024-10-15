package user

import "pact/database"

// makeUser creates a blank pointer to an uninitialized user.
func makeUser() *database.User {
	return &database.User{}
}
