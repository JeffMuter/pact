#! /usr/bin/env nix-shell
#! nix-shell -i bash -p nixpkgs.shellcheck
{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = with pkgs; [ 
      # Go development
      go
      
      # Build and database tools
      sqlc
      air
      sqlite
      
      # Web/CSS
      tailwindcss
      nodejs_20
      
      # Development utilities
      git
      pkg-config
      gcc
      gnumake
      
      # Code quality (testing & linting)
      golangci-lint
      
      # PostgreSQL client (for database interaction if needed)
      postgresql
    ];

  shellHook = ''
    export PATH="$PWD/node_modules/.bin:$PATH"
    
    cat << "EOF"
╔════════════════════════════════════════════════════════════════╗
║                         PACT Environment                       ║
╚════════════════════════════════════════════════════════════════╝

🚀 GETTING STARTED:
  • npm install           — Install dependencies
  • make dev              — Run dev server with hot reload
  • make help             — Show all available commands

🧪 TESTING:
  • make test             — Run all tests
  • make test-verbose     — Run tests with verbose output
  • make test-coverage    — Generate coverage report (coverage.html)

🔨 COMMON TASKS:
  • sqlc generate         — Generate database code from SQL
  • npm run build:css     — Build Tailwind CSS
  • npm run watch:css     — Watch CSS changes
  • air                   — Hot reload Go server (used by: make dev)

📊 DATABASE:
  • sqlite3 ./database/database.db — Access SQLite database
  • Review ./database/schema.sql for schema

📁 PROJECT STRUCTURE:
  • ./cmd/          — Go entry points
  • ./internal/     — Go packages
  • ./database/     — SQL queries & migrations
  • ./static/       — Frontend assets (CSS, HTML)
  • ./web/          — Web dependencies

💡 TIP: Run 'make help' for all available commands

EOF
  '';
}
