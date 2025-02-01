package connections

import (
	"fmt"
	"net/http"
	"pact/internal/pages"
	"strconv"
)

func ServeConnectionsContent(w http.ResponseWriter, r *http.Request) {
	// should have a user id added in the context of this req here. lets check
	// got get list of requests.

	// connection requests added to data here...
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		fmt.Printf("userID not found in ctx, not good... userId: %v\n", userId)
		http.Error(w, "no userID in context of request", http.StatusUnauthorized)
		return
	}
	rows, err := getUsersPendingConnectionRequests(userId)
	if len(rows) == 0 {
		// worth seeing for debugging. but no error here.
		fmt.Println("no pending connections...")
	}
	if err != nil {
		http.Error(w, "error getting pending requests: %v\n", http.StatusInternalServerError)
	}

	data := pages.TemplateData{
		Data: map[string]any{
			"Title":                     "Connection",
			"PendingConnectionRequests": rows,
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
		return
	}
	// no errs, all done
	w.WriteHeader(200)
}

// HandleDeleteConnectionRequest parses the request url to get the sender and reciever id's for the sql query we need to make to delete the connection request
func HandleDeleteConnectionRequest(w http.ResponseWriter, r *http.Request) {
	senderId, err := strconv.Atoi(r.PathValue("sender_id"))
	if err != nil {
		http.Error(w, "error, url sender_id was not an int\n", http.StatusBadRequest)
		return
	}
	recieverId, err := strconv.Atoi(r.PathValue("reciever_id"))
	if err != nil {
		http.Error(w, "error, url reciever_id was not an int\n", http.StatusBadRequest)
		return
	}
	err = deleteConnectionRequest(senderId, recieverId)
	if err != nil {
		msg := fmt.Sprintf("error, deleting connection request failed: %v\n", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
}

func HandleCreateConnection(w http.ResponseWriter, r *http.Request) {
	senderId, err := strconv.Atoi(r.PathValue("sender_id"))
	if err != nil {
		fmt.Printf("senderId value not a number in create-connection request: %v\n", err)
	}
	recieverId, err := strconv.Atoi(r.PathValue("reciever_id"))
	if err != nil {
		fmt.Printf("recieverId non a number from create connection request: %v\n", err)
	}

	err = createConnection(senderId, recieverId)
	if err != nil {
		fmt.Printf("problem in creating connection: %v\n", err)
	}
	w.WriteHeader(200)
}
