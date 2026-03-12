package support

import (
	"context"
	"database/sql"
	"pact/database"
)

// CreateSupportTicket submits a support ticket to the database
func CreateSupportTicket(db *sql.DB, userID int, email, issueDescription string) error {
	queries := database.New(db)
	ctx := context.Background()

	err := queries.CreateSupportTicket(ctx, database.CreateSupportTicketParams{
		UserID:           int64(userID),
		Email:            email,
		IssueDescription: issueDescription,
	})

	return err
}
