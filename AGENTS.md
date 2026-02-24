# AGENTS.md - Pact Development Guide

## Overview

**Pact** is a Go-based task management system with HTMX frontend and SQLite database. It allows users to manage manager-worker relationships, assign tasks, and track reward/punishment systems. Users authenticate via JWT, can subscribe via Stripe, and manage multiple concurrent "connections" (manager/worker relationships).

**Stack**: Go 1.23.3, HTMX, TailwindCSS + DaisyUI, SQLite, JavaScript, SQLC, Stripe API

---

## Essential Commands

### Development Setup & Running

```bash
# First time setup (install dependencies)
nix-shell          # Enter development environment
npm install        # Install npm dependencies

# Development with live reload
./buildAir.sh      # Recommended for development
                   # Runs: npm install → npm run build:css → npm run watch:css & air
                   # Builds CSS, watches for changes, hot-reloads Go code

# Alternative: Direct Go execution
go run main.go     # Simple run without live reload (port 8080)
```

### CSS/Styling

```bash
npm run build:css  # Build Tailwind CSS once
npm run watch:css  # Watch mode for CSS changes (used in buildAir.sh)
```

### Database

```bash
# SQLC code generation from SQL
sqlc generate      # Generates Go code from database/queries/ and database/schema.sql
                   # Output: database/query.sql.go (NEVER edit this file - auto-generated)

# Database is SQLite at database/database.db
# Queries stored in: database/queries/query.sql
# Schema stored in: database/schema.sql
```

### Testing

```bash
npm test           # Currently not configured - placeholder in package.json
# No Go tests currently in repo
```

### Server Details

- **Default Port**: 8080
- **Live Reload Proxy Port**: 8081 (when using Air)
- **Air Config**: `.air.toml` - defines build/reload behavior, includes templates in watch list

---

## Project Structure

```
pact/
├── main.go                          # Entry point: initializes DB, templates, router
├── go.mod / go.sum                  # Go dependencies
├── package.json / npm scripts       # CSS/styling builds
├── sqlc.yaml                        # SQLC configuration
├── .air.toml                        # Air live reload config
├── buildAir.sh                      # Development startup script
│
├── database/                        # Database layer
│   ├── db.go                        # Database connection initialization
│   ├── database.db                  # SQLite database file
│   ├── schema.sql                   # Database schema (tables, constraints)
│   ├── models.go                    # SQLC-generated Go type definitions
│   ├── query.sql.go                 # SQLC-generated database queries (DO NOT EDIT)
│   └── queries/
│       └── query.sql                # SQL query definitions for SQLC
│
├── internal/                        # Application logic
│   ├── router/
│   │   └── router.go                # HTTP route definitions (all endpoints)
│   ├── pages/
│   │   ├── handlers.go              # Page rendering handlers (login, register, etc.)
│   │   ├── render.go                # Template rendering system
│   │   └── models.go                # Page-specific data structures
│   ├── auth/
│   │   ├── jwt.go                   # JWT token generation/validation
│   │   ├── middleware.go            # AuthMiddleware - protects authenticated routes
│   │   ├── service.go               # Auth service functions
│   │   └── sessions.go              # Session management
│   ├── connections/
│   │   ├── handlers.go              # Connection request/management handlers
│   │   └── services.go              # Connection business logic
│   ├── user/
│   │   ├── handler.go               # User-related endpoints
│   │   └── service.go               # User service functions
│   ├── stripe/
│   │   └── handlers.go              # Stripe subscription handlers
│   ├── assistant/ / manager/        # Placeholder packages (minimal implementation)
│   ├── db/
│   │   └── db.go                    # Unused (legacy)
│   └── templates/
│       ├── contentTemplates/        # Full-page layout templates
│       │   ├── defaultLayout.html   # Base layout for all pages
│       │   ├── bucketsPage.html     # Tasks/buckets page
│       │   ├── accountPage.html     # User account page
│       │   ├── loginPage.html       # Login page
│       │   ├── registerPage.html    # Registration page
│       │   └── ...
│       └── fractions/               # Reusable template fragments (HTMX responses)
│           ├── buckets.html         # Buckets list fragment
│           ├── connections.html     # Connections UI fragment
│           ├── guestNavbar.html     # Navigation for guests
│           ├── memberNavbar.html    # Navigation for members
│           └── ... (other fragments)
│
├── static/                          # Static CSS assets
│   ├── styles.css                   # Input CSS for Tailwind
│   └── index.css                    # Generated CSS output (built by Tailwind)
│
├── web/                             # Frontend assets
│   ├── package.json                 # (appears unused)
│   └── index.html                   # (appears unused)
│
└── shell.nix                        # Nix development environment definition
```

---

## Key Patterns & Conventions

### Naming Conventions

**Database Layer (SQLC-generated)**:
- Table names: `snake_case` (e.g., `user_id`, `connection_id`)
- Query functions: `PascalCase` (e.g., `GetUserPendingRequests()`, `CreateConnection()`)
- Generated types: `PascalCase` (e.g., `GetUserPendingRequestsRow`)

**Handler Functions**:
- Naming: `Serve<FeatureName>` or `Handle<Action>` (e.g., `ServeBucketsPage`, `HandleCreateCheckoutSession`)
- All handlers: `func(w http.ResponseWriter, r *http.Request)`

**Template Naming**:
- Full page layouts: `<name>Page.html` in `contentTemplates/`
- Fragments/partials: `<name>.html` in `fractions/` (returned by HTMX requests)

**Context Values**:
- User ID stored in request context as `"userID"` (set by `AuthMiddleware`)
- Type assertion: `r.Context().Value("userID").(int)`

### HTTP & Routing

- **Router**: Standard Go `http.ServeMux` in `internal/router/router.go`
- **Route Format**: `"METHOD /path"` (e.g., `"GET /", "POST /login"`)
- **Authentication**: Routes using `auth.AuthMiddleware()` are protected - only authenticated users can access
- **HTMX Integration**: Handlers return HTML fragments (not JSON) for HTMX to swap into the DOM

Example route:
```go
mux.HandleFunc("GET /bucketContent", auth.AuthMiddleware(pages.ServeBucketsContent))
```

### Template System

**Architecture**:
- Initialized in `main.go` via `pages.InitTemplates()`
- Stored in `TemplateConstruct` with two maps: `layouts` and `fractions`
- Layouts include `defaultLayout.html` which wraps page content
- Fractions are reusable snippets for HTMX requests

**Rendering**:
```go
// Full page (wraps with layout)
pages.RenderLayoutTemplate(w, r, "templateName", data)

// Fragment only (e.g., HTMX response)
pages.RenderTemplateFraction(w, "fragmentName", data)
```

**Template Data Structure**:
```go
type TemplateData struct {
    Data map[string]any
}
```

**Adding Templates**:
1. Create `.html` in `contentTemplates/` (full pages) or `fractions/` (fragments)
2. Call `pages.InitTemplates()` on startup (already done in `main.go`)
3. Reference by name (without `.html`) in render calls
4. Air automatically watches `internal/templates/` - changes trigger rebuild

### Authentication & JWT

**Token Generation**:
- `auth.GenerateToken(userId uint)` → JWT string
- Token valid for 6 hours from generation time
- JWT_SECRET_KEY read from `.env` file at init time
- Token stored in HTTP cookie (set by auth handlers)

**Middleware Pattern**:
```go
auth.AuthMiddleware(handler func) → wrapped handler
// Extracts & validates JWT from cookie
// Sets userID in request context
// Rejects unauthenticated requests with 401
```

**Usage**:
```go
mux.HandleFunc("GET /protected", auth.AuthMiddleware(pages.ServeProtectedPage))
```

### Database & SQLC

**Workflow**:
1. Write SQL queries in `database/queries/query.sql`
2. Run `sqlc generate` to auto-generate Go functions
3. Output in `database/query.sql.go` (DO NOT EDIT by hand)

**Query Structure**:
- Named queries: `-- name: GetUserByID :one` (returns single row)
- SQLC generates type-safe Go functions with return types

**Usage Pattern**:
```go
// Get queries from SQLC-generated code
queries := database.New(db)
rows, err := queries.GetUserPendingRequests(ctx, userId)
```

**Database Connection**:
- Initialized in `database/OpenDatabase()` called from `main.go`
- SQLite database file: `database/database.db`
- Connection managed globally

### Services & Handlers

**Pattern**:
- `handlers.go`: HTTP handlers that:
  - Extract data from request
  - Call service functions
  - Render templates with results
- `services.go`: Business logic that:
  - Queries database
  - Processes data
  - Returns results to handlers

**Error Handling Convention**:
- Service functions return `(result, error)`
- Handlers check errors and call `http.Error()` with appropriate status
- Example: `http.Error(w, "error message", http.StatusInternalServerError)`

### Recent Work: Connections System

**Current State** (as of last commits):
- Refactoring "connections" (manager/worker relationships)
- Pending requests tracked with roles (who wants to be manager/worker)
- In-progress: Mapping pending requests to roles, sending UI data
- Known issues: Errors in `internal/connections/services.go` related to data structures

**Data Structures**:
- `GetUserPendingRequestsRow`: Rows from pending connection requests
- Maps created to associate requests with desired roles
- Being refactored to send comprehensive data to frontend

---

## Environment & Configuration

### `.env` File (Required)

```
JWT_SECRET_KEY=<your-secret-key>
STRIPE_SECRET_KEY=<stripe-key>
# Other potential keys as needed
```

### Development Environment

- **Nix Shell**: Provided via `shell.nix` - ensures reproducible Go/Node versions
- **Air**: Live reload tool for Go (configured in `.air.toml`)
- **Tailwind**: CSS framework with DaisyUI component library

### Build Artifacts

- Air output: `./tmp/main` (binary)
- CSS output: `static/index.css` (generated from `static/styles.css`)
- SQLC output: `database/query.sql.go` (auto-generated from SQL queries)

---

## Common Tasks & Workflows

### Adding a New Route & Handler

1. Add route in `internal/router/router.go`
2. Create handler in appropriate package (e.g., `internal/pages/handlers.go`)
3. Handler signature: `func(w http.ResponseWriter, r *http.Request)`
4. Use `auth.AuthMiddleware()` if authentication required
5. Render template with `RenderLayoutTemplate()` or `RenderTemplateFraction()`

### Adding a Database Query

1. Write SQL query in `database/queries/query.sql`
   - Use SQLC comment syntax: `-- name: QueryName :one/:many`
2. Run `sqlc generate`
3. Use generated function from `database.New(db)` in handlers/services
4. Example: `queries.QueryName(ctx, param1, param2)`

### Adding a Template

1. Create `.html` file in `contentTemplates/` (full page) or `fractions/` (snippet)
2. Use `.Data.KeyName` to access template data
3. Air watches and rebuilds on save
4. Reference by name in render call (no `.html` extension)

### Debugging User Context

```go
userId := r.Context().Value("userID").(int)
if userId < 1 {
    http.Error(w, "userID not found in context", http.StatusUnauthorized)
    return
}
```

### Common Output/Logging

```go
fmt.Println("message")      // Standard output
fmt.Printf("format: %v\n", value)
log.Println("message")      // Logging (used in main.go)
log.Fatalf("fatal: %v", err) // Fatal error
```

---

## Gotchas & Non-Obvious Patterns

### 1. **Context User ID Type Assertion**
Always assert as `int`, not `uint` or string:
```go
userId := r.Context().Value("userID").(int)  // ✓ Correct
```

### 2. **SQLC Auto-Generated Files**
`database/query.sql.go` is **auto-generated by sqlc** - NEVER edit by hand. Always modify `database/queries/query.sql` and regenerate.

### 3. **Template System Initialization**
Templates are parsed once on startup in `pages.InitTemplates()`. Changes to templates require server restart (Air handles this automatically with `include_ext = ["go", "tpl", "tmpl", "html"]` and `include_dir = ["internal/templates"]`).

### 4. **Pending Connection Requests**
The connections system tracks who is requesting to be manager vs worker via:
- `SenderID` = who initiated the request
- `SuggestedManagerID` / `SuggestedWorkerID` = role preferences
- Logic: If `SenderID == SuggestedManagerID`, sender wants to be manager

### 5. **Fragment vs Layout Templates**
- **Layouts** (`RenderLayoutTemplate`): Include `defaultLayout.html` wrapper, used for full page loads
- **Fractions** (`RenderTemplateFraction`): Raw fragment only, used for HTMX partial updates
- HTMX requests typically target fractions; browser navigation uses layouts

### 6. **Air Configuration for Templates**
Air is configured to watch `internal/templates/` for changes:
```toml
include_dir = ["internal/templates"]
include_ext = ["go", "tpl", "tmpl", "html"]
```
CSS changes also trigger rebuild via `pre_cmd = "npm run build:css"`.

### 7. **Stripe & AWS Integration Mentioned**
README mentions:
- Stripe subscription system (implemented)
- AWS S3 bucket for file storage (referenced in README, minimal handler code visible)
- These are partially implemented - check handlers for actual integration details

### 8. **Multiple Connections Per User**
A user can have multiple manager/worker relationships with different users, and can be both manager and worker in different connections. `active_connection_id` tracks current active context.

### 9. **Database Connection Management**
Database opened once in `main.go` via `database.OpenDatabase()`. All handlers receive same DB connection. No transaction handling visible in current code - watch for SQLite concurrency if adding complex features.

### 10. **Error Handling Convention**
Most handlers return errors to client via `http.Error()`. Check for patterns of silent failures (missing error checks) - ongoing refactoring is improving error handling.

---

## Testing & Verification

- **No automated tests** currently in repo
- Manual testing via browser at `http://localhost:8080`
- Development flow: Edit code → `./buildAir.sh` auto-rebuilds → browser reload

---

## Dependencies & Versions

**Go 1.23.3**:
- `github.com/golang-jwt/jwt/v5` - JWT authentication
- `github.com/lib/pq` - PostgreSQL driver (installed but SQLite used)
- `github.com/mattn/go-sqlite3` - SQLite driver
- `github.com/stripe/stripe-go/v79` - Stripe API
- `golang.org/x/crypto` - Password hashing (bcrypt)
- `github.com/joho/godotenv` - Environment variable loading

**Node/npm**:
- `tailwindcss` ^3.4.12 - CSS framework
- `daisyui` ^4.12.14 - Component library
- Air - live reload (installed via nix-shell or system)
- sqlc - SQL code generation (installed via nix-shell or system)

---

## Next Steps / In-Progress Work

Based on recent commits, the team is currently:
1. **Refactoring connection request handling** - improving how pending requests are mapped to roles
2. **Improving error handling throughout** - replacing silent failures with proper error messages
3. **Data structure refactoring** - moving data into maps for cleaner frontend integration

Check git history and modified files for latest context on ongoing work.

