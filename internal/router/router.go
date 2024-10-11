package router

import (
	"net/http"
	"pact/internal/auth"
	"pact/internal/pages"
	"pact/internal/relationships"
	"pact/internal/stripe"
)

func Router() *http.ServeMux {

	mux := http.NewServeMux()

	// home page: page seen by logged in users
	mux.HandleFunc("GET /", auth.AuthMiddleware(pages.ServeBucketsPage))
	mux.HandleFunc("GET /bucketContent", auth.AuthMiddleware(pages.ServeBucketsContent))

	// guest page: the page non-logged-in users see
	mux.HandleFunc("GET /description", pages.ServeDescriptionPage)
	mux.HandleFunc("GET /descriptionContent", pages.ServeDescriptionContent)

	// log in
	mux.HandleFunc("GET /loginPage", auth.ServeLoginPage)
	mux.HandleFunc("GET /loginForm", auth.ServeLoginForm)
	mux.HandleFunc("POST /login", auth.LoginFormHandler)

	// registration
	mux.HandleFunc("GET /registerPage", auth.ServeRegistrationPage)
	mux.HandleFunc("GET /registerForm", auth.ServeRegistrationForm)
	mux.HandleFunc("POST /register", auth.RegisterHandler)

	// navbars for the different types of user authorization.
	mux.HandleFunc("GET /guestNavbar", pages.ServeGuestNavbar)
	mux.HandleFunc("GET /loggedInNavbar", auth.AuthMiddleware(pages.ServeLoggedInNavbar))
	mux.HandleFunc("GET /memberNavbar", auth.AuthMiddleware(pages.ServeMemberNavbar))

	// stripe
	mux.HandleFunc("GET /stripePage", auth.AuthMiddleware(stripe.ServeMembershipPage))
	mux.HandleFunc("GET /stripeForm", auth.AuthMiddleware(stripe.ServeStripeForm))
	mux.HandleFunc("POST /createSession", auth.AuthMiddleware(stripe.HandleCreateCheckoutSession))

	// relationships
	mux.HandleFunc("GET /relationshipContent", auth.AuthMiddleware(relationships.ServePageContent))

	// Serve static files
	fileServer := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))
	mux.Handle("GET /js/", http.StripPrefix("/js/", fileServer))
	mux.Handle("GET /images/", http.StripPrefix("/images/", fileServer))

	return mux
}
