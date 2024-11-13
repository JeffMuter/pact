package router

import (
	"net/http"
	"pact/internal/auth"
	"pact/internal/connections"
	"pact/internal/pages"
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
	mux.HandleFunc("GET /loginPage", pages.ServeLoginPage)
	mux.HandleFunc("GET /loginForm", pages.ServeLoginForm)
	mux.HandleFunc("POST /login", pages.LoginFormHandler)

	// registration
	mux.HandleFunc("GET /registerPage", pages.ServeRegistrationPage)
	mux.HandleFunc("GET /registerForm", pages.ServeRegistrationForm)
	mux.HandleFunc("POST /register", pages.RegisterHandler)

	mux.HandleFunc("GET /logout", auth.Logout)

	// navbars for the different types of user authorization.
	mux.HandleFunc("GET /guestNavbar", pages.ServeGuestNavbar)
	mux.HandleFunc("GET /registeredNavbar", auth.AuthMiddleware(pages.ServeRegisteredNavbar))
	mux.HandleFunc("GET /memberNavbar", auth.AuthMiddleware(pages.ServeMemberNavbar))

	// account page
	mux.HandleFunc("GET /accountPage", auth.AuthMiddleware(pages.ServeAccountPage))
	mux.HandleFunc("GET /accountContent", auth.AuthMiddleware(pages.ServeAccountContent))

	// stripe
	mux.HandleFunc("GET /stripePage", auth.AuthMiddleware(stripe.ServeMembershipPage))
	mux.HandleFunc("GET /stripeForm", auth.AuthMiddleware(stripe.ServeStripeForm))
	mux.HandleFunc("POST /createSession", auth.AuthMiddleware(stripe.HandleCreateCheckoutSession))

	// connections
	mux.HandleFunc("GET /connectionsContent", auth.AuthMiddleware(connections.ServeConnectionsContent))
	mux.HandleFunc("POST /addConnectionRequest", auth.AuthMiddleware(connections.AddRequest))

	// Serve static files
	fileServer := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))
	mux.Handle("GET /js/", http.StripPrefix("/js/", fileServer))
	mux.Handle("GET /images/", http.StripPrefix("/images/", fileServer))

	return mux
}
