package auth

import (
	"context"
	"fmt"
	"net/http"
	"pact/database"
)

// AuthMiddleware is a function that returns an http.HandlerFunc
// It's used to authenticate requests and set user information in the context
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	// Return a new handler function that wraps the input handler
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the current context from the request
		ctx := r.Context()

		// Initialize default values
		authStatus := "guest"
		var userID int

		// Try to get the "Bearer" cookie from the request
		cookie, err := r.Cookie("Bearer")
		if err == nil {
			// If cookie exists, extract its value (the token)
			token := cookie.Value

			// Validate the token and get the user ID
			userID, err = ValidateToken(token)
			if err == nil {
				// If token is valid, get the database queries object
				queries := database.GetQueries()

				authStatus = "registered"
				activeConnId, err := queries.GetActiveConnectionId(ctx, int64(userID))
				if err == nil && activeConnId.Valid {
					authStatus = "member"
				}
			} else {
				http.Redirect(w, r, "/loginPage", http.StatusSeeOther)

			}
		} else {
			http.Redirect(w, r, "/loginPage", http.StatusSeeOther)
		}

		// Always set the authStatus and userID in the context
		// This ensures that even if authentication fails, we have default values
		ctx = context.WithValue(ctx, "authStatus", authStatus)
		ctx = context.WithValue(ctx, "userID", userID)

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// OptionalAuthMiddleware checks auth status but doesn't redirect
// Used for public pages that should show different content based on auth state
func OptionalAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authStatus := "guest"
		var userID int

		cookie, err := r.Cookie("Bearer")
		fmt.Printf("[OptionalAuth] Cookie lookup error: %v\n", err)
		if err == nil {
			fmt.Printf("[OptionalAuth] Cookie value present: %d chars\n", len(cookie.Value))
			token := cookie.Value
			userID, err = ValidateToken(token)
			fmt.Printf("[OptionalAuth] ValidateToken error: %v, userID: %d\n", err, userID)
			if err == nil {
				queries := database.GetQueries()
				authStatus = "registered"
				activeConnId, err := queries.GetActiveConnectionId(ctx, int64(userID))
				fmt.Printf("[OptionalAuth] GetActiveConnectionId error: %v, valid: %v\n", err, activeConnId.Valid)
				if err == nil && activeConnId.Valid {
					authStatus = "member"
				}
			}
		}

		fmt.Printf("[OptionalAuth] Final authStatus: %s, userID: %d\n", authStatus, userID)
		ctx = context.WithValue(ctx, "authStatus", authStatus)
		ctx = context.WithValue(ctx, "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
