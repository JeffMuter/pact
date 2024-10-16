package auth

import (
	"context"
	"fmt"
	"net/http"
	"pact/database"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("authenticating...")

		// Get the Bearer cookie
		cookie, err := r.Cookie("Bearer")
		if err != nil {
			if err == http.ErrNoCookie {
				fmt.Println("No Bearer cookie found")
				http.Redirect(w, r, "/loginPage", http.StatusFound)
				return
			}
			// For any other type of error, return a bad request status
			fmt.Printf("Error getting cookie: %v\n", err)
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// The JWT is the value of the cookie
		token := cookie.Value
		fmt.Printf("JWT from cookie: %v\n", token)

		userID, err := ValidateToken(token)
		if err != nil {
			fmt.Printf("Error validating token: %v\n", err)
			// If the token is invalid, send them to the login page
			http.Redirect(w, r, "/loginPage", http.StatusFound)
			return
		}

		// determine if a member to add to context
		queries := database.GetQueries()
		ctx := context.Background()

		isMember, err := queries.UserIsMemberById(ctx, int64(userID))

		// if isMember == true, then we want to apply this to the context somehow?
		if isMember == 1 {
			ctx = context.WithValue(r.Context(), "authStatus", "member")
		} else if userID != 0 {
			ctx = context.WithValue(r.Context(), "authStatus", "registered")
		}
		// Add the user ID to the request context
		ctx = context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
