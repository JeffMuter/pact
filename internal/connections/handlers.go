package connections

import (
	"fmt"
	"net/http"
	"pact/internal/pages"
)

func ServeConnectionsContent(w http.ResponseWriter, r *http.Request) {
	// should have a user id added in the context of this req here. lets check
	// got get list of requests.

	// connection requests added to data here...
	userId := r.Context().Value("userID").(int)
	getUsersPendingConnectionRequests(userId)

	data := pages.TemplateData{
		Data: map[string]string{
			"Title": "Connection",
		},
	}
	pages.RenderTemplateFraction(w, "connections", data)
}

func HandleCreateConnectionRequest(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "error parsing form", http.StatusBadRequest)
		return
	}
	formEmail := r.FormValue("email")
	if len(formEmail) == 0 {
		http.Error(w, "Email input was empty", http.StatusBadRequest)
		return
	}

	userId := r.Context().Value("userID").(int)

	err = CreateConnectionRequest(userId, formEmail)
	if err != nil {
		fmt.Printf("error creating connection request: %v\n", err)
	}
}
