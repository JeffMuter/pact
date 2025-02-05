package pages

import (
	"context"
	"fmt"
	"net/http"
	"pact/database"
	"pact/internal/auth"

	"golang.org/x/crypto/bcrypt"
)

func ServeDescriptionPage(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Data: map[string]any{
			"Title": "Pact",
		}}
	RenderLayoutTemplate(w, r, "guestPage", data)
}
func ServeDescriptionContent(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Data: map[string]any{
			"Title": "Pact",
		}}
	RenderTemplateFraction(w, "guest", data)
}

func ServeGuestNavbar(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{}
	RenderTemplateFraction(w, "guestNavbar", data)
}

func ServeRegisteredNavbar(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{}
	RenderTemplateFraction(w, "registeredNavbar", data)
}

func ServeMemberNavbar(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{}
	RenderTemplateFraction(w, "memberNavbar", data)
}

func ServeBucketsPage(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Data: map[string]any{
			"Title": "Buckets",
		}}
	RenderLayoutTemplate(w, r, "bucketsPage", data)
}
func ServeBucketsContent(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Data: map[string]any{
			"Title": "Buckets",
		}}
	RenderTemplateFraction(w, "buckets", data)
}
func ServeLoginPage(w http.ResponseWriter, r *http.Request) { // show login form page
	data := TemplateData{
		Data: map[string]any{
			"Heading": "Login",
			"Title":   "Login",
		}}
	fmt.Println("login handler ran")
	RenderLayoutTemplate(w, r, "loginPage", data)
}

func ServeRegistrationPage(w http.ResponseWriter, r *http.Request) { // registration form page
	data := TemplateData{
		Data: map[string]any{
			"Heading": "Registration Page",
			"Title":   "Register",
		}}

	fmt.Println("registerPage handler ran")
	RenderLayoutTemplate(w, r, "registerPage", data)
}

func ServeLoginForm(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Data: map[string]any{
			"Heading": "Login Page",
		}}

	fmt.Println("loginForm handler ran")
	RenderTemplateFraction(w, "loginForm", data)
}

func ServeRegistrationForm(w http.ResponseWriter, r *http.Request) { // registration form page
	data := TemplateData{
		Data: map[string]any{
			"Heading": "Registration Page",
		}}

	fmt.Println("registerForm handler ran")
	RenderTemplateFraction(w, "registerForm", data)
}

func LoginFormHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "error parsing form", http.StatusBadRequest)
		return
	}
	formEmail := r.FormValue("email")
	formPassword := r.FormValue("password")
	user, err := auth.ValidateUsernamePassword(formEmail, formPassword)
	if err != nil {
		fmt.Println("username & password validation failed")
		http.Error(w, fmt.Sprintf("error validating user by email and password: %v", err), http.StatusBadRequest)
		return
	}
	auth.HandleLoginProcedure(w, r, &user)
}

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
	token, err := auth.GenerateToken(uint(userId))
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

func ServeAccountPage(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Data: map[string]any{
			"Heading": "Account Page",
			"Title":   "Account Page",
		},
	}
	RenderLayoutTemplate(w, r, "accountPage", data)
}

func ServeAccountContent(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Data: map[string]any{
			"Heading": "Account Page",
			"Title":   "Account Page",
		},
	}
	RenderTemplateFraction(w, "accountContent", data)
}
