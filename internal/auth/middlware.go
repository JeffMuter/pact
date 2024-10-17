package auth

import (
	"context"
	"net/http"
	"pact/database"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authStatus := "guest"
		var userID int

		cookie, err := r.Cookie("Bearer")
		if err == nil {
			token := cookie.Value
			userID, err = ValidateToken(token)
			if err == nil {
				queries := database.GetQueries()
				isMember, err := queries.UserIsMemberById(ctx, int64(userID))
				if err == nil {
					if isMember == 1 {
						authStatus = "member"
					} else {
						authStatus = "registered"
					}
				}
			}
		}

		// Always set the authStatus and userID in the context
		ctx = context.WithValue(ctx, "authStatus", authStatus)
		ctx = context.WithValue(ctx, "userID", userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
