package auth

import (
	"context"
	"net/http"
	"strings"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Redirect(w, r, "/loginPage", http.StatusFound)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			http.Redirect(w, r, "/loginPage", http.StatusFound)
			return
		}

		userID, err := ValidateToken(bearerToken[1])
		if err != nil {
			// if the token is invalid, send them to the login page...
			http.Redirect(w, r, "/loginPage", http.StatusFound)
			return
		}

		// Add the user ID to the request context
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
