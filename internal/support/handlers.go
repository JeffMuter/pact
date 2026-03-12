package support

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"pact/database"
	"pact/internal/pages"
)

// ServeSupportPage renders the full support page
func ServeSupportPage(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		http.Error(w, "userID not found in context", http.StatusUnauthorized)
		return
	}

	queries := database.GetQueries()
	ctx := r.Context()

	// Fetch user email
	user, err := queries.GetUserById(ctx, int64(userId))
	if err != nil {
		http.Error(w, "Failed to load user data", http.StatusInternalServerError)
		return
	}

	data := pages.TemplateData{
		Data: map[string]any{
			"Email": user.Email,
		},
	}

	pages.RenderLayoutTemplate(w, r, "supportPage", data)
}

// HandleSupportTicketSubmission handles the form submission for support tickets
func HandleSupportTicketSubmission(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userID").(int)
		if userId < 1 {
			http.Error(w, "userID not found in context", http.StatusUnauthorized)
			return
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		issueDescription := r.FormValue("issue_description")

		// Validate inputs
		if email == "" || issueDescription == "" {
			http.Error(w, "Email and issue description are required", http.StatusBadRequest)
			return
		}

		// Create support ticket
		err := CreateSupportTicket(db, userId, email, issueDescription)
		if err != nil {
			log.Printf("Failed to create support ticket: %v", err)
			http.Error(w, "Failed to submit support ticket", http.StatusInternalServerError)
			return
		}

		// Return success message
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `
			<div class="alert alert-success">
				<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
				<span>Support ticket submitted successfully! We'll contact you via email.</span>
			</div>
		`)
	}
}
