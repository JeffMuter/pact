package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetupTestDB verifies the test database helper works
func TestSetupTestDB(t *testing.T) {
	db, queries, cleanup := SetupTestDBWithPath(t, "../../database/schema.sql")
	defer cleanup()

	// Verify database connection
	err := db.Ping()
	require.NoError(t, err)

	// Verify queries object is valid
	assert.NotNil(t, queries)
}

// TestCreateTestUser verifies user creation helper works
func TestCreateTestUser(t *testing.T) {
	_, queries, cleanup := SetupTestDBWithPath(t, "../../database/schema.sql")
	defer cleanup()

	// Create a test user
	user := CreateTestUser(t, queries, "test@example.com", "testuser", "TestPass123!")

	// Verify user was created
	assert.Greater(t, user.UserID, int64(0))
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "TestPass123!", user.Password)
	assert.NotEmpty(t, user.PasswordHash)

	// Verify password hash is not plaintext
	assert.NotEqual(t, user.Password, user.PasswordHash)
}

// TestCreateTestUserFromPersona verifies persona-based user creation
func TestCreateTestUserFromPersona(t *testing.T) {
	_, queries, cleanup := SetupTestDBWithPath(t, "../../database/schema.sql")
	defer cleanup()

	// Create manager and worker users
	manager := CreateTestUserFromPersona(t, queries, ManagerUser)
	worker := CreateTestUserFromPersona(t, queries, WorkerUser)

	// Verify users were created with different IDs
	assert.Greater(t, manager.UserID, int64(0))
	assert.Greater(t, worker.UserID, int64(0))
	assert.NotEqual(t, manager.UserID, worker.UserID)
	assert.Equal(t, "manager@test.com", manager.Email)
	assert.Equal(t, "worker@test.com", worker.Email)
}

// TestCreateTestConnection verifies connection creation helper works
func TestCreateTestConnection(t *testing.T) {
	_, queries, cleanup := SetupTestDBWithPath(t, "../../database/schema.sql")
	defer cleanup()

	// Create two users
	manager := CreateTestUser(t, queries, "manager@test.com", "manager", "ManagerPass!")
	worker := CreateTestUser(t, queries, "worker@test.com", "worker", "WorkerPass!")

	// Create connection between them
	connectionId := CreateTestConnection(t, queries, manager.UserID, worker.UserID)

	// Verify connection was created
	assert.Greater(t, connectionId, int64(0))
}

// TestCreateTestConnectionRequest verifies connection request creation helper works
func TestCreateTestConnectionRequest(t *testing.T) {
	_, queries, cleanup := SetupTestDBWithPath(t, "../../database/schema.sql")
	defer cleanup()

	// Create two users
	sender := CreateTestUser(t, queries, "sender@test.com", "sender", "SenderPass!")
	receiver := CreateTestUser(t, queries, "receiver@test.com", "receiver", "ReceiverPass!")

	// Create connection request (sender wants to be worker, receiver to be manager)
	requestId := CreateTestConnectionRequest(t, queries, sender.UserID, receiver.UserID, receiver.UserID, sender.UserID)

	// Verify request was created
	assert.Greater(t, requestId, int64(0))
}

// TestSetActiveConnection verifies setting active connection works
func TestSetActiveConnection(t *testing.T) {
	_, queries, cleanup := SetupTestDBWithPath(t, "../../database/schema.sql")
	defer cleanup()

	// Create users and connection
	manager := CreateTestUser(t, queries, "manager@test.com", "manager", "ManagerPass!")
	worker := CreateTestUser(t, queries, "worker@test.com", "worker", "WorkerPass!")
	connectionId := CreateTestConnection(t, queries, manager.UserID, worker.UserID)

	// Set active connection for manager
	SetActiveConnection(t, queries, manager.UserID, connectionId)

	// This test just verifies no error occurred
	// A more thorough test would query the database to verify the value was set
}
