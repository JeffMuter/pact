package pages

import (
	"context"
	"fmt"
	"net/http"
	"pact/database"
	"pact/internal/auth"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func ServeDescriptionPage(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Data: map[string]any{
			"Title": "Pact - Description",
		}}
	
	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		RenderTemplateFraction(w, "description", data)
		return
	}
	
	RenderLayoutTemplate(w, r, "descriptionPage", data)
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
	
	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		RenderTemplateFraction(w, "buckets", data)
		return
	}
	
	RenderLayoutTemplate(w, r, "bucketsPage", data)
}
func ServeLoginPage(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Data: map[string]any{
			"Heading": "Login",
			"Title":   "Login",
		}}
	
	fmt.Println("login handler ran")
	
	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		RenderTemplateFraction(w, "loginForm", data)
		return
	}
	
	RenderLayoutTemplate(w, r, "loginPage", data)
}

func ServeRegistrationPage(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Data: map[string]any{
			"Heading": "Registration Page",
			"Title":   "Register",
		}}

	fmt.Println("registerPage handler ran")
	
	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		RenderTemplateFraction(w, "registerForm", data)
		return
	}

	RenderLayoutTemplate(w, r, "registerPage", data)
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
	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// response successful
	token, err := auth.GenerateToken(uint(userId))
	if err != nil {
		fmt.Printf("error generating token (userId: %d): %v\n", userId, err)
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    token,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
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
	
	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		RenderTemplateFraction(w, "accountContent", data)
		return
	}
	
	RenderLayoutTemplate(w, r, "accountPage", data)
}

func ServeStripePage(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Data: map[string]any{
			"Title": "Stripe - Membership",
		},
	}
	
	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		RenderTemplateFraction(w, "stripeForm", data)
		return
	}
	
	RenderLayoutTemplate(w, r, "stripePage", data)
}

func ServeHomePage(w http.ResponseWriter, r *http.Request) {
	data := TemplateData{
		Data: map[string]any{
			"Title": "Home",
		}}
	
	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		RenderTemplateFraction(w, "buckets", data)
		return
	}
	
	RenderLayoutTemplate(w, r, "bucketsPage", data)
}

func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	queries := database.GetQueries()
	ctx := context.Background()

	userId, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "error getting user ID from context", http.StatusUnauthorized)
		return
	}

	err := queries.DeleteUser(ctx, int64(userId))
	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting user: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Println("user account deleted successfully")

	// Log out the user by expiring their cookie
	cookie := &http.Cookie{
		Name:     "Bearer",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)

	// Redirect to login page using HX-Redirect for HTMX requests
	w.Header().Set("HX-Redirect", "/loginPage")
}
