# Pact Testing Plan

**Last Updated:** March 4, 2026  
**Status:** Initial Plan - Zero to Production-Ready Testing

---

## Executive Summary

This document outlines a comprehensive testing strategy for Pact, prioritizing **stable core features** that won't change before launch. We focus on backend business logic, authentication flows, and critical user journeys while avoiding brittle UI-specific tests that will break with styling changes.

**Philosophy:** Test behavior, not implementation. Test contracts, not cosmetics.

---

## Testing Stack Recommendation

### Backend Testing
- **Framework:** Go's `testing` package + `testify` for assertions
- **Database:** In-memory SQLite for fast, isolated tests
- **HTTP Testing:** `httptest` package for handler testing

### End-to-End Testing
- **Framework:** Playwright (Go bindings) or Selenium
- **Focus:** Critical user journeys only
- **Frequency:** Pre-deployment smoke tests

### Tools to Add
```bash
# Go testing dependencies
go get github.com/stretchr/testify
go get github.com/DATA-DOG/go-sqlmock  # For DB mocking if needed

# E2E (choose one)
# Option A: Playwright
npm install -D @playwright/test

# Option B: Selenium with Go bindings
go get github.com/tebeka/selenium
```

---

## Priority 1: Core Backend Unit Tests

These test isolated business logic that will **never change**. No HTTP, no UI, pure Go functions.

### Authentication & Security (`internal/auth/`)

#### `jwt_test.go`
- ✅ `TestGenerateToken_ValidUserID` - Token generation succeeds
- ✅ `TestGenerateToken_ContainsUserID` - Token payload includes correct user ID
- ✅ `TestValidateToken_ValidToken` - Fresh token validates successfully
- ✅ `TestValidateToken_ExpiredToken` - 7-hour-old token fails validation
- ✅ `TestValidateToken_MalformedToken` - Garbage string fails validation
- ✅ `TestValidateToken_WrongSecret` - Token signed with different key fails

#### `service_test.go`
- ✅ `TestHashPassword_NotPlaintext` - Bcrypt hash doesn't equal input
- ✅ `TestCheckPasswordHash_CorrectPassword` - Valid password verifies
- ✅ `TestCheckPasswordHash_WrongPassword` - Invalid password fails
- ✅ `TestCheckPasswordHash_EmptyPassword` - Empty string fails

### Connection Management (`internal/connections/`)

#### `services_test.go`
**Connection Request Creation**
- ✅ `TestCreateConnectionRequest_ValidManager` - Manager role request succeeds
- ✅ `TestCreateConnectionRequest_ValidWorker` - Worker role request succeeds
- ✅ `TestCreateConnectionRequest_InvalidRole` - "admin" role fails
- ✅ `TestCreateConnectionRequest_NonexistentEmail` - Unknown email fails
- ✅ `TestCreateConnectionRequest_SelfRequest` - User requesting self fails (DB constraint)
- ✅ `TestCreateConnectionRequest_DuplicateRequest` - Second identical request fails

**Connection Request Acceptance**
- ✅ `TestAcceptConnectionRequest_CreatesConnection` - Connection record created
- ✅ `TestAcceptConnectionRequest_CorrectRoles` - Manager/worker IDs match request
- ✅ `TestAcceptConnectionRequest_DeactivatesRequest` - is_active set to 0
- ✅ `TestAcceptConnectionRequest_AlreadyAccepted` - Second accept fails
- ✅ `TestAcceptConnectionRequest_NonexistentRequest` - Invalid ID fails

**Connection Request Rejection**
- ✅ `TestRejectConnectionRequest_DeactivatesRequest` - is_active set to 0
- ✅ `TestRejectConnectionRequest_NoConnectionCreated` - connections table unchanged
- ✅ `TestRejectConnectionRequest_AlreadyRejected` - Second reject fails

**Active Connection Management**
- ✅ `TestUpdateActiveConnection_SetsActiveConnection` - user.active_connection_id updated
- ✅ `TestGetActiveConnectionDetails_ReturnsCorrectPartner` - Returns other user, not self
- ✅ `TestGetActiveConnectionDetails_ReturnsCorrectRole` - "manager" or "worker" accurate
- ✅ `TestGetActiveConnectionDetails_NoActiveConnection` - Returns error when null

**Connection Deletion**
- ✅ `TestDeleteConnection_RemovesConnection` - Connection deleted from DB
- ✅ `TestDeleteConnection_ClearsActiveIfMatch` - active_connection_id nulled if matches
- ✅ `TestDeleteConnection_UnauthorizedUser` - User not in connection can't delete
- ✅ `TestDeleteConnection_BothUsersCanDelete` - Manager and worker both authorized

### Task & Reward Management (`internal/buckets/`)

#### `services_test.go`
**Task Assignment**
- ✅ `TestAssignTask_CreatesAssignedTask` - assigned_tasks record created
- ✅ `TestAssignTask_SetsCorrectDueTime` - due_time = assigned_at + duration
- ✅ `TestAssignTask_CopiesTemplateFields` - Points, requirements, timers copied
- ✅ `TestAssignTask_DefaultsToTodoStatus` - status = 'todo'
- ✅ `TestAssignTask_RequiresValidConnection` - Invalid connection_id fails

**Task Submission**
- ✅ `TestSubmitTask_ChangesStatusToInReview` - status updated from 'todo' to 'in_review'
- ✅ `TestSubmitTask_CreatesSubmissionRecord` - task_submissions entry created
- ✅ `TestSubmitTask_RequiresSubmissionText` - Empty submission text allowed (might be media-only)
- ✅ `TestSubmitTask_CannotResubmit` - Submitting 'completed' task fails

**Task Approval**
- ✅ `TestApproveTask_MarksCompleted` - status = 'completed', completed_at set
- ✅ `TestApproveTask_AwardsPoints` - connection.worker_points increased
- ✅ `TestApproveTask_OnlyManagerCanApprove` - Worker calling endpoint fails authorization
- ✅ `TestApproveTask_RequiresInReviewStatus` - Approving 'todo' task fails

**Task Disapproval**
- ✅ `TestDisapproveTask_MarksFailed` - status = 'failed', completed_at set
- ✅ `TestDisapproveTask_NoPointsAwarded` - worker_points unchanged
- ✅ `TestDisapproveTask_OnlyManagerCanDisapprove` - Worker calling endpoint fails
- ✅ `TestDisapproveTask_AssignsPunishmentIfSet` - punishment_task_id creates new assigned_task

**Reward Purchase**
- ✅ `TestPurchaseReward_DeductsPoints` - worker_points reduced by point_cost
- ✅ `TestPurchaseReward_CreatesAssignedTask` - reward task assigned to worker
- ✅ `TestPurchaseReward_InsufficientPoints` - Purchase fails if points < cost
- ✅ `TestPurchaseReward_OnlyWorkerCanPurchase` - Manager calling endpoint fails

**Points Calculation**
- ✅ `TestCalculateDurationMinutes_OnlyDays` - 2 days = 2880 minutes
- ✅ `TestCalculateDurationMinutes_OnlyHours` - 3 hours = 180 minutes
- ✅ `TestCalculateDurationMinutes_OnlyMinutes` - 45 minutes = 45
- ✅ `TestCalculateDurationMinutes_Combined` - 1d 2h 30m = 1590 minutes
- ✅ `TestCalculateDurationMinutes_AllNull` - Returns 1440 (default 24 hours)

---

## Priority 2: Integration Tests (Handlers + Database)

These test HTTP handlers with a real SQLite database in a transaction (rollback after each test).

### Setup Pattern
```go
func setupTestDB(t *testing.T) (*sql.DB, func()) {
    db, err := sql.Open("sqlite3", ":memory:")
    require.NoError(t, err)
    
    // Load schema
    schema, err := os.ReadFile("../../database/schema.sql")
    require.NoError(t, err)
    _, err = db.Exec(string(schema))
    require.NoError(t, err)
    
    cleanup := func() { db.Close() }
    return db, cleanup
}

func setupTestUser(t *testing.T, db *sql.DB) (userID int, token string) {
    // Create user, generate token
    // Return both for use in tests
}
```

### Authentication Integration Tests (`internal/pages/handlers_test.go`)

- ✅ `TestRegisterHandler_Success` - POST /register with valid data creates user
- ✅ `TestRegisterHandler_DuplicateEmail` - Returns 400 for existing email
- ✅ `TestRegisterHandler_DuplicateUsername` - Returns 400 for existing username
- ✅ `TestRegisterHandler_WeakPassword` - Password < 8 chars fails
- ✅ `TestRegisterHandler_SetsCookie` - Response includes Bearer cookie
- ✅ `TestLoginHandler_ValidCredentials` - Returns 200 and sets cookie
- ✅ `TestLoginHandler_WrongPassword` - Returns 401
- ✅ `TestLoginHandler_NonexistentUser` - Returns 401
- ✅ `TestLogout_ClearsCookie` - Cookie expires immediately

### Connection Integration Tests (`internal/connections/handlers_test.go`)

- ✅ `TestHandleCreateConnectionRequest_Success` - POST with valid email creates request
- ✅ `TestHandleAcceptConnectionRequest_Success` - POST accepts and creates connection
- ✅ `TestHandleRejectConnectionRequest_Success` - POST rejects without creating connection
- ✅ `TestHandleUpdateActiveConnection_Success` - PUT updates active_connection_id
- ✅ `TestHandleDeleteConnection_Success` - DELETE removes connection
- ✅ `TestServeConnectionsContent_ShowsPendingRequests` - Response includes pending requests
- ✅ `TestServeConnectionsContent_ShowsActiveConnection` - Response includes active connection details

### Task Integration Tests (`internal/buckets/handlers_test.go`)

- ✅ `TestHandleCreateTask_Success` - POST creates task template
- ✅ `TestHandleAssignTask_Success` - POST /task/assign/{id} creates assigned_task
- ✅ `TestHandleSubmitTask_Success` - POST /task/submit/{id} changes status
- ✅ `TestHandleApproveTask_Success` - POST /task/approve/{id} completes task
- ✅ `TestHandleDisapproveTask_Success` - POST /task/disapprove/{id} fails task
- ✅ `TestHandlePurchaseReward_Success` - POST purchases reward with points
- ✅ `TestHandleDeleteTask_Success` - DELETE removes task template
- ✅ `TestHandleDeleteAssignedTask_Success` - DELETE removes assigned task

### Middleware Tests (`internal/auth/middleware_test.go`)

- ✅ `TestAuthMiddleware_ValidToken` - Sets userID in context, calls next handler
- ✅ `TestAuthMiddleware_InvalidToken` - Redirects to /loginPage
- ✅ `TestAuthMiddleware_NoCookie` - Redirects to /loginPage
- ✅ `TestOptionalAuthMiddleware_NoToken` - Sets authStatus="guest", calls next handler

---

## Priority 3: Critical End-to-End Journeys

These are **smoke tests** to run before deployment. Test complete user flows in a real browser.

### E2E Test Suite (`tests/e2e/`)

#### `auth_journey_test.go`
1. **Registration Flow**
   - Navigate to /registerPage
   - Fill form (email, username, password)
   - Submit
   - Assert: Redirected to / (home/buckets page)
   - Assert: No error messages visible

2. **Login Flow**
   - Navigate to /loginPage
   - Fill credentials
   - Submit
   - Assert: Redirected to /
   - Assert: Navbar shows username (not "Login" button)

3. **Logout Flow**
   - Click logout link
   - Assert: Redirected to /description
   - Assert: Navbar shows "Login" button

#### `connection_journey_test.go`
1. **Send Connection Request**
   - User A logs in
   - Navigate to connections page
   - Fill email for User B, select "manager" role
   - Submit request
   - Assert: Success message appears
   - Logout

2. **Accept Connection Request**
   - User B logs in
   - Navigate to connections page
   - Assert: Pending request from User A visible
   - Click "Accept"
   - Assert: Connection appears in "Active Connections" list
   - Assert: Pending request removed

3. **Switch Active Connection**
   - User with multiple connections
   - Click different connection to activate
   - Assert: Active connection badge updates
   - Assert: Navbar shows correct partner name/role

4. **Delete Connection**
   - Click delete button on a connection
   - Confirm deletion
   - Assert: Connection removed from list
   - Assert: If was active connection, active badge disappears

#### `task_lifecycle_test.go`
1. **Manager Creates and Assigns Task**
   - Manager logs in with active connection
   - Click "Create Task"
   - Fill task form (title, points, timer)
   - Submit
   - Assert: Task appears in saved tasks
   - Click "Assign to Worker"
   - Assert: Task moves to assigned section

2. **Worker Submits Task**
   - Worker logs in
   - See assigned task in todo list
   - Click task to open
   - Fill submission text
   - Click "Submit for Review"
   - Assert: Task status changes to "In Review"

3. **Manager Approves Task**
   - Manager logs in
   - Navigate to "In Review" section
   - Open submitted task
   - Click "Approve"
   - Assert: Task moves to "Completed"
   - Assert: Worker points increase in navbar

4. **Worker Purchases Reward**
   - Worker navigates to /rewards
   - Assert: Rewards list visible
   - Click "Purchase" on affordable reward
   - Assert: Points deducted in navbar
   - Assert: Reward task appears in assigned tasks

#### `membership_journey_test.go`
1. **Access Member-Only Feature Without Connection**
   - Registered user without active connection
   - Navigate to / (buckets page)
   - Assert: Prompt to create connection visible
   - Assert: Task creation disabled/hidden

2. **Stripe Checkout Flow** (if implementing)
   - Navigate to /stripe
   - Click subscribe button
   - Assert: Stripe checkout modal appears
   - (Stop here - don't test Stripe's form)

---

## Priority 4: Database Query Tests

Test SQLC-generated queries directly to ensure schema integrity.

### `database/queries_test.go`

**User Queries**
- ✅ `TestCreateUser_Success` - Returns user_id
- ✅ `TestGetUserByEmail_Exists` - Finds user
- ✅ `TestGetUserByEmail_NotExists` - Returns error
- ✅ `TestGetUsernameByUserId_Success` - Returns correct username

**Connection Queries**
- ✅ `TestGetConnectionsById_ReturnsAllConnectionsForUser` - User in multiple connections
- ✅ `TestGetActiveConnectionId_NullWhenNotSet` - Returns sql.NullInt64{Valid: false}
- ✅ `TestGetActiveConnectionUserDetails_ReturnsPartner` - Not the requesting user

**Task Queries**
- ✅ `TestGetAssignedTasksByConnectionAndStatus_FiltersByStatus` - Only 'todo' returned
- ✅ `TestGetAssignedTasksByConnectionAndStatus_OrdersByDueTime` - Soonest first
- ✅ `TestGetExpiredTodoTasks_ReturnsOverdueTasks` - due_time < now

**Points Queries**
- ✅ `TestAddWorkerPoints_IncreasesPoints` - worker_points += amount
- ✅ `TestDeductWorkerPoints_DecreasesPoints` - worker_points -= amount
- ✅ `TestDeductWorkerPoints_CannotGoNegative` - Fails or stops at 0 (add constraint if needed)

---

## Non-Priority: What We're NOT Testing (Yet)

These are intentionally excluded to avoid brittle tests that break with UI changes:

### ❌ Template Rendering
- Don't test specific HTML structure
- Don't test CSS classes or styling
- Don't test exact wording of labels

**Why:** You said UI will change frequently. Testing "button has class 'btn-primary'" breaks every time you tweak Tailwind.

**Alternative:** E2E tests verify buttons *work*, not what they look like.

### ❌ Stripe Payment Processing
- Don't test Stripe's form validation
- Don't test actual payment capture

**Why:** Stripe owns this. Test your code's response to Stripe webhooks/callbacks instead.

**Future:** Add webhook handler tests when implemented.

### ❌ File Upload/Storage
- Don't test S3 upload logic until finalized
- Don't test file type validation until requirements locked

**Why:** `storage.ServeUploadedFile` exists but implementation unclear. Wait for stability.

### ❌ Repeating Tasks
- Don't test cron job logic until scheduler implemented

**Why:** `GetAllDueRepeatingTasks` query exists but no handler calls it yet.

---

## Test Organization Structure

```
pact/
├── internal/
│   ├── auth/
│   │   ├── jwt_test.go
│   │   ├── service_test.go
│   │   └── middleware_test.go
│   ├── connections/
│   │   ├── services_test.go
│   │   └── handlers_test.go
│   ├── buckets/
│   │   ├── services_test.go
│   │   └── handlers_test.go
│   └── pages/
│       └── handlers_test.go
├── database/
│   └── queries_test.go
├── tests/
│   ├── e2e/
│   │   ├── auth_journey_test.go
│   │   ├── connection_journey_test.go
│   │   ├── task_lifecycle_test.go
│   │   └── membership_journey_test.go
│   └── helpers/
│       ├── test_db.go       # Shared DB setup
│       └── test_users.go    # User creation helpers
└── TESTING_PLAN.md (this file)
```

---

## Implementation Order

### Week 1: Foundation
1. Add `testify` dependency
2. Create `tests/helpers/` with shared setup functions
3. Write `internal/auth/jwt_test.go` (easiest, high value)
4. Write `internal/auth/service_test.go`

### Week 2: Core Business Logic
1. Write `internal/connections/services_test.go`
2. Write `internal/buckets/services_test.go` (points, duration)
3. Write `database/queries_test.go`

### Week 3: HTTP Integration
1. Write `internal/auth/middleware_test.go`
2. Write `internal/pages/handlers_test.go` (auth flows)
3. Write `internal/connections/handlers_test.go`
4. Write `internal/buckets/handlers_test.go`

### Week 4: E2E & Polish
1. Set up Playwright/Selenium
2. Write `auth_journey_test.go`
3. Write `connection_journey_test.go`
4. Write `task_lifecycle_test.go`
5. Add test coverage reporting

---

## Running Tests

### Unit & Integration Tests
```bash
# All tests
go test ./...

# Specific package
go test ./internal/auth

# With coverage
go test -cover ./...

# Verbose output
go test -v ./internal/connections

# Run single test
go test -run TestCreateConnectionRequest_ValidManager ./internal/connections
```

### E2E Tests
```bash
# Run all E2E
go test ./tests/e2e/...

# Run specific journey
go test -run TestAuthJourney ./tests/e2e

# Headless mode (CI)
E2E_HEADLESS=true go test ./tests/e2e/...
```

### CI Integration (Future)
```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - run: go test -v -cover ./...
      - run: go test -v ./tests/e2e/...
        env:
          E2E_HEADLESS: true
```

---

## Success Metrics

### Code Coverage Targets
- **Auth package:** 90%+ (critical security)
- **Connections package:** 85%+ (complex logic)
- **Buckets package:** 80%+ (many edge cases)
- **Overall:** 70%+ (excludes templates/static)

### E2E Stability
- All E2E tests pass 3 runs in a row before considering stable
- Max 5 minutes total E2E execution time

### Regression Prevention
- Zero regressions on core flows (auth, connections, task lifecycle)
- UI changes don't break backend tests

---

## Test Data Strategy

### User Personas for Testing
```go
// tests/helpers/test_users.go
type TestUser struct {
    Email    string
    Username string
    Password string
    Role     string // "manager", "worker", "both"
}

var (
    ManagerUser = TestUser{
        Email:    "manager@test.com",
        Username: "test_manager",
        Password: "TestPass123!",
        Role:     "manager",
    }
    WorkerUser = TestUser{
        Email:    "worker@test.com",
        Username: "test_worker",
        Password: "TestPass123!",
        Role:     "worker",
    }
    DualUser = TestUser{
        Email:    "dual@test.com",
        Username: "test_dual",
        Password: "TestPass123!",
        Role:     "both",
    }
)
```

### Fixture Data
- **Tasks:** 1 normal, 1 punishment, 1 reward per test
- **Connections:** Pre-create 1-2 established connections for complex scenarios
- **Points:** Set to known values (e.g., 100) for predictable math

---

## Maintenance Guidelines

### When to Update Tests
1. **Always update:** When core business logic changes (e.g., points formula)
2. **Sometimes update:** When API contracts change (new required fields)
3. **Never update:** When only UI styling changes

### Red Flags
- Test names like `TestButtonHasCorrectClass` ❌
- Tests that parse HTML to check element order ❌
- Tests that depend on exact error message wording ❌

### Green Flags
- Test names like `TestApproveTask_AwardsPoints` ✅
- Tests that verify status codes and data presence ✅
- Tests that check business rules (e.g., "can't delete other user's task") ✅

---

## Open Questions & Future Decisions

1. **Stripe Webhooks:** When subscription expires, how is `is_member` set to 0? Need handler test.
2. **File Upload Validation:** What file types allowed? Size limits? Test after requirements finalized.
3. **Repeating Tasks:** When is the cron job triggered? Test scheduler separately.
4. **Punishment Auto-Assignment:** Does disapproving a task auto-assign punishment? Need business rule confirmation.
5. **Points Going Negative:** Should database constraint prevent negative points? Add test + migration if yes.

---

## Appendix: Example Test Snippets

### Example Unit Test
```go
// internal/auth/jwt_test.go
func TestValidateToken_ExpiredToken(t *testing.T) {
    // Create token with past expiry
    claims := jwt.MapClaims{
        "user_id": 123,
        "exp":     time.Now().Add(-1 * time.Hour).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
    
    // Validate should fail
    _, err := ValidateToken(tokenString)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "expired")
}
```

### Example Integration Test
```go
// internal/connections/handlers_test.go
func TestHandleAcceptConnectionRequest_Success(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()
    
    // Create two users and a pending request
    managerID := createTestUser(t, db, "manager@test.com")
    workerID := createTestUser(t, db, "worker@test.com")
    requestID := createTestRequest(t, db, workerID, managerID)
    
    // Build HTTP request
    req := httptest.NewRequest("POST", "/acceptConnectionRequest/"+strconv.Itoa(requestID), nil)
    req = req.WithContext(context.WithValue(req.Context(), "userID", managerID))
    w := httptest.NewRecorder()
    
    // Execute handler
    HandleAcceptConnectionRequest(w, req)
    
    // Assert response
    assert.Equal(t, http.StatusOK, w.Code)
    
    // Assert connection created
    var count int
    db.QueryRow("SELECT COUNT(*) FROM connections WHERE manager_id=? AND worker_id=?", 
        managerID, workerID).Scan(&count)
    assert.Equal(t, 1, count)
    
    // Assert request deactivated
    var isActive int
    db.QueryRow("SELECT is_active FROM connection_requests WHERE request_id=?", 
        requestID).Scan(&isActive)
    assert.Equal(t, 0, isActive)
}
```

### Example E2E Test (Playwright)
```go
// tests/e2e/auth_journey_test.go
func TestRegistrationFlow(t *testing.T) {
    browser := setupBrowser(t)
    defer browser.Close()
    
    page := browser.NewPage()
    page.Goto("http://localhost:8080/registerPage")
    
    // Fill form
    page.Fill("input[name=email]", "newuser@test.com")
    page.Fill("input[name=username]", "newuser")
    page.Fill("input[name=password]", "SecurePass123!")
    
    // Submit
    page.Click("button[type=submit]")
    
    // Assert redirect to home
    page.WaitForURL("http://localhost:8080/")
    
    // Assert no error messages
    errorMsg := page.Locator(".error-message")
    assert.False(t, errorMsg.IsVisible())
}
```

---

**End of Testing Plan**

This plan prioritizes stability and maintainability. Start with unit tests for business logic, add integration tests for API contracts, and finish with minimal E2E tests for critical journeys. As features stabilize post-launch, expand coverage strategically.
