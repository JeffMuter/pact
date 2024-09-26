package auth

import (
	"fmt"
	"net/http"
	"pact/internal/db"
	"pact/internal/pages"
	"pact/internal/user"

	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("registration in progress...")
	db := db.GetDB()

	var user user.User

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "error parsing form", http.StatusBadRequest)
		return
	}

	user.Email = r.FormValue("email")
	user.Username = r.FormValue("username")
	user.Role = r.FormValue("role")
	user.Password = r.FormValue("password")
	if user.Email == "" || user.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	// create sql statement
	query := `INSERT INTO users (email, username, role, password_hash) VALUES ($1, $2, $3, $4)`

	_, err = db.Exec(query, user.Email, user.Username, user.Role, user.Password)
	if err != nil {
		fmt.Println("error excecuting query: %w", err)
		http.Error(w, "Error executing query", http.StatusInternalServerError)
		return
	}
	// response successful

	err = SetSession(user.Email, w)
	if err != nil {
		fmt.Printf("error setting session: %v", err)
		http.Error(w, "Failed to set session", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func LoginFormHandler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "error parsing form", http.StatusBadRequest)
		return
	}

	formEmail := r.FormValue("email")
	formPassword := r.FormValue("password")

	err = validateUsernamePassword(formEmail, formPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	err = SetSession(formEmail, w)
	if err != nil {
		http.Error(w, "Failed to set session", http.StatusBadRequest)
		fmt.Printf("error setting session: %v", err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
	http.Error(w, "Username or password incorrect... Try again.", http.StatusBadRequest)
}

func ServeLoginPage(w http.ResponseWriter, r *http.Request) { // show login form page
	data := pages.TemplateData{
		Data: map[string]string{
			"Heading": "Login",
			"Title":   "Login",
		}}
	fmt.Println("login handler ran")
	pages.RenderLayoutTemplate(w, "loginPage", data)
}

func ServeRegistrationPage(w http.ResponseWriter, r *http.Request) { // registration form page
	data := pages.TemplateData{
		Data: map[string]string{
			"Heading": "Registration Page",
			"Title":   "Register",
		}}

	fmt.Println("registerPage handler ran")
	pages.RenderLayoutTemplate(w, "registerPage", data)
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
