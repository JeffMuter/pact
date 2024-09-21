package auth

import (
	"net/http"
	"pact/internal/db"
	"pact/internal/pages"
	"pact/internal/user"

	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {

	db := db.GetDB()

	var user user.User

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "error parsing form", http.StatusBadRequest)
		return
	}

	user.Email = r.FormValue("email")
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
	query, err := db.Prepare("INSERT INTO users (email, password) VALUES ($1, $2)")
	if err != nil {
		http.Error(w, "Error preparing query", http.StatusInternalServerError)
		return
	}
	defer query.Close()

	_, err = query.Exec(user.Email, user.Password)
	if err != nil {
		http.Error(w, "Error executing query", http.StatusInternalServerError)
		return
	}
	// response successful

	err = SetSession(user.Email, w)
	if err != nil {
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
		// should log this
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
	pages.RenderTemplate(w, "loginForm.html", data)
}

func ServeRegistrationPage(w http.ResponseWriter, r *http.Request) { // registration form page
	data := pages.TemplateData{
		Data: map[string]string{
			"Heading": "Registration Page",
			"Title":   "Register",
		}}
	pages.RenderTemplate(w, "registerForm.html", data)
}
