.PHONY: help test test-verbose test-coverage build dev clean lint install

help:
	@echo "Pact Development Commands"
	@echo "=========================="
	@echo ""
	@echo "Testing:"
	@echo "  make test              - Run all tests"
	@echo "  make test-verbose      - Run tests with verbose output"
	@echo "  make test-coverage     - Generate and view coverage report"
	@echo ""
	@echo "Development:"
	@echo "  make install           - Install npm dependencies"
	@echo "  make dev               - Start dev server with hot reload"
	@echo "  make build             - Build CSS and Go binary"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint              - Run code linter (requires golangci-lint)"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean             - Remove build artifacts and coverage files"
	@echo ""

# Testing targets
test:
	@echo "Running tests..."
	go test ./...

test-verbose:
	@echo "Running tests (verbose)..."
	go test -v ./...

test-coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

# Development targets
install:
	@echo "Installing npm dependencies..."
	npm install
	@echo "✓ Dependencies installed"

dev:
	@echo "Starting development server with hot reload..."
	./buildAir.sh

build:
	@echo "Building CSS..."
	npm run build:css
	@echo "Building Go binary..."
	go build -o ./tmp/main .
	@echo "✓ Build complete"

# Code quality
lint:
	@if ! command -v golangci-lint &> /dev/null; then \
		echo "golangci-lint not found. Install with:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
	@echo "Running linter..."
	golangci-lint run ./internal/... ./database/... ./tests/...

# Cleanup
clean:
	@echo "Cleaning up build artifacts..."
	rm -rf ./tmp/main coverage.out coverage.html
	@echo "✓ Cleanup complete"

# Convenience: run tests before commit
pre-commit: test lint
	@echo "✓ All checks passed"
