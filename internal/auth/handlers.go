package auth

import (
	"fmt"
	"net/http"
	"pact/database"
)

func HandleLoginProcedure(w http.ResponseWriter, r *http.Request, user *database.User) {
	fmt.Println("handling login procedure...")
	token, err := GenerateToken(uint(user.UserID))
	if err != nil {
		http.Error(w, "error generating token", http.StatusInternalServerError)
		return
	}
	fmt.Println("token generated...")

	isSecure := r.TLS != nil
	sameSite := http.SameSiteStrictMode
	if !isSecure {
		sameSite = http.SameSiteLaxMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    token,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: sameSite,
		Path:     "/",
	})
	fmt.Println("cookie set in server...")
	http.Redirect(w, r, "/homeContent", http.StatusSeeOther)
}
