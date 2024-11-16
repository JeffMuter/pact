package connections

import (
	"fmt"
	"net/http"
	"pact/internal/pages"
)

func ServeConnectionsContent(w http.ResponseWriter, r *http.Request) {
	// should have a user id added in the context of this req here. lets check
	// got get list of requests.

	// go get connections, if existing, if not, need to know.

	data := pages.TemplateData{
		Data: map[string]string{
			"Title": "Connection",
		},
	}
	pages.RenderTemplateFraction(w, "connections", data)
}

func HandleCreateRequest(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "error parsing form", http.StatusBadRequest)
		return
	}
	formEmail := r.FormValue("email")

	userId := r.Context().Value("userID").(int)

	fmt.Println(userId)

	AddRequest(userId, formEmail)
}
