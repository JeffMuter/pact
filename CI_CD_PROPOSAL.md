# CI/CD & Test Automation Proposal for Pact

## Overview
Implement automated testing via GitHub Actions with local Nix integration for reproducible builds and tests.

---

## Part 1: GitHub Actions Workflow

### Create `.github/workflows/test.yml`

This workflow runs on every push and PR:
- Run Go unit/integration tests
- Check code formatting
- Run linters (optional)
- Generate coverage reports
- Fail if tests don't pass

**Key decisions:**
- Run tests in parallel jobs for speed
- Cache Go modules and Nix closures to speed up repeated runs
- Generate coverage badge/report for visibility
- Require tests to pass before merge (via branch protection rule)

---

## Part 2: Enhanced Nix Shell

### Update `shell.nix`

Add testing tools:
- `go` (already present)
- `golangci-lint` (code linting - optional but recommended)
- Keep existing build tools (`sqlc`, `air`, `tailwindcss`, etc.)

Add helpful test commands to `shellHook`:
```bash
go test ./...              # Run all tests
go test -v ./...           # Verbose output
go test -cover ./...       # Coverage report
go test -run TestName ./pkg  # Run specific test
```

---

## Part 3: Local Testing Helpers

### Create `Makefile` or shell aliases

Make common tasks accessible:
```bash
make test          # Run all tests
make test-verbose  # Verbose test output
make test-coverage # Generate coverage report
make lint          # Run linters (if added)
make build         # Build project
make dev           # Start dev server with hot reload
```

---

## Part 4: Git Hooks (Optional but Recommended)

### Add pre-commit hook

Automatically run tests before committing (can be skipped with `--no-verify`):
```bash
#!/bin/bash
go test ./... || exit 1
```

This prevents accidentally committing broken code.

---

## Proposed File Structure

```
pact/
├── .github/
│   └── workflows/
│       └── test.yml           # GitHub Actions workflow (NEW)
├── shell.nix                  # Enhanced with test tools (MODIFIED)
├── Makefile                   # Test/build shortcuts (NEW)
├── buildAir.sh                # Keep as-is (unchanged)
├── .air.toml                  # Keep as-is (unchanged)
├── .gitignore                 # Update to exclude coverage files (MODIFIED)
├── go.mod, go.sum             # Already have testify
└── [rest of project...]
```

---

## Detailed Changes

### 1. `.github/workflows/test.yml` (NEW)

```yaml
name: Tests

on:
  push:
    branches: [master, develop]
  pull_request:
    branches: [master]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.23']
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go-version }}
          cache: 'go'
      
      - name: Run tests
        run: go test -v -cover ./...
      
      - name: Generate coverage report
        run: go test -coverprofile=coverage.out ./...
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

### 2. `shell.nix` (MODIFIED)

Update `buildInputs` to include:
- `golangci-lint` (optional, for linting)
- Keep everything else

Update `shellHook` with test commands.

### 3. `Makefile` (NEW)

```makefile
.PHONY: test test-verbose test-coverage build dev clean lint

test:
	go test ./...

test-verbose:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	golangci-lint run

build:
	npm run build:css
	go build -o ./tmp/main .

dev:
	./buildAir.sh

clean:
	rm -rf ./tmp/main coverage.out coverage.html
```

### 4. `.gitignore` (MODIFIED)

Add:
```
coverage.out
coverage.html
.coverage/
```

---

## Workflow Summary

### Local Development
```bash
nix-shell
npm install
./buildAir.sh           # Dev server with hot reload
# In another terminal:
make test               # Run tests locally
make test-coverage      # Check coverage
```

### Before Pushing Code
```bash
make test               # Ensure tests pass locally
make lint               # Check code quality (optional)
```

### On Push/PR
- GitHub Actions automatically runs full test suite
- Coverage report uploaded to codecov.io
- Tests must pass before merge (can be enforced with branch protection)

---

## Benefits

✅ **Reproducibility:** Nix ensures identical environment everywhere  
✅ **Speed:** Cached dependencies reduce CI time  
✅ **Visibility:** Coverage badges & reports show test health  
✅ **Quality:** Linting catches issues automatically  
✅ **Safety:** Pre-commit hooks prevent broken commits  
✅ **CI/CD:** Tests run on every push/PR, blocks bad merges  

---

## Implementation Order

1. **First:** Create `.github/workflows/test.yml` (enables CI)
2. **Second:** Update `shell.nix` (better DX)
3. **Third:** Create `Makefile` (convenient shortcuts)
4. **Fourth:** Add `.git/hooks/pre-commit` (safety net)
5. **Fifth:** Update `.gitignore` (clean repo)

---

## Next Steps After Implementation

Once CI/CD is working:
1. ✅ Fix any failing tests in CI
2. ✅ Set up branch protection on `master`: require tests to pass
3. ✅ Add codecov badge to README
4. ✅ Share CI status in team communications
5. ✅ Continue adding tests incrementally (one test suite per week)

---

## Notes

- The GitHub Actions workflow uses `go test ./...` which will discover and run all `*_test.go` files automatically
- Cache enabled via `actions/setup-go@v4` with `cache: 'go'` - this speeds up subsequent runs significantly
- Coverage upload is optional but recommended for tracking test health over time
- Nix shell ensures the same Go version everywhere (1.23.3)
- Can add more jobs (lint, security scan, etc.) later as needed

