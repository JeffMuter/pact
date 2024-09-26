package router

import (
	"net/http"
	"pact/internal/auth"
	"pact/internal/pages"
	"pact/internal/stripe"
)

func Router() *http.ServeMux {

	mux := http.NewServeMux()

	// Register handlers
	//	mux.Handle("/", auth.AuthMiddleware(http.HandlerFunc(pages.ServeHomePage)))

	mux.HandleFunc("GET /loginPage", auth.ServeLoginPage)
	mux.HandleFunc("GET /loginForm", auth.ServeLoginForm)
	mux.HandleFunc("POST /login", auth.LoginFormHandler)

	mux.HandleFunc("GET /registerPage", auth.ServeRegistrationPage)
	mux.HandleFunc("GET /registerForm", auth.ServeRegistrationForm)
	mux.HandleFunc("POST /registeruser", auth.RegisterHandler)

	mux.HandleFunc("GET /stripePage", stripe.ServeMembershipPage)
	mux.HandleFunc("GET /stripeForm", stripe.ServeStripeForm)
	mux.HandleFunc("POST /createSession", stripe.HandleCreateCheckoutSession)

	// Serve static files
	fileServer := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))
	mux.Handle("GET /js/", http.StripPrefix("/js/", fileServer))
	mux.Handle("GET /images/", http.StripPrefix("/images/", fileServer))

	mux.HandleFunc("GET /", pages.ServeHomePage)
	mux.HandleFunc("GET /homeContent", pages.ServeHomeContent)

	return mux
}
