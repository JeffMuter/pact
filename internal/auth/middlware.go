package auth

import (
	"context"
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

				// Check if the user is a member
				isMember, err := queries.UserIsMemberById(ctx, int64(userID))
				if err == nil {
					// Set authStatus based on membership
					if isMember == 1 {
						authStatus = "member"
					} else {
						authStatus = "registered"
					}
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
