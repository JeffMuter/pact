# How Code Push is Checked - Complete Visual Guide

## File Location

**GitHub Actions Workflow File:**
```
.github/workflows/test.yml  ← This file controls CI/CD
```

This file is automatically executed by GitHub when you push code.

---

## What Happens When You Push

### Step-by-Step Process

```
1. You push code to GitHub
   └─→ git push origin master
   
2. GitHub detects the push
   └─→ Looks at .github/workflows/test.yml
   
3. GitHub Actions starts (automatically)
   └─→ Creates a virtual machine (ubuntu-latest)
   └─→ Runs the "test" job
   
4. Inside the virtual machine:
   
   Step 1: Checkout code
   ├─ Downloads your code
   
   Step 2: Set up Go
   ├─ Installs Go 1.23
   ├─ Sets up dependency cache
   
   Step 3: Download dependencies
   ├─ Runs: go mod download && go mod verify
   └─ Ensures all Go dependencies are valid
   
   Step 4: RUN TESTS ⭐ (This is the critical step)
   ├─ Runs: go test -v -race ./...
   ├─ Finds ALL *_test.go files
   ├─ Executes every test function
   ├─ -v = verbose output (show each test)
   ├─ -race = detect race conditions
   └─ ./... = test all packages
   
   Step 5: Generate coverage report
   ├─ Runs: go test -coverprofile=coverage.out ...
   └─ Creates a coverage file
   
   Step 6: Upload to Codecov
   ├─ Uploads coverage.out to codecov.io
   └─ Shows coverage trend over time
   
5. GitHub shows result on your PR/push
   ├─ ✅ Green checkmark = all tests passed
   └─ ❌ Red X = tests failed
```

---

## Where the Tests Are

All test files follow the naming pattern: `*_test.go`

```
pact/
├── internal/
│   ├── auth/
│   │   ├── jwt.go              ← Source code
│   │   ├── jwt_test.go         ← TESTS for jwt.go
│   │   ├── middleware.go       ← Source code
│   │   ├── middleware_test.go  ← TESTS for middleware.go
│   │   ├── service.go          ← Source code
│   │   └── service_test.go     ← TESTS for service.go
│   └── buckets/
│       ├── services.go         ← Source code
│       └── services_test.go    ← TESTS for services.go
└── tests/
    └── helpers/
        ├── test_db.go          ← Helper functions
        ├── test_users.go       ← Helper functions
        └── helpers_test.go     ← TESTS for helpers
```

**Current test files (5 total):**
1. `internal/auth/jwt_test.go` - Tests JWT token generation/validation
2. `internal/auth/service_test.go` - Tests password hashing
3. `internal/auth/middleware_test.go` - Tests auth middleware
4. `internal/buckets/services_test.go` - Tests duration calculations
5. `tests/helpers/helpers_test.go` - Tests helper functions

---

## What Kind of Tests?

### Unit Tests (what we have)

These test individual functions in isolation:

```go
// Example: Testing JWT token generation
func TestGenerateToken_ValidUserID(t *testing.T) {
    // Generate a token
    token, err := GenerateToken(123)
    
    // Verify it works
    if err != nil {
        t.Fail()
    }
    if token == "" {
        t.Fail()
    }
}
```

**What gets tested:**
- ✅ JWT generation (6 tests)
- ✅ Password hashing (4 tests)
- ✅ Authentication middleware (5 tests)
- ✅ Duration/points calculations (5 tests)
- ✅ Test helper functions (6 tests)

**Total: 26 tests**

### How Tests Run on GitHub

When you push:

```bash
# This command runs on GitHub Actions
go test -v -race ./...

# Output looks like:
=== RUN   TestGenerateToken_ValidUserID
--- PASS: TestGenerateToken_ValidUserID (0.00s)
=== RUN   TestGenerateToken_ContainsUserID
--- PASS: TestGenerateToken_ContainsUserID (0.00s)
=== RUN   TestValidateToken_ValidToken
--- PASS: TestValidateToken_ValidToken (0.00s)
...
PASS
ok  	pact/internal/auth	0.701s
```

---

## The GitHub Actions Workflow File (Explained)

Here's `.github/workflows/test.yml` with annotations:

```yaml
name: Tests                          # Name of the workflow

on:                                  # When to run
  push:
    branches: [master, develop]      # Run on push to master or develop
  pull_request:
    branches: [master]               # Run on PRs to master

jobs:                                # What to run
  test:                              # Job name: "test"
    runs-on: ubuntu-latest           # Run on Linux
    strategy:
      matrix:
        go-version: ['1.23']         # Test with Go 1.23
    
    steps:                           # Steps to execute
      - name: Checkout code
        uses: actions/checkout@v4    # Download your code from GitHub
      
      - name: Set up Go
        uses: actions/setup-go@v4    # Install Go 1.23
        with:
          go-version: ${{ matrix.go-version }}
          cache: 'go'                # Cache dependencies for speed
      
      - name: Download dependencies
        run: go mod download && go mod verify  # Download all Go packages
      
      - name: Run tests              # ⭐ THE KEY STEP
        run: go test -v -race ./...  # Run ALL *_test.go files
      
      - name: Generate coverage report
        run: go test -coverprofile=coverage.out -covermode=atomic ./...
      
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          fail_ci_if_error: false    # Don't fail if Codecov is down
```

---

## What Happens If Tests Fail?

### Scenario: You push broken code

```
1. You push code with a failing test
   └─→ git push origin master

2. GitHub Actions runs
   └─→ Detects 1 test failure
   └─→ Step "Run tests" FAILS
   
3. GitHub marks the push as FAILED
   └─→ Shows ❌ red X on your PR
   └─→ Can add branch protection to prevent merge

4. You see output like:
   
   === RUN   TestSomething
   --- FAIL: TestSomething (0.01s)
       some_test.go:45: Expected X but got Y
   FAIL
   ok	pact/internal/auth	0.701s
```

---

## Visual: Pull Request Status

When you create a PR, GitHub shows this:

```
┌─────────────────────────────────────────────────────────┐
│ Your PR                                                 │
├─────────────────────────────────────────────────────────┤
│ Changes:  +150 lines, -20 lines                         │
│                                                         │
│ Checks:                                                 │
│ ✅ Tests (1/1 passed)                                   │
│    All tests passed in 32 seconds                       │
│                                                         │
│ ✅ Code review (pending)                               │
│                                                         │
│ [Merge Pull Request] ← Can click once tests pass       │
└─────────────────────────────────────────────────────────┘
```

Or if tests fail:

```
┌─────────────────────────────────────────────────────────┐
│ Your PR                                                 │
├─────────────────────────────────────────────────────────┤
│ Changes:  +150 lines, -20 lines                         │
│                                                         │
│ Checks:                                                 │
│ ❌ Tests (0/1 passed)                                   │
│    1 test failed in 32 seconds                          │
│    See details below ↓                                  │
│                                                         │
│ [Merge is blocked]                                      │
└─────────────────────────────────────────────────────────┘
```

---

## How to View Test Results on GitHub

### Option 1: In the PR

1. Go to your PR on GitHub
2. Scroll to "Checks" section
3. Click "Details" next to "Tests"
4. See full test output

### Option 2: In Actions Tab

1. Go to your repository
2. Click "Actions" tab at top
3. Click the workflow run
4. See all test results

### Option 3: Locally (Faster)

```bash
# Run tests on your machine before pushing
make test                 # Runs all tests
make test-verbose         # Shows each test
make test-coverage        # Generates HTML report
```

---

## Test Discovery

Go automatically finds all test files using this pattern:

```
Filename:     *_test.go      (must end with _test.go)
Package:      Same as source (jwt_test.go in auth/ package)
Function:     Test*          (must start with Test)
Parameter:    *testing.T     (receives test object)
```

**Example:**

```go
// File: internal/auth/jwt_test.go
package auth

func TestGenerateToken_ValidUserID(t *testing.T) {
    // Go automatically finds and runs this function
}

func TestGenerateToken_ContainsUserID(t *testing.T) {
    // And this one
}
```

When you run `go test ./...`:
1. Go finds all `*_test.go` files
2. Finds all `Test*` functions in them
3. Runs each one
4. Reports pass/fail for each

---

## Summary: The Complete Flow

```
You commit code
     ↓
git push origin master
     ↓
GitHub receives push
     ↓
GitHub reads .github/workflows/test.yml
     ↓
GitHub Actions starts job
     ↓
Virtual machine:
  1. Checks out your code
  2. Installs Go 1.23
  3. Downloads dependencies
  4. Runs: go test -v -race ./...     ← Finds all *_test.go files
  5. Executes all Test* functions
  6. Generates coverage report
     ↓
GitHub shows result:
  ✅ All tests passed → Can merge
  ❌ Tests failed → Can't merge (with branch protection)
```

---

## Key Points

- **File:** `.github/workflows/test.yml` controls everything
- **Tests:** All `*_test.go` files (5 files, 26 tests total)
- **Command:** `go test -v -race ./...` (automatic discovery)
- **When:** Every push + every PR
- **Speed:** ~30 seconds (includes setup)
- **Result:** ✅ or ❌ shown on GitHub

---

## Related Files

For more context:
- `Makefile` - Run tests locally (`make test`)
- `TESTING.md` - Complete testing guide
- `TESTING_PLAN.md` - What tests to write next
- `scripts/pre-commit` - Tests run BEFORE you even push (locally)

