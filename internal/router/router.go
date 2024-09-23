package router

import (
	"net/http"
	"pact/internal/auth"
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

	return mux

}
