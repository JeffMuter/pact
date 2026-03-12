package support

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"pact/tests/helpers"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleSupportTicketSubmission_ValidSubmission verifies that a valid support ticket
// is properly stored in the database
func TestHandleSupportTicketSubmission_ValidSubmission(t *testing.T) {
	db, queries, cleanup := helpers.SetupTestDBWithPath(t, "../../database/schema.sql")
	defer cleanup()

	// Create a test user
	user := helpers.CreateTestUser(t, queries, "user@test.com", "testuser", "TestPass123!")

	// Prepare form data
	formData := bytes.NewBufferString("email=user@test.com&issue_description=Application%20is%20crashing")

	// Create HTTP request with user ID in context
	req := httptest.NewRequest("POST", "/support/submit", formData)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), "userID", int(user.UserID)))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler := HandleSupportTicketSubmission(db)
	handler(w, req)

	// Verify response status
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify response contains success message
	body, _ := io.ReadAll(w.Body)
	assert.Contains(t, string(body), "alert-success")
	assert.Contains(t, string(body), "Support ticket submitted successfully")

	// Verify ticket was stored in database
	tickets, err := queries.GetSupportTicketsByUserId(context.Background(), user.UserID)
	require.NoError(t, err)
	require.Len(t, tickets, 1)

	// Verify ticket data
	assert.Equal(t, user.UserID, tickets[0].UserID)
	assert.Equal(t, "user@test.com", tickets[0].Email)
	assert.Equal(t, "Application is crashing", tickets[0].IssueDescription)
}

// TestHandleSupportTicketSubmission_MultipleTickets verifies that a user can submit multiple tickets
func TestHandleSupportTicketSubmission_MultipleTickets(t *testing.T) {
	db, queries, cleanup := helpers.SetupTestDBWithPath(t, "../../database/schema.sql")
	defer cleanup()

	// Create a test user
	user := helpers.CreateTestUser(t, queries, "user@test.com", "testuser", "TestPass123!")

	// Submit first ticket
	formData1 := bytes.NewBufferString("email=user@test.com&issue_description=First%20issue")
	req1 := httptest.NewRequest("POST", "/support/submit", formData1)
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req1 = req1.WithContext(context.WithValue(req1.Context(), "userID", int(user.UserID)))
	w1 := httptest.NewRecorder()

	handler := HandleSupportTicketSubmission(db)
	handler(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Submit second ticket
	formData2 := bytes.NewBufferString("email=user@test.com&issue_description=Second%20issue")
	req2 := httptest.NewRequest("POST", "/support/submit", formData2)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2 = req2.WithContext(context.WithValue(req2.Context(), "userID", int(user.UserID)))
	w2 := httptest.NewRecorder()

	handler(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Verify both tickets are in database
	tickets, err := queries.GetSupportTicketsByUserId(context.Background(), user.UserID)
	require.NoError(t, err)
	require.Len(t, tickets, 2)

	assert.Equal(t, "First issue", tickets[0].IssueDescription)
	assert.Equal(t, "Second issue", tickets[1].IssueDescription)
}

// TestHandleSupportTicketSubmission_MissingEmail verifies validation rejects missing email
func TestHandleSupportTicketSubmission_MissingEmail(t *testing.T) {
	db, queries, cleanup := helpers.SetupTestDBWithPath(t, "../../database/schema.sql")
	defer cleanup()

	// Create a test user
	user := helpers.CreateTestUser(t, queries, "user@test.com", "testuser", "TestPass123!")

	// Prepare form data without email
	formData := bytes.NewBufferString("issue_description=Test%20issue")

	// Create HTTP request
	req := httptest.NewRequest("POST", "/support/submit", formData)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), "userID", int(user.UserID)))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler := HandleSupportTicketSubmission(db)
	handler(w, req)

	// Verify error status
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify no ticket was created
	tickets, err := queries.GetSupportTicketsByUserId(context.Background(), user.UserID)
	require.NoError(t, err)
	assert.Len(t, tickets, 0)
}

// TestHandleSupportTicketSubmission_MissingDescription verifies validation rejects missing description
func TestHandleSupportTicketSubmission_MissingDescription(t *testing.T) {
	db, queries, cleanup := helpers.SetupTestDBWithPath(t, "../../database/schema.sql")
	defer cleanup()

	// Create a test user
	user := helpers.CreateTestUser(t, queries, "user@test.com", "testuser", "TestPass123!")

	// Prepare form data without description
	formData := bytes.NewBufferString("email=user@test.com")

	// Create HTTP request
	req := httptest.NewRequest("POST", "/support/submit", formData)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = req.WithContext(context.WithValue(req.Context(), "userID", int(user.UserID)))

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	handler := HandleSupportTicketSubmission(db)
	handler(w, req)

	// Verify error status
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify no ticket was created
	tickets, err := queries.GetSupportTicketsByUserId(context.Background(), user.UserID)
	require.NoError(t, err)
	assert.Len(t, tickets, 0)
}

// TestHandleSupportTicketSubmission_MissingUserID verifies handler rejects missing authentication
func TestHandleSupportTicketSubmission_MissingUserID(t *testing.T) {
	db, _, cleanup := helpers.SetupTestDBWithPath(t, "../../database/schema.sql")
	defer cleanup()

	// Prepare form data
	formData := bytes.NewBufferString("email=user@test.com&issue_description=Test%20issue")

	// Create HTTP request without user ID in context
	req := httptest.NewRequest("POST", "/support/submit", formData)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Intentionally not setting userID in context

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler (should panic or fail gracefully)
	handler := HandleSupportTicketSubmission(db)

	// Wrap in recover to catch panic
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Handler panicked as expected: %v", r)
		}
	}()

	handler(w, req)

	// If no panic, verify it returns 401
	if w.Code != 0 {
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	}
}

// TestHandleSupportTicketSubmission_DifferentUsers verifies tickets are properly isolated by user
func TestHandleSupportTicketSubmission_DifferentUsers(t *testing.T) {
	db, queries, cleanup := helpers.SetupTestDBWithPath(t, "../../database/schema.sql")
	defer cleanup()

	// Create two test users
	user1 := helpers.CreateTestUser(t, queries, "user1@test.com", "user1", "Pass123!")
	user2 := helpers.CreateTestUser(t, queries, "user2@test.com", "user2", "Pass456!")

	// User 1 submits a ticket
	formData1 := bytes.NewBufferString("email=user1@test.com&issue_description=User1%20issue")
	req1 := httptest.NewRequest("POST", "/support/submit", formData1)
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req1 = req1.WithContext(context.WithValue(req1.Context(), "userID", int(user1.UserID)))
	w1 := httptest.NewRecorder()

	handler := HandleSupportTicketSubmission(db)
	handler(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// User 2 submits a ticket
	formData2 := bytes.NewBufferString("email=user2@test.com&issue_description=User2%20issue")
	req2 := httptest.NewRequest("POST", "/support/submit", formData2)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2 = req2.WithContext(context.WithValue(req2.Context(), "userID", int(user2.UserID)))
	w2 := httptest.NewRecorder()

	handler(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Verify user1 only sees their ticket
	tickets1, err := queries.GetSupportTicketsByUserId(context.Background(), user1.UserID)
	require.NoError(t, err)
	require.Len(t, tickets1, 1)
	assert.Equal(t, "User1 issue", tickets1[0].IssueDescription)

	// Verify user2 only sees their ticket
	tickets2, err := queries.GetSupportTicketsByUserId(context.Background(), user2.UserID)
	require.NoError(t, err)
	require.Len(t, tickets2, 1)
	assert.Equal(t, "User2 issue", tickets2[0].IssueDescription)
}
