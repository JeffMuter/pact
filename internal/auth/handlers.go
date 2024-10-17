package auth

import (
	"context"
	"fmt"
	"net/http"
	"pact/database"
	"pact/internal/pages"

	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("registration in progress...")
	queries := database.GetQueries()
	ctx := context.Background()

	var user database.CreateUserParams

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "error parsing form", http.StatusBadRequest)
		return
	}

	// set User fields
	user.Email = r.FormValue("email")
	user.Username = r.FormValue("username")
	user.Role = r.FormValue("role")
	user.PasswordHash = r.FormValue("password")
	if user.Email == "" || user.PasswordHash == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// get hashed version of password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	user.PasswordHash = string(hashedPassword)

	// create user in db
	userId, err := queries.CreateUser(ctx, user)

	// response successful
	token, err := GenerateToken(uint(userId))
	if err != nil {
		fmt.Printf("error generating token (userId: %d): %v\n", userId, err)
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer token",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func LoginFormHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "error parsing form", http.StatusBadRequest)
		return
	}
	formEmail := r.FormValue("email")
	formPassword := r.FormValue("password")
	user, err := validateUsernamePassword(formEmail, formPassword)
	if err != nil {
		fmt.Println("username & password validation failed")
		http.Error(w, fmt.Sprintf("error validating user by email and password: %v", err), http.StatusBadRequest)
		return
	}
	handleLoginProcedure(w, r, &user)
}

func handleLoginProcedure(w http.ResponseWriter, r *http.Request, user *database.User) {
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

func ServeLoginPage(w http.ResponseWriter, r *http.Request) { // show login form page
	data := pages.TemplateData{
		Data: map[string]string{
			"Heading": "Login",
			"Title":   "Login",
		}}
	fmt.Println("login handler ran")
	pages.RenderLayoutTemplate(w, r, "loginPage", data)
}

func ServeRegistrationPage(w http.ResponseWriter, r *http.Request) { // registration form page
	data := pages.TemplateData{
		Data: map[string]string{
			"Heading": "Registration Page",
			"Title":   "Register",
		}}

	fmt.Println("registerPage handler ran")
	pages.RenderLayoutTemplate(w, r, "registerPage", data)
}

func ServeLoginForm(w http.ResponseWriter, r *http.Request) {
	data := pages.TemplateData{
		Data: map[string]string{
			"Heading": "Login Page",
		}}

	fmt.Println("loginForm handler ran")
	pages.RenderTemplateFraction(w, "loginForm", data)
}

func ServeRegistrationForm(w http.ResponseWriter, r *http.Request) { // registration form page
	data := pages.TemplateData{
		Data: map[string]string{
			"Heading": "Registration Page",
		}}

	fmt.Println("registerForm handler ran")
	pages.RenderTemplateFraction(w, "registerForm", data)
}
