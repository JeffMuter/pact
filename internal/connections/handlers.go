package connections

import (
	"fmt"
	"net/http"
	"pact/database"
	"pact/internal/pages"
	"strconv"
)

// ServeConnectionsContent needs to send data on all existing connections & requests
func ServeConnectionsContent(w http.ResponseWriter, r *http.Request) {
	// should have a user id added in the context of this req here. lets check
	// got get list of requests.

	// Checking for & getting userId in context as prerequisite...
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		fmt.Printf("error: userID not found in ctx, not good... userId: %v\n", userId)
		http.Error(w, "no userID in context of request", http.StatusUnauthorized)
		return
	}

	// get pending requests
	pendingRequestRows, err := getUsersPendingConnectionRequests(userId)
	if len(pendingRequestRows) == 0 {
		// worth seeing for debugging. but no error here.
		fmt.Println("no pending connections...")
	}
	if err != nil {
		http.Error(w, "error getting pending requests: %v\n", http.StatusInternalServerError)
	}

	// prior pending request rows didn't establish what role the requester desired to be. need to assertain
	pendingRequestMap := make(map[database.GetUserPendingRequestsRow]string)

	for _, row := range pendingRequestRows {
		if row.SenderID == row.SuggestedManagerID { // sender wants to be your...
			pendingRequestMap[row] = "manager"
		} else if row.SenderID == row.SuggestedWorkerID {
			pendingRequestMap[row] = "worker"
		} else { // massive booboo
			http.Error(w, "a pending request row sender id didnt match sugg worker or manager ids", http.StatusInternalServerError)
		}
	}

	// we get connections, then make map of usernames to role that the user we see listed has accepted to be.
	connectionRows, err := getConnectionsByUserId(userId)
	if err != nil {
		fmt.Printf("getting all connection for user by the userId failed: %v\n", err)
	}
	if len(connectionRows) == 0 {
		fmt.Println("no connections found for this user")
	}

	connections := []struct {
		ConnectionId int
		Username     string
		Role         string
	}{}

	for _, connectionRow := range connectionRows {
		// figure out if user is manager or worker, set the connection
		if userId == int(connectionRow.ManagerID) {
			connections = append(connections, struct {
				ConnectionId int
				Username     string
				Role         string
			}{
				ConnectionId: int(connectionRow.ConnectionID),
				Username:     connectionRow.WorkerUsername,
				Role:         "worker",
			})
		}
		if userId == int(connectionRow.WorkerID) {
			connections = append(connections, struct {
				ConnectionId int
				Username     string
				Role         string
			}{
				ConnectionId: int(connectionRow.ConnectionID),
				Username:     connectionRow.ManagerUsername,
				Role:         "manager",
			})
		}
	}

	// get active connection details for the person we're connected to. their id, username, and role.
	activeConnectionId, activeConnectionUsername, activeConnectionRole, err := getActiveConnectionDetails(userId)

	data := pages.TemplateData{
		Data: map[string]any{
			"Title":                     "Connection",
			"Connections":               connections,
			"PendingConnectionRequests": pendingRequestMap,
			"ActiveConnectionId":        activeConnectionId,
			"ActiveUserUsername":        activeConnectionUsername,
			"ActiveConnectionRole":      activeConnectionRole,
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

	senderRole := r.FormValue("senderRole")
	if len(senderRole) == 0 {
		fmt.Println("sender role not recieved...")
		http.Error(w, "sender role not recieved...", http.StatusBadRequest)
		return
	}
	if senderRole != "manager" && senderRole != "worker" {
		fmt.Printf("sender role form value is invalid. Role recieved: %s\n", senderRole)
		http.Error(w, "sender role form value is invalid. Role recieved: %s\n", http.StatusBadRequest)
		return
	}

	userId := r.Context().Value("userID").(int)

	err = CreateConnectionRequest(userId, senderRole, formEmail)
	if err != nil {
		fmt.Printf("error creating connection request: %v from userId: %d, and email given: %s\n", err, userId, formEmail)
		http.Error(w, fmt.Sprintf("error creating connection request: %v from userId: %d, and email given: %s\n", err, userId, formEmail), http.StatusBadRequest)
		return
	}
	// no errs, all done
	w.WriteHeader(200)
}

// HandleDeleteConnectionRequest parses the request url to get the sender and receiver id's for the sql query we need to make to delete the connection request
func HandleDeleteConnectionRequest(w http.ResponseWriter, r *http.Request) {
	senderId, err := strconv.Atoi(r.PathValue("sender_id"))
	if err != nil {
		http.Error(w, "error, url sender_id was not an int\n", http.StatusBadRequest)
		return
	}
	receiverId, err := strconv.Atoi(r.PathValue("receiver_id"))
	if err != nil {
		http.Error(w, "error, url receiver_id was not an int\n", http.StatusBadRequest)
		return
	}
	err = deleteConnectionRequest(senderId, receiverId)
	if err != nil {
		msg := fmt.Sprintf("error, deleting connection request failed: %v\n", err)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
}

// HandleCreateConnection using sender & receiver id used from path values to create a new connection in the db.
func HandleCreateConnection(w http.ResponseWriter, r *http.Request) {
	senderId, err := strconv.Atoi(r.PathValue("sender_id"))
	if err != nil {
		fmt.Printf("senderId value not a number in create-connection request: %v\n", err)
	}

	receiverId, err := strconv.Atoi(r.PathValue("receiver_id"))
	if err != nil {
		fmt.Printf("receiverId non a number from create connection request: %v\n", err)
	}

	err = createConnection(senderId, receiverId)
	if err != nil {
		fmt.Printf("problem in creating connection: %v\n", err)
	}
	w.WriteHeader(200)
}

func HandleUpdateActiveConnection(w http.ResponseWriter, r *http.Request) {
	connectionId, err := strconv.Atoi(r.PathValue("connection_id"))
	if err != nil {
		fmt.Printf("err, connection Id from template data not an integer: %v", err)
		http.Error(w, "err, connection Id from template data not an integer.", http.StatusBadRequest)
	}

	connectionRole := r.PathValue("connection_role")

	connectionUsername := r.PathValue("connection_username")

	// update the users table with connectionId
	err = updateActiveConnection(connectionId)
	if err != nil {
		fmt.Printf("error updating active connection: %v\n", err)
		http.Error(w, "error updating active connection: %v\n", http.StatusInternalServerError)
	}

	data := pages.TemplateData{
		Data: map[string]any{
			"ActiveConnectionId":   connectionId,
			"ActiveUserUsername":   connectionUsername,
			"ActiveConnectionRole": connectionRole,
		},
	}
	pages.RenderTemplateFraction(w, "activeConnection", data)
}
