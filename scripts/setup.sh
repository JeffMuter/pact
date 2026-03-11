#!/bin/bash
# Setup script for Pact development environment
# Run this after cloning the repository

set -e

echo "🔧 Setting up Pact development environment..."
echo ""

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo "✗ Not in a git repository. Please run this from the project root."
    exit 1
fi

# Install git hooks
echo "📌 Installing git hooks..."
if [ -f "scripts/pre-commit" ]; then
    cp scripts/pre-commit .git/hooks/pre-commit
    chmod +x .git/hooks/pre-commit
    echo "  ✓ Pre-commit hook installed"
else
    echo "  ✗ scripts/pre-commit not found"
    exit 1
fi

echo ""
echo "✓ Setup complete!"
echo ""
echo "Next steps:"
echo "  1. Enter nix shell: nix-shell"
echo "  2. Install dependencies: npm install"
echo "  3. Start dev server: make dev"
echo "  4. Run tests: make test"
echo ""
echo "For more commands, run: make help"
echo ""
