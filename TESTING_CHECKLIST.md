# Pact Testing Implementation Checklist

**Companion to:** `TESTING_PLAN.md`  
**Purpose:** Actionable, ordered checklist with starter code templates

---

## Phase 0: Setup (Day 1)

### Dependencies
- [ ] `go get github.com/stretchr/testify/assert`
- [ ] `go get github.com/stretchr/testify/require`
- [ ] Update `AGENTS.md` with test command: `go test ./...`
- [ ] Create `.github/workflows/test.yml` (optional, for CI)

### Shared Test Utilities
- [ ] Create `tests/helpers/test_db.go`
  ```go
  package helpers
  
  import (
      "database/sql"
      "os"
      "testing"
      _ "github.com/mattn/go-sqlite3"
      "github.com/stretchr/testify/require"
  )
  
  func SetupTestDB(t *testing.T) (*sql.DB, func()) {
      db, err := sql.Open("sqlite3", ":memory:")
      require.NoError(t, err)
      
      schema, err := os.ReadFile("../../database/schema.sql")
      require.NoError(t, err)
      
      _, err = db.Exec(string(schema))
      require.NoError(t, err)
      
      cleanup := func() { db.Close() }
      return db, cleanup
  }
  ```

- [ ] Create `tests/helpers/test_users.go`
  ```go
  package helpers
  
  import (
      "context"
      "testing"
      "pact/database"
      "pact/internal/auth"
      "github.com/stretchr/testify/require"
  )
  
  type TestUser struct {
      UserID   int
      Email    string
      Username string
      Password string
      Token    string
  }
  
  func CreateTestUser(t *testing.T, db *sql.DB, email, username, password string) TestUser {
      queries := database.New(db)
      ctx := context.Background()
      
      hashedPass, err := auth.HashPassword(password)
      require.NoError(t, err)
      
      userID, err := queries.CreateUser(ctx, database.CreateUserParams{
          Email:        email,
          Username:     username,
          PasswordHash: hashedPass,
      })
      require.NoError(t, err)
      
      token, err := auth.GenerateToken(uint(userID))
      require.NoError(t, err)
      
      return TestUser{
          UserID:   int(userID),
          Email:    email,
          Username: username,
          Password: password,
          Token:    token,
      }
  }
  ```

---

## Phase 1: Auth Package (Days 2-3)

### `internal/auth/jwt_test.go`
- [ ] `TestGenerateToken_ValidUserID`
- [ ] `TestGenerateToken_ContainsUserID` (decode token, check claims)
- [ ] `TestValidateToken_ValidToken`
- [ ] `TestValidateToken_ExpiredToken` (manually create expired token)
- [ ] `TestValidateToken_MalformedToken` (pass garbage string)
- [ ] `TestValidateToken_WrongSecret` (sign with different key)

**Starter Template:**
```go
package auth

import (
    "os"
    "testing"
    "time"
    "github.com/golang-jwt/jwt/v5"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestGenerateToken_ValidUserID(t *testing.T) {
    os.Setenv("JWT_SECRET_KEY", "test-secret-key")
    
    token, err := GenerateToken(123)
    
    require.NoError(t, err)
    assert.NotEmpty(t, token)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
    os.Setenv("JWT_SECRET_KEY", "test-secret-key")
    
    claims := jwt.MapClaims{
        "user_id": float64(123),
        "exp":     time.Now().Add(-1 * time.Hour).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString([]byte("test-secret-key"))
    
    _, err := ValidateToken(tokenString)
    assert.Error(t, err)
}
```

### `internal/auth/service_test.go`
- [ ] `TestHashPassword_NotPlaintext`
- [ ] `TestCheckPasswordHash_CorrectPassword`
- [ ] `TestCheckPasswordHash_WrongPassword`
- [ ] `TestCheckPasswordHash_EmptyPassword`

**Starter Template:**
```go
package auth

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestHashPassword_NotPlaintext(t *testing.T) {
    password := "SecurePass123!"
    
    hash, err := HashPassword(password)
    
    require.NoError(t, err)
    assert.NotEqual(t, password, hash)
    assert.Contains(t, hash, "$2a$") // Bcrypt prefix
}

func TestCheckPasswordHash_CorrectPassword(t *testing.T) {
    password := "SecurePass123!"
    hash, _ := HashPassword(password)
    
    err := CheckPasswordHash(password, hash)
    
    assert.NoError(t, err)
}
```

### `internal/auth/middleware_test.go`
- [ ] `TestAuthMiddleware_ValidToken`
- [ ] `TestAuthMiddleware_InvalidToken`
- [ ] `TestAuthMiddleware_NoCookie`
- [ ] `TestOptionalAuthMiddleware_NoToken`

**Starter Template:**
```go
package auth

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_ValidToken(t *testing.T) {
    os.Setenv("JWT_SECRET_KEY", "test-secret-key")
    token, _ := GenerateToken(123)
    
    nextCalled := false
    next := func(w http.ResponseWriter, r *http.Request) {
        nextCalled = true
        userID := r.Context().Value("userID").(int)
        assert.Equal(t, 123, userID)
    }
    
    req := httptest.NewRequest("GET", "/", nil)
    req.AddCookie(&http.Cookie{Name: "Bearer", Value: token})
    w := httptest.NewRecorder()
    
    handler := AuthMiddleware(next)
    handler(w, req)
    
    assert.True(t, nextCalled)
}
```

---

## Phase 2: Connections Package (Days 4-5)

### `internal/connections/services_test.go`

#### Connection Request Creation (3 tests)
- [ ] `TestCreateConnectionRequest_ValidManager`
- [ ] `TestCreateConnectionRequest_InvalidRole`
- [ ] `TestCreateConnectionRequest_NonexistentEmail`

**Starter Template:**
```go
package connections

import (
    "context"
    "testing"
    "pact/tests/helpers"
    "pact/database"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCreateConnectionRequest_ValidManager(t *testing.T) {
    db, cleanup := helpers.SetupTestDB(t)
    defer cleanup()
    
    sender := helpers.CreateTestUser(t, db, "sender@test.com", "sender", "pass123")
    receiver := helpers.CreateTestUser(t, db, "receiver@test.com", "receiver", "pass123")
    
    ctx := context.Background()
    err := CreateConnectionRequest(ctx, sender.UserID, "manager", receiver.Email)
    
    require.NoError(t, err)
    
    // Verify request created
    queries := database.New(db)
    requests, err := queries.GetUserPendingRequests(ctx, int64(receiver.UserID))
    require.NoError(t, err)
    assert.Len(t, requests, 1)
    assert.Equal(t, int64(sender.UserID), requests[0].SuggestedManagerID)
}
```

#### Connection Request Acceptance (4 tests)
- [ ] `TestAcceptConnectionRequest_CreatesConnection`
- [ ] `TestAcceptConnectionRequest_CorrectRoles`
- [ ] `TestAcceptConnectionRequest_DeactivatesRequest`
- [ ] `TestAcceptConnectionRequest_NonexistentRequest`

#### Connection Request Rejection (2 tests)
- [ ] `TestRejectConnectionRequest_DeactivatesRequest`
- [ ] `TestRejectConnectionRequest_NoConnectionCreated`

#### Active Connection (3 tests)
- [ ] `TestUpdateActiveConnection_SetsActiveConnection`
- [ ] `TestGetActiveConnectionDetails_ReturnsCorrectPartner`
- [ ] `TestGetActiveConnectionDetails_NoActiveConnection`

#### Connection Deletion (3 tests)
- [ ] `TestDeleteConnection_RemovesConnection`
- [ ] `TestDeleteConnection_ClearsActiveIfMatch`
- [ ] `TestDeleteConnection_UnauthorizedUser`

---

## Phase 3: Buckets Package (Days 6-7)

### `internal/buckets/services_test.go`

#### Helper Functions (test first - easy wins)
- [ ] `TestCalculateDurationMinutes_OnlyDays`
- [ ] `TestCalculateDurationMinutes_Combined`
- [ ] `TestCalculateDurationMinutes_AllNull`

**Starter Template:**
```go
package buckets

import (
    "database/sql"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestCalculateDurationMinutes_OnlyDays(t *testing.T) {
    days := sql.NullInt64{Int64: 2, Valid: true}
    hours := sql.NullInt64{}
    minutes := sql.NullInt64{}
    
    result := calculateDurationMinutes(days, hours, minutes)
    
    assert.Equal(t, int64(2880), result) // 2 days * 24 hours * 60 minutes
}

func TestCalculateDurationMinutes_AllNull(t *testing.T) {
    result := calculateDurationMinutes(sql.NullInt64{}, sql.NullInt64{}, sql.NullInt64{})
    
    assert.Equal(t, int64(1440), result) // Default 24 hours
}
```

#### Task Assignment (4 tests)
- [ ] `TestAssignTask_CreatesAssignedTask` (use handler or service function)
- [ ] `TestAssignTask_SetsCorrectDueTime`
- [ ] `TestAssignTask_CopiesTemplateFields`
- [ ] `TestAssignTask_DefaultsToTodoStatus`

#### Task Lifecycle (8 tests)
- [ ] `TestSubmitTask_ChangesStatusToInReview`
- [ ] `TestApproveTask_MarksCompleted`
- [ ] `TestApproveTask_AwardsPoints`
- [ ] `TestApproveTask_OnlyManagerCanApprove`
- [ ] `TestDisapproveTask_MarksFailed`
- [ ] `TestDisapproveTask_NoPointsAwarded`
- [ ] `TestPurchaseReward_DeductsPoints`
- [ ] `TestPurchaseReward_InsufficientPoints`

---

## Phase 4: Database Queries (Day 8)

### `database/queries_test.go`

#### User Queries (3 tests)
- [ ] `TestCreateUser_Success`
- [ ] `TestGetUserByEmail_Exists`
- [ ] `TestGetUserByEmail_NotExists`

**Starter Template:**
```go
package database

import (
    "context"
    "testing"
    "pact/tests/helpers"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCreateUser_Success(t *testing.T) {
    db, cleanup := helpers.SetupTestDB(t)
    defer cleanup()
    
    queries := New(db)
    ctx := context.Background()
    
    userID, err := queries.CreateUser(ctx, CreateUserParams{
        Email:        "test@test.com",
        Username:     "testuser",
        PasswordHash: "hashedpass",
    })
    
    require.NoError(t, err)
    assert.Greater(t, userID, int64(0))
}
```

#### Connection Queries (3 tests)
- [ ] `TestGetConnectionsById_ReturnsAllConnectionsForUser`
- [ ] `TestGetActiveConnectionId_NullWhenNotSet`
- [ ] `TestGetActiveConnectionUserDetails_ReturnsPartner`

#### Task Queries (2 tests)
- [ ] `TestGetAssignedTasksByConnectionAndStatus_FiltersByStatus`
- [ ] `TestGetAssignedTasksByConnectionAndStatus_OrdersByDueTime`

#### Points Queries (2 tests)
- [ ] `TestAddWorkerPoints_IncreasesPoints`
- [ ] `TestDeductWorkerPoints_DecreasesPoints`

---

## Phase 5: Handler Integration Tests (Days 9-10)

### `internal/pages/handlers_test.go`

#### Registration (5 tests)
- [ ] `TestRegisterHandler_Success`
- [ ] `TestRegisterHandler_DuplicateEmail`
- [ ] `TestRegisterHandler_DuplicateUsername`
- [ ] `TestRegisterHandler_SetsCookie`
- [ ] `TestLoginHandler_ValidCredentials`

**Starter Template:**
```go
package pages

import (
    "net/http"
    "net/http/httptest"
    "net/url"
    "strings"
    "testing"
    "pact/tests/helpers"
    "github.com/stretchr/testify/assert"
)

func TestRegisterHandler_Success(t *testing.T) {
    db, cleanup := helpers.SetupTestDB(t)
    defer cleanup()
    
    form := url.Values{}
    form.Add("email", "new@test.com")
    form.Add("username", "newuser")
    form.Add("password", "SecurePass123!")
    
    req := httptest.NewRequest("POST", "/register", strings.NewReader(form.Encode()))
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    w := httptest.NewRecorder()
    
    RegisterHandler(w, req)
    
    assert.Equal(t, http.StatusSeeOther, w.Code) // Redirect on success
    
    // Check cookie set
    cookies := w.Result().Cookies()
    assert.NotEmpty(t, cookies)
    assert.Equal(t, "Bearer", cookies[0].Name)
}
```

### `internal/connections/handlers_test.go`

#### Connection Handlers (6 tests)
- [ ] `TestHandleCreateConnectionRequest_Success`
- [ ] `TestHandleAcceptConnectionRequest_Success`
- [ ] `TestHandleRejectConnectionRequest_Success`
- [ ] `TestHandleUpdateActiveConnection_Success`
- [ ] `TestHandleDeleteConnection_Success`
- [ ] `TestServeConnectionsContent_ShowsPendingRequests`

### `internal/buckets/handlers_test.go`

#### Task Handlers (7 tests)
- [ ] `TestHandleCreateTask_Success`
- [ ] `TestHandleAssignTask_Success`
- [ ] `TestHandleSubmitTask_Success`
- [ ] `TestHandleApproveTask_Success`
- [ ] `TestHandleDisapproveTask_Success`
- [ ] `TestHandlePurchaseReward_Success`
- [ ] `TestHandleDeleteTask_Success`

---

## Phase 6: E2E Setup (Days 11-12)

### Choose E2E Framework
- [ ] **Option A:** Install Playwright (`npm install -D @playwright/test`)
- [ ] **Option B:** Install Selenium Go (`go get github.com/tebeka/selenium`)

### E2E Helpers
- [ ] Create `tests/e2e/setup_test.go`
  ```go
  package e2e
  
  import (
      "testing"
      "github.com/playwright-community/playwright-go"
  )
  
  func SetupBrowser(t *testing.T) playwright.Browser {
      pw, err := playwright.Run()
      if err != nil {
          t.Fatalf("could not start playwright: %v", err)
      }
      
      browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
          Headless: playwright.Bool(true),
      })
      if err != nil {
          t.Fatalf("could not launch browser: %v", err)
      }
      
      return browser
  }
  ```

---

## Phase 7: E2E Tests (Days 13-14)

### `tests/e2e/auth_journey_test.go`
- [ ] `TestRegistrationFlow`
- [ ] `TestLoginFlow`
- [ ] `TestLogoutFlow`

**Starter Template:**
```go
package e2e

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestRegistrationFlow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }
    
    browser := SetupBrowser(t)
    defer browser.Close()
    
    page, err := browser.NewPage()
    if err != nil {
        t.Fatalf("could not create page: %v", err)
    }
    
    // Navigate
    if err := page.Goto("http://localhost:8080/registerPage"); err != nil {
        t.Fatalf("could not goto: %v", err)
    }
    
    // Fill form
    page.Fill("input[name='email']", "e2e@test.com")
    page.Fill("input[name='username']", "e2euser")
    page.Fill("input[name='password']", "E2EPass123!")
    
    // Submit
    page.Click("button[type='submit']")
    
    // Assert redirect
    page.WaitForURL("http://localhost:8080/")
    assert.Contains(t, page.URL(), "localhost:8080")
}
```

### `tests/e2e/connection_journey_test.go`
- [ ] `TestSendConnectionRequest`
- [ ] `TestAcceptConnectionRequest`
- [ ] `TestSwitchActiveConnection`
- [ ] `TestDeleteConnection`

### `tests/e2e/task_lifecycle_test.go`
- [ ] `TestManagerCreatesAndAssignsTask`
- [ ] `TestWorkerSubmitsTask`
- [ ] `TestManagerApprovesTask`
- [ ] `TestWorkerPurchasesReward`

---

## Phase 8: CI/CD (Day 15)

### GitHub Actions
- [ ] Create `.github/workflows/test.yml`
  ```yaml
  name: Tests
  on: [push, pull_request]
  
  jobs:
    unit-tests:
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@v3
        - uses: actions/setup-go@v4
          with:
            go-version: '1.23.3'
        - name: Run unit tests
          run: go test -v -cover ./internal/... ./database/...
    
    e2e-tests:
      runs-on: ubuntu-latest
      needs: unit-tests
      steps:
        - uses: actions/checkout@v3
        - uses: actions/setup-go@v4
          with:
            go-version: '1.23.3'
        - name: Install Playwright
          run: npm install -D @playwright/test && npx playwright install
        - name: Start server
          run: ./buildAir.sh &
        - name: Wait for server
          run: sleep 5
        - name: Run E2E tests
          run: go test -v ./tests/e2e/...
  ```

### Coverage Reporting
- [ ] Add coverage badge to README
- [ ] Set up Codecov or Coveralls (optional)

---

## Quick Reference: Test Commands

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/auth

# Run single test
go test -run TestGenerateToken_ValidUserID ./internal/auth

# Skip E2E tests (fast feedback)
go test -short ./...

# Run only E2E tests
go test ./tests/e2e/...

# Verbose output
go test -v ./...

# Generate coverage HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

---

## Progress Tracking

### Week 1: Foundation (Days 1-3)
- [ ] Setup complete
- [ ] Auth tests complete (18 tests)

### Week 2: Business Logic (Days 4-8)
- [ ] Connections tests complete (15 tests)
- [ ] Buckets tests complete (15 tests)
- [ ] Database tests complete (10 tests)

### Week 3: Integration (Days 9-10)
- [ ] Handler tests complete (18 tests)

### Week 4: E2E & Deploy (Days 11-15)
- [ ] E2E setup complete
- [ ] E2E tests complete (10 tests)
- [ ] CI/CD configured

**Total Test Count Goal:** 86 tests

---

## When Stuck: Debugging Tests

### Common Issues

**Issue:** `panic: sql: database is closed`  
**Fix:** Don't call `defer cleanup()` before test finishes. Move to end of function.

**Issue:** `context deadline exceeded`  
**Fix:** Use `context.Background()` in tests, not request contexts.

**Issue:** E2E test hangs  
**Fix:** Add timeouts to page actions: `page.Click("button", playwright.PageClickOptions{Timeout: 5000})`

**Issue:** Cookie not set in handler test  
**Fix:** Check `w.Result().Cookies()`, not `w.Header().Get("Set-Cookie")` directly.

**Issue:** Test DB schema out of sync  
**Fix:** Ensure `database/schema.sql` matches current schema. Run `sqlc generate`.

---

**End of Checklist**

Start with Phase 0 and work sequentially. Each phase builds on the previous. Mark checkboxes as you complete tests. Good luck!
