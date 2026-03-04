# Development Guide

## Table of Contents
- [Development Workflow](#development-workflow)
- [Creating New Pages](#creating-new-pages)
- [Authentication Middleware](#authentication-middleware)
- [Working with Air](#working-with-air)
- [Common Issues](#common-issues)

## Development Workflow

### Starting the Server

**Development mode (recommended):**
```bash
air
```

This starts:
- Go server on `:8080` (built automatically)
- Proxy server on `:8081` (with asset pipeline)
- Auto-rebuild on `.go`, `.html` file changes
- Tailwind CSS compilation

**Always use `http://localhost:8081` in your browser for development.**

### Manual Build

If you need to test without Air:
```bash
npm run build:css && go build -o ./tmp/main .
./tmp/main
```

Server runs on `:8080` but **CSS won't hot-reload**.

## Creating New Pages

### Step 1: Create the Templates

**Fraction template** (`internal/templates/fractions/mypage.html`):
```html
{{ define "mypage" }}
<div>
  <!-- Your page content here -->
</div>
{{ end }}
```

**Content template** (`internal/templates/contentTemplates/mypagePage.html`):
```html
{{ define "content" }}
{{ template "mypage" . }}
{{ end }}
```

### Step 2: Create the Handler

In `internal/pages/handlers.go`:
```go
func ServeMyPage(w http.ResponseWriter, r *http.Request) {
    data := TemplateData{
        Data: map[string]any{
            "Title": "My Page Title",
        },
    }
    
    // Check if this is an HTMX request
    if r.Header.Get("HX-Request") == "true" {
        RenderTemplateFraction(w, "mypage", data)
        return
    }
    
    RenderLayoutTemplate(w, r, "mypagePage", data)
}
```

**Important:** Always handle both HTMX requests (for SPA navigation) and full page loads (for refreshes/direct access).

### Step 3: Register the Route

In `internal/router/router.go`:
```go
// Public page (no auth required, but shows different content based on auth state)
mux.HandleFunc("GET /mypage", auth.OptionalAuthMiddleware(pages.ServeMyPage))

// Protected page (requires login)
mux.HandleFunc("GET /mypage", auth.AuthMiddleware(pages.ServeMyPage))

// Completely public page (no auth checking at all)
mux.HandleFunc("GET /mypage", pages.ServeMyPage)
```

## Authentication Middleware

### Three Types of Middleware

#### 1. `auth.AuthMiddleware`
**Use for:** Protected pages that require login

```go
mux.HandleFunc("GET /account", auth.AuthMiddleware(pages.ServeAccountPage))
```

**Behavior:**
- Checks for valid `Bearer` cookie
- Redirects to `/loginPage` if not authenticated
- Sets `authStatus` in context: `"registered"` or `"member"`
- Always requires authentication

#### 2. `auth.OptionalAuthMiddleware`
**Use for:** Public pages that show different content based on auth state

```go
mux.HandleFunc("GET /description", auth.OptionalAuthMiddleware(pages.ServeDescriptionPage))
```

**Behavior:**
- Checks for valid `Bearer` cookie
- **Does NOT redirect** if not authenticated
- Sets `authStatus` in context: `"guest"`, `"registered"`, or `"member"`
- Allows both logged-in and logged-out users

#### 3. No Middleware
**Use for:** Static pages, API endpoints, or pages that don't need auth context

```go
mux.HandleFunc("GET /health", pages.ServeHealthCheck)
```

**Behavior:**
- No auth checking
- No `authStatus` in context
- If using `RenderLayoutTemplate`, it will default to guest navbar

### Why OptionalAuthMiddleware Exists

**Problem:** Pages without middleware don't have `authStatus` in the request context. When you refresh a page (full reload), `RenderLayoutTemplate` tries to read `authStatus` to determine which navbar to show. If it's missing, it defaults to "guest", making you appear logged out.

**Solution:** `OptionalAuthMiddleware` ensures `authStatus` is set even for public pages, so refreshing preserves your logged-in state.

### Cookie Configuration

The authentication cookie must have these properties:

```go
http.SetCookie(w, &http.Cookie{
    Name:     "Bearer",           // Must be exactly "Bearer"
    Value:    token,
    HttpOnly: true,
    Secure:   isSecure,           // true for HTTPS
    SameSite: sameSite,           // Strict or Lax
    Path:     "/",                // REQUIRED - cookie must be available on all routes
    Expires:  time.Now().Add(24 * time.Hour), // REQUIRED - set expiration
})
```

**Common mistakes:**
- ❌ Wrong name: `"Bearer token"` instead of `"Bearer"`
- ❌ Missing `Path: "/"` - cookie won't be sent on all routes
- ❌ Missing `Expires` - cookie becomes session-only

## Working with Air

### How Air Works

Air watches for file changes and automatically:
1. Runs `npm run build:css` (compiles Tailwind)
2. Builds Go binary: `go build -o ./tmp/main .`
3. Restarts the server
4. Proxies `:8080` → `:8081` with proper headers

### What Air Watches

**Included:**
- All `.go` files in the project
- All `.html` files in `internal/templates/`

**Excluded:**
- `tmp/` - build artifacts
- `vendor/` - dependencies
- `testdata/` - test fixtures
- `*_test.go` - test files

### When Air Doesn't Rebuild

**Symptom:** You make code changes but the server behavior doesn't change.

**Causes:**
1. Air crashed silently
2. Build error (check `build-errors.log`)
3. Air is still running old process

**Solution:**
```bash
# Stop Air
Ctrl+C

# Clean build artifacts
rm -rf tmp/

# Restart Air
air
```

### Forcing a Rebuild

If Air is running but not picking up changes:

```bash
# In another terminal
touch internal/templates/fractions/description.html
```

This triggers Air's watcher since it monitors `.html` files.

### Debugging Air

**Check if Air is running:**
```bash
ps aux | grep air
```

**Check if server is running:**
```bash
ps aux | grep tmp/main
```

**View build errors:**
```bash
cat build-errors.log
```

**Test direct connection (bypass proxy):**
```bash
curl http://localhost:8080/
```

## Common Issues

### Issue: "Logged out" after refreshing page

**Symptoms:**
- Navigate to page via HTMX: ✅ Shows correct navbar
- Refresh the same page: ❌ Shows guest navbar

**Cause:** Route has no middleware, so `authStatus` isn't set in context.

**Solution:** Use `OptionalAuthMiddleware` for public pages:
```go
mux.HandleFunc("GET /mypage", auth.OptionalAuthMiddleware(pages.ServeMyPage))
```

### Issue: Changes not appearing in browser

**Symptoms:**
- Code changes don't affect the running server
- Old behavior persists

**Causes & Solutions:**

1. **Using wrong port**
   - ❌ `http://localhost:8080` - direct server, no CSS hot-reload
   - ✅ `http://localhost:8081` - proxied server with full pipeline

2. **Air not rebuilding**
   ```bash
   # Restart Air
   Ctrl+C
   air
   ```

3. **Browser cache**
   - Hard refresh: `Ctrl+Shift+R` (Linux/Windows) or `Cmd+Shift+R` (Mac)

4. **CSS changes not appearing**
   ```bash
   # Manually rebuild CSS
   npm run build:css
   ```

### Issue: Cookie not persisting

**Symptoms:**
- Login works, but refreshing logs you out
- Cookie doesn't appear in browser DevTools

**Checklist:**
1. Cookie name is exactly `"Bearer"`
2. Cookie has `Path: "/"`
3. Cookie has `Expires` set
4. Using HTTPS? Set `Secure: true`
5. Check browser's cookie storage in DevTools

### Issue: Template not found

**Symptoms:**
- Error: "The template X does not exist"

**Causes & Solutions:**

1. **Fraction template naming**
   ```go
   // File: internal/templates/fractions/mypage.html
   {{ define "mypage" }}  // ✅ Name matches filename
   {{ define "my-page" }} // ❌ Wrong name
   ```

2. **Content template naming**
   ```go
   // File: internal/templates/contentTemplates/mypagePage.html
   {{ define "content" }}  // ✅ Always "content"
   {{ define "mypage" }}   // ❌ Wrong name
   ```

3. **Handler mismatch**
   ```go
   // Template: internal/templates/contentTemplates/mypagePage.html
   RenderLayoutTemplate(w, r, "mypagePage", data)  // ✅ Matches filename
   RenderLayoutTemplate(w, r, "mypage", data)      // ❌ Wrong name
   ```

### Issue: HTMX navigation works, but refresh breaks

**Symptoms:**
- Clicking navbar links: ✅ Works
- Refreshing the page: ❌ Error or wrong content

**Cause:** Handler only handles HTMX requests, not full page loads.

**Solution:** Always handle both:
```go
func ServeMyPage(w http.ResponseWriter, r *http.Request) {
    data := TemplateData{...}
    
    if r.Header.Get("HX-Request") == "true" {
        RenderTemplateFraction(w, "mypage", data)  // HTMX navigation
        return
    }
    
    RenderLayoutTemplate(w, r, "mypagePage", data)  // Full page load
}
```

## Quick Reference

### Development Checklist for New Pages

- [ ] Create fraction template in `internal/templates/fractions/`
- [ ] Create content template in `internal/templates/contentTemplates/`
- [ ] Create handler in `internal/pages/handlers.go`
- [ ] Handle both HTMX and full page loads
- [ ] Register route in `internal/router/router.go`
- [ ] Choose appropriate middleware (`Auth`, `OptionalAuth`, or none)
- [ ] Test via HTMX navigation (click links)
- [ ] Test via refresh (F5)
- [ ] Test while logged out
- [ ] Test while logged in

### Essential Commands

```bash
# Start development server
air

# Manual build
go build -o ./tmp/main .

# Build CSS
npm run build:css

# Watch CSS (separate terminal)
npm run watch:css

# Check running processes
ps aux | grep -E "(air|tmp/main)"

# Kill everything
pkill -f air && pkill -f "tmp/main"
```

### Ports

- `:8080` - Direct Go server (no hot-reload)
- `:8081` - Air proxy (use this for development)
