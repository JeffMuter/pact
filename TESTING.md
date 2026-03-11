# Pact CI/CD & Testing Guide

## Quick Start

### First-Time Setup

```bash
# Clone the repository and cd into it
git clone <repo>
cd pact

# Run the setup script to install git hooks
bash scripts/setup.sh

# Enter nix shell and install dependencies
nix-shell
npm install
```

### Daily Development

```bash
# Terminal 1: Start dev server with hot reload
make dev

# Terminal 2: Run tests
make test

# Or run tests once and exit
make test
```

---

## Available Commands

### Testing

```bash
make test              # Run all tests (fast)
make test-verbose      # Run tests with detailed output
make test-coverage     # Generate HTML coverage report (coverage.html)
```

### Development

```bash
make dev               # Start dev server with hot reload (uses Air)
make build             # Build CSS and Go binary
make install           # Install npm dependencies
make clean             # Remove build artifacts and coverage files
```

### Code Quality

```bash
make lint              # Run golangci-lint on all packages
make help              # Show all available commands
```

---

## Automated Testing (CI/CD)

### Local: Pre-Commit Hook

When you commit, tests automatically run:

```bash
git commit -m "your message"
# → pre-commit hook runs tests
# → if tests fail, commit is blocked
# → if tests pass, commit succeeds
```

To skip the hook (only when necessary):

```bash
git commit --no-verify -m "your message"
```

### Remote: GitHub Actions

On every push and pull request:
- GitHub Actions runs the full test suite automatically
- Tests must pass before merging (can be enforced with branch protection)
- Coverage reports are uploaded to Codecov
- CI status shows in PR reviews

See `.github/workflows/test.yml` for workflow details.

---

## Understanding Test Results

### Coverage by Package

Current coverage after initial tests:

```
internal/auth       56.9% ✓ Good (JWT, middleware, password hashing)
tests/helpers       63.3% ✓ Good (test utilities fully tested)
internal/buckets     1.2% ⚠ Low (only duration calculation tested)
[others]            0.0% — No tests yet
```

### Coverage Goals

- **80%+** for security-critical code (auth, connections)
- **70%+** for business logic (tasks, rewards)
- **50%+** for handlers (can be brittle)
- **0%** is OK for UI-only code, vendor code, or templates

### Reading Coverage Reports

```bash
# Generate and open the report
make test-coverage
open coverage.html  # or equivalent for your OS
```

The report shows:
- Green = covered code (tested)
- Red = untested code
- Hover over code to see why it's untested

---

## Test Organization

Tests are colocated with source code:

```
pact/
├── internal/
│   ├── auth/
│   │   ├── jwt.go
│   │   ├── jwt_test.go           ← Tests for jwt.go
│   │   ├── middleware.go
│   │   ├── middleware_test.go     ← Tests for middleware.go
│   │   ├── service.go
│   │   └── service_test.go        ← Tests for service.go
│   └── ...
├── tests/
│   ├── helpers/
│   │   ├── test_db.go            ← Helper functions for tests
│   │   ├── test_users.go         ← Helper functions for tests
│   │   └── helpers_test.go       ← Tests for helpers
│   └── e2e/                      ← E2E tests (not yet implemented)
└── ...
```

**Naming convention:** `<file>_test.go` always contains tests for `<file>.go`

---

## Writing Tests

See `TESTING_PLAN.md` for detailed test structure and patterns.

### Quick Example

```go
// internal/mypackage/service_test.go
package mypackage

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestMyFunction(t *testing.T) {
    result := MyFunction("input")
    assert.Equal(t, "expected", result)
}
```

### Test Helpers

The `tests/helpers/` package provides:

```go
// Database setup
db, queries, cleanup := helpers.SetupTestDBWithPath(t, "../../database/schema.sql")
defer cleanup()

// User creation
user := helpers.CreateTestUser(t, queries, "email@test.com", "username", "password")

// Connection creation
connId := helpers.CreateTestConnection(t, queries, managerId, workerId)
```

---

## Continuous Integration Workflow

### Your Actions

1. **Code** → Make changes locally
2. **Test** → Run `make test` to verify locally
3. **Commit** → Pre-commit hook runs tests automatically
4. **Push** → GitHub Actions runs full test suite
5. **Review** → Code review + CI status must be green

### GitHub Actions (Automatic)

When you push or open a PR:

1. GitHub checks out your code
2. Sets up Go 1.23
3. Downloads and caches dependencies
4. Runs `go test ./...` with race detector
5. Generates coverage report
6. Uploads to Codecov.io

**View status:**
- GitHub: Shows in PR checks (green ✓ or red ✗)
- Codecov: Shows coverage trend at codecov.io/gh/yourusername/pact

---

## Troubleshooting

### Tests fail locally but work in CI

Usually means different Go version or environment:

```bash
# Check your Go version (should be 1.23.3)
go version

# Force dependencies to match go.mod
go mod download
go mod verify

# Clear test cache
go clean -testcache

# Re-run tests
make test
```

### Pre-commit hook failing

```bash
# See what tests are failing
make test-verbose

# Fix the issues, then try committing again
git commit -m "your message"

# Or skip hook (only if testing locally is hard)
git commit --no-verify -m "your message"
```

### Coverage report won't generate

```bash
# Make sure you're in nix-shell
nix-shell

# Regenerate
make clean
make test-coverage
```

### Golangci-lint not found

```bash
# If golangci-lint is missing:
nix-shell  # This should provide it

# Or install manually:
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

---

## Next Steps

### To Add More Tests

1. Pick a package with 0% coverage (see `make test-coverage`)
2. Create `<file>_test.go` next to the source
3. Write tests following patterns in existing test files
4. Run `make test` to verify
5. Commit with clear message: "test: add tests for <package>"

### To Improve CI/CD

Future enhancements:
- Add lint job to GitHub Actions
- Add security scanning (gosec)
- Add performance benchmarks
- Set up automatic dependabot updates
- Add code coverage badges to README
- Require 80% coverage before merge

---

## Resources

- **Testing Plan:** See `TESTING_PLAN.md` for comprehensive test strategy
- **Go Testing:** https://golang.org/pkg/testing/
- **Testify:** https://github.com/stretchr/testify (assertions we use)
- **GitHub Actions:** https://docs.github.com/en/actions

---

## Helpful Aliases (Optional)

Add to your shell profile for quick access:

```bash
# In ~/.bashrc or ~/.zshrc
alias pt='make test'
alias ptv='make test-verbose'
alias ptc='make test-coverage'
alias pdev='make dev'
```

Then use: `pt` instead of `make test`, etc.

