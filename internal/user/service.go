package user

import (
	"context"
	"pact/database"
)

// makeUser creates a blank pointer to an uninitialized user.
func makeUser() *database.User {
	return &database.User{}
}

// GetAccountPageData retrieves all account page data for a user
func GetAccountPageData(ctx context.Context, userId int64) (database.GetAccountPageDataRow, error) {
	queries := database.GetQueries()
	return queries.GetAccountPageData(ctx, userId)
}
