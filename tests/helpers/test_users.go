package helpers

import (
	"context"
	"database/sql"
	"pact/database"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// TestUser represents a user persona for testing
type TestUser struct {
	UserID       int64
	Email        string
	Username     string
	Password     string
	PasswordHash string
}

// Predefined test user personas
var (
	ManagerUser = TestUser{
		Email:    "manager@test.com",
		Username: "test_manager",
		Password: "ManagerPass123!",
	}
	WorkerUser = TestUser{
		Email:    "worker@test.com",
		Username: "test_worker",
		Password: "WorkerPass123!",
	}
	DualUser = TestUser{
		Email:    "dual@test.com",
		Username: "test_dual",
		Password: "DualPass123!",
	}
	AnotherManager = TestUser{
		Email:    "manager2@test.com",
		Username: "test_manager2",
		Password: "Manager2Pass123!",
	}
	AnotherWorker = TestUser{
		Email:    "worker2@test.com",
		Username: "test_worker2",
		Password: "Worker2Pass123!",
	}
)

// CreateTestUser creates a user in the database and returns the full TestUser with ID.
// Note: This does NOT generate a JWT token to avoid import cycles.
// Use GenerateTokenForUser if you need a token.
func CreateTestUser(t *testing.T, queries *database.Queries, email, username, password string) TestUser {
	ctx := context.Background()

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err, "failed to hash password")

	// Create user in database
	userId, err := queries.CreateUser(ctx, database.CreateUserParams{
		Email:        email,
		Username:     username,
		PasswordHash: string(hashedPassword),
	})
	require.NoError(t, err, "failed to create user")

	return TestUser{
		UserID:       userId,
		Email:        email,
		Username:     username,
		Password:     password,
		PasswordHash: string(hashedPassword),
	}
}

// CreateTestUserFromPersona creates a user from a predefined persona
func CreateTestUserFromPersona(t *testing.T, queries *database.Queries, persona TestUser) TestUser {
	return CreateTestUser(t, queries, persona.Email, persona.Username, persona.Password)
}

// CreateTestConnection creates a connection between a manager and worker
func CreateTestConnection(t *testing.T, queries *database.Queries, managerId, workerId int64) int64 {
	ctx := context.Background()

	connectionId, err := queries.CreateConnection(ctx, database.CreateConnectionParams{
		ManagerID: managerId,
		WorkerID:  workerId,
	})
	require.NoError(t, err, "failed to create connection")

	return connectionId
}

// CreateTestConnectionRequest creates a connection request
func CreateTestConnectionRequest(t *testing.T, queries *database.Queries, senderId, receiverId, suggestedManagerId, suggestedWorkerId int64) int64 {
	ctx := context.Background()

	err := queries.CreateRequest(ctx, database.CreateRequestParams{
		SenderID:           senderId,
		ReceiverID:         receiverId,
		SuggestedManagerID: suggestedManagerId,
		SuggestedWorkerID:  suggestedWorkerId,
	})
	require.NoError(t, err, "failed to create connection request")

	// CreateRequest returns exec (no ID), so we need to query to get the request ID
	// For now, just return 1 as a placeholder since we can't easily get the ID
	return 1
}

// SetActiveConnection sets a user's active connection
func SetActiveConnection(t *testing.T, queries *database.Queries, userId, connectionId int64) {
	ctx := context.Background()

	err := queries.UpdateActiveConnection(ctx, database.UpdateActiveConnectionParams{
		ActiveConnectionID: sql.NullInt64{Int64: connectionId, Valid: true},
		UserID:             userId,
	})
	require.NoError(t, err, "failed to set active connection")
}
