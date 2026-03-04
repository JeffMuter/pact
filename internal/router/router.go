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

	// guest page: the page non-logged-in users see
	mux.HandleFunc("GET /description", auth.OptionalAuthMiddleware(pages.ServeDescriptionPage))

	// log in
	mux.HandleFunc("GET /loginPage", pages.ServeLoginPage)
	mux.HandleFunc("POST /login", pages.LoginFormHandler)

	// registration
	mux.HandleFunc("GET /registerPage", pages.ServeRegistrationPage)
	mux.HandleFunc("POST /register", pages.RegisterHandler)

	mux.HandleFunc("GET /logout", auth.Logout)

	// navbars for the different types of user authorization.
	mux.HandleFunc("GET /guestNavbar", pages.ServeGuestNavbar)
	mux.HandleFunc("GET /registeredNavbar", auth.AuthMiddleware(pages.ServeRegisteredNavbar))
	mux.HandleFunc("GET /memberNavbar", auth.AuthMiddleware(pages.ServeMemberNavbar))

	// account page
	mux.HandleFunc("GET /account", auth.AuthMiddleware(pages.ServeAccountPage))
	mux.HandleFunc("DELETE /deleteAccount", auth.AuthMiddleware(pages.DeleteAccountHandler))

	// stripe/membership
	mux.HandleFunc("GET /stripe", auth.AuthMiddleware(stripe.ServeMembershipPage))
	mux.HandleFunc("POST /createSession", auth.AuthMiddleware(stripe.HandleCreateCheckoutSession))

	// buckets/home pages
	mux.HandleFunc("GET /buckets", auth.AuthMiddleware(pages.ServeBucketsPage))
	mux.HandleFunc("GET /home", auth.AuthMiddleware(pages.ServeHomePage))

	// connections
	mux.HandleFunc("GET /connectionsContent", auth.AuthMiddleware(connections.ServeConnectionsContent))
	mux.HandleFunc("POST /createConnectionRequest", auth.AuthMiddleware(connections.HandleCreateConnectionRequest))
	mux.HandleFunc("POST /acceptConnectionRequest/{request_id}", auth.AuthMiddleware(connections.HandleAcceptConnectionRequest))
	mux.HandleFunc("POST /rejectConnectionRequest/{request_id}", auth.AuthMiddleware(connections.HandleRejectConnectionRequest))
	mux.HandleFunc("PUT /updateActiveConnection/{connection_id}/{connection_username}/{connection_role}", auth.AuthMiddleware(connections.HandleUpdateActiveConnection))
	mux.HandleFunc("DELETE /connection/{connection_id}", auth.AuthMiddleware(connections.HandleDeleteConnection))

	// Serve static files
	fileServer := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))
	mux.Handle("GET /js/", http.StripPrefix("/js/", fileServer))
	mux.Handle("GET /images/", http.StripPrefix("/images/", fileServer))

	return mux
}
