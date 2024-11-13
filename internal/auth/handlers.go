package auth

import (
	"fmt"
	"net/http"
	"pact/database"
	"time"
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
		Expires:  time.Now().Add(24 * time.Hour),
	})
	fmt.Println("cookie set in server...")
	http.Redirect(w, r, "/homeContent", http.StatusSeeOther)
}

// Logout is a handler meant to alter the cookie to expire it, and reroute the user to a
// full rerender of the login page.
func Logout(w http.ResponseWriter, r *http.Request) {
	fmt.Println("logging out begin...")
	// create a cookie called 'Bearer', set it in the past to invalidate it. overwriting a good cookie
	cookie := &http.Cookie{
		Name:     "Bearer",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	fmt.Println("logout cookie MaxAge set -1")
	http.SetCookie(w, cookie)

	// redirect to login page.
	http.Redirect(w, r, "/loginPage", http.StatusSeeOther)
}
