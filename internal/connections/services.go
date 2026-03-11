package connections

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"pact/database"
)

// CreateConnectionRequest takes the current user's id and the email they submitted
// to send a request. Returns an error if the receiver doesn't exist or the request
// can't be created (e.g. duplicate, self-request).
func CreateConnectionRequest(ctx context.Context, userId int, senderRole, email string) error {
	queries := database.GetQueries()

	receiverUser, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("could not find a user with that email: %w", err)
	}

	var suggestedManager, suggestedWorker int64
	switch senderRole {
	case "manager":
		suggestedManager = int64(userId)
		suggestedWorker = receiverUser.UserID
	case "worker":
		suggestedManager = receiverUser.UserID
		suggestedWorker = int64(userId)
	default:
		return fmt.Errorf("invalid sender role: %s", senderRole)
	}

	err = queries.CreateRequest(ctx, database.CreateRequestParams{
		SenderID:           int64(userId),
		ReceiverID:         receiverUser.UserID,
		SuggestedManagerID: suggestedManager,
		SuggestedWorkerID:  suggestedWorker,
	})
	if err != nil {
		return fmt.Errorf("could not create connection request: %w", err)
	}

	return nil
}

// getUsersPendingConnectionRequests returns all active incoming requests for a user.
func getUsersPendingConnectionRequests(ctx context.Context, userId int) ([]database.GetUserPendingRequestsRow, error) {
	queries := database.GetQueries()

	rows, err := queries.GetUserPendingRequests(ctx, int64(userId))
	if err != nil {
		return nil, fmt.Errorf("could not get pending requests for user %d: %w", userId, err)
	}

	return rows, nil
}

// acceptConnectionRequest looks up the request by ID, creates the connection using
// the roles stored in the request, and deactivates the request — all within a
// single transaction so partial state can't occur.
func acceptConnectionRequest(ctx context.Context, requestId int64) error {
	db := database.GetDB()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := database.New(tx)

	req, err := qtx.GetConnectionRequestById(ctx, requestId)
	if err != nil {
		return fmt.Errorf("could not find connection request %d: %w", requestId, err)
	}
	if req.IsActive == 0 {
		return fmt.Errorf("connection request %d is no longer active", requestId)
	}

	_, err = qtx.CreateConnection(ctx, database.CreateConnectionParams{
		ManagerID: req.SuggestedManagerID,
		WorkerID:  req.SuggestedWorkerID,
	})
	if err != nil {
		return fmt.Errorf("could not create connection: %w", err)
	}

	err = qtx.DeactivateConnectionRequest(ctx, requestId)
	if err != nil {
		return fmt.Errorf("could not deactivate request %d: %w", requestId, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	log.Printf("connection created from request %d (manager=%d, worker=%d)",
		requestId, req.SuggestedManagerID, req.SuggestedWorkerID)

	return nil
}

// rejectConnectionRequest deactivates a request without creating a connection.
func rejectConnectionRequest(ctx context.Context, requestId int64) error {
	db := database.GetDB()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := database.New(tx)

	req, err := qtx.GetConnectionRequestById(ctx, requestId)
	if err != nil {
		return fmt.Errorf("could not find connection request %d: %w", requestId, err)
	}
	if req.IsActive == 0 {
		return fmt.Errorf("connection request %d is already inactive", requestId)
	}

	err = qtx.DeactivateConnectionRequest(ctx, requestId)
	if err != nil {
		return fmt.Errorf("could not deactivate request %d: %w", requestId, err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	return nil
}

// getConnectionsByUserId returns all active connections for a user.
func getConnectionsByUserId(ctx context.Context, userId int) ([]database.GetConnectionsByIdRow, error) {
	queries := database.GetQueries()

	rows, err := queries.GetConnectionsById(ctx, int64(userId))
	if err != nil {
		return nil, fmt.Errorf("could not get connections for user %d: %w", userId, err)
	}

	return rows, nil
}

// updateActiveConnection sets the given connection as active for the given user.
func updateActiveConnection(ctx context.Context, userId, connectionId int) error {
	queries := database.GetQueries()

	err := queries.UpdateActiveConnection(ctx, database.UpdateActiveConnectionParams{
		ActiveConnectionID: sql.NullInt64{Int64: int64(connectionId), Valid: true},
		UserID:             int64(userId),
	})
	if err != nil {
		return fmt.Errorf("could not update active connection for user %d: %w", userId, err)
	}

	return nil
}

// getActiveConnectionDetails returns the partner's id, username, and role for
// the current user's active connection.
func getActiveConnectionDetails(ctx context.Context, userId int) (int, string, string, error) {
	queries := database.GetQueries()

	params := database.GetActiveConnectionUserDetailsParams{
		ManagerID:   int64(userId),
		ManagerID_2: int64(userId),
		ManagerID_3: int64(userId),
		UserID:      int64(userId),
	}

	row, err := queries.GetActiveConnectionUserDetails(ctx, params)
	if err != nil {
		return 0, "", "", fmt.Errorf("could not get active connection details: %w", err)
	}

	userIDInt64, ok := row.UserID.(int64)
	if !ok {
		return 0, "", "", fmt.Errorf("active connection user ID type assertion failed")
	}

	return int(userIDInt64), row.Username, row.Role, nil
}

// deleteConnection removes a connection after verifying the user is part of it.
// If the connection being deleted is the user's active connection, it clears that first.
func deleteConnection(ctx context.Context, connectionId int, userId int) error {
	db := database.GetDB()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := database.New(tx)

	// Verify user is part of this connection
	conn, err := qtx.GetActiveConnectionDetails(ctx, int64(connectionId))
	if err != nil {
		return fmt.Errorf("connection %d not found: %w", connectionId, err)
	}

	// Check user is either manager or worker
	if int64(userId) != conn.ManagerID && int64(userId) != conn.WorkerID {
		return fmt.Errorf("user %d is not part of connection %d", userId, connectionId)
	}

	// Clear active_connection_id if this connection is active for the user
	err = qtx.ClearActiveConnectionIfMatch(ctx, database.ClearActiveConnectionIfMatchParams{
		UserID:             int64(userId),
		ActiveConnectionID: sql.NullInt64{Int64: int64(connectionId), Valid: true},
	})
	if err != nil {
		return fmt.Errorf("could not clear active connection: %w", err)
	}

	// Delete the connection
	err = qtx.DeleteConnection(ctx, int64(connectionId))
	if err != nil {
		return fmt.Errorf("could not delete connection: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	log.Printf("connection %d deleted by user %d", connectionId, userId)
	return nil
}
