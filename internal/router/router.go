package router

import (
	"net/http"
	"pact/internal/auth"
	"pact/internal/stripe"
)

func Router() *http.ServeMux {

	mux := http.NewServeMux()

	// Register handlers
	//	mux.Handle("/", auth.AuthMiddleware(http.HandlerFunc(pages.ServeHomePage)))

	mux.HandleFunc("GET /login", auth.ServeLoginPage)
	mux.HandleFunc("GET /loginForm", auth.ServeLoginForm)
	mux.HandleFunc("POST /login", auth.LoginFormHandler)

	mux.HandleFunc("GET /register", auth.ServeRegistrationPage)
	mux.HandleFunc("GET /registerForm", auth.ServeRegistrationForm)
	mux.HandleFunc("POST /registeruser", auth.RegisterHandler)

	mux.HandleFunc("GET /stripePage", stripe.ServeMembershipPage)
	mux.HandleFunc("GET /stripeForm", stripe.ServeMembershipForm)

	// Serve static files
	fileServer := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
	mux.Handle("/js/", http.StripPrefix("/js/", fileServer))
	mux.Handle("/images/", http.StripPrefix("/images/", fileServer))
	return mux
}
