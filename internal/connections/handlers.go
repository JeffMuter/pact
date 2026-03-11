package connections

import (
	"fmt"
	"log"
	"net/http"
	"pact/database"
	"pact/internal/pages"
	"strconv"
)

// ServeConnectionsContent renders the connections page with current connections,
// pending incoming requests, and active connection details.
func ServeConnectionsContent(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	pendingRequestRows, err := getUsersPendingConnectionRequests(r.Context(), userId)
	if err != nil {
		log.Printf("error getting pending requests for user %d: %v", userId, err)
		http.Error(w, "could not load pending requests", http.StatusInternalServerError)
		return
	}

	type PendingRequest struct {
		RequestID int64
		Email     string
		Role      string
	}

	var pendingRequests []PendingRequest
	for _, row := range pendingRequestRows {
		role := "worker"
		if row.SenderID == row.SuggestedManagerID {
			role = "manager"
		}
		pendingRequests = append(pendingRequests, PendingRequest{
			RequestID: row.RequestID,
			Email:     row.Email,
			Role:      role,
		})
	}

	connectionRows, err := getConnectionsByUserId(r.Context(), userId)
	if err != nil {
		log.Printf("error getting connections for user %d: %v", userId, err)
		http.Error(w, "could not load connections", http.StatusInternalServerError)
		return
	}

	type ConnectionDisplay struct {
		ConnectionId int
		Username     string
		Role         string
	}

	var connections []ConnectionDisplay
	for _, row := range connectionRows {
		if userId == int(row.ManagerID) {
			connections = append(connections, ConnectionDisplay{
				ConnectionId: int(row.ConnectionID),
				Username:     row.WorkerUsername,
				Role:         "manager",
			})
		} else if userId == int(row.WorkerID) {
			connections = append(connections, ConnectionDisplay{
				ConnectionId: int(row.ConnectionID),
				Username:     row.ManagerUsername,
				Role:         "worker",
			})
		}
	}

	data := pages.TemplateData{
		Data: map[string]any{
			"Title":                     "Connections",
			"Connections":               connections,
			"PendingConnectionRequests": pendingRequests,
		},
	}

	activeId, activeUsername, activeRole, err := getActiveConnectionDetails(r.Context(), userId)
	if err == nil && activeId > 0 {
		data.Data["ActiveConnectionId"] = activeId
		data.Data["ActiveUserUsername"] = activeUsername
		data.Data["ActiveConnectionRole"] = activeRole
	}

	pages.RenderTemplateFraction(w, "connections", data)
}

// HandleCreateConnectionRequest processes the form to send a new connection request.
func HandleCreateConnectionRequest(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "could not parse form", http.StatusBadRequest)
		return
	}

	formEmail := r.FormValue("email")
	if formEmail == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	senderRole := r.FormValue("senderRole")
	if senderRole != "manager" && senderRole != "worker" {
		http.Error(w, "role must be 'manager' or 'worker'", http.StatusBadRequest)
		return
	}

	userId := r.Context().Value("userID").(int)

	err = CreateConnectionRequest(r.Context(), userId, senderRole, formEmail)
	if err != nil {
		log.Printf("error creating connection request from user %d to %s: %v", userId, formEmail, err)
		http.Error(w, "could not send connection request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<p class="text-emerald-400 font-semibold mt-4">Connection request sent successfully.</p>`)
}

// ServePendingRequestsList renders only the pending connection requests section.
func ServePendingRequestsList(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("userID").(int)
	if userId < 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	pendingRequestRows, err := getUsersPendingConnectionRequests(r.Context(), userId)
	if err != nil {
		log.Printf("error getting pending requests for user %d: %v", userId, err)
		http.Error(w, "could not load pending requests", http.StatusInternalServerError)
		return
	}

	type PendingRequest struct {
		RequestID int64
		Email     string
		Role      string
	}

	var pendingRequests []PendingRequest
	for _, row := range pendingRequestRows {
		role := "worker"
		if row.SenderID == row.SuggestedManagerID {
			role = "manager"
		}
		pendingRequests = append(pendingRequests, PendingRequest{
			RequestID: row.RequestID,
			Email:     row.Email,
			Role:      role,
		})
	}

	data := pages.TemplateData{
		Data: map[string]any{
			"PendingConnectionRequests": pendingRequests,
		},
	}

	pages.RenderTemplateFraction(w, "pendingRequestsList", data)
}

// HandleAcceptConnectionRequest accepts a pending request by its ID — creates
// the connection with correct roles and deactivates the request atomically.
func HandleAcceptConnectionRequest(w http.ResponseWriter, r *http.Request) {
	requestId, err := strconv.ParseInt(r.PathValue("request_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid request id", http.StatusBadRequest)
		return
	}

	userId := r.Context().Value("userID").(int)

	queries := database.GetQueries()
	req, err := queries.GetConnectionRequestById(r.Context(), requestId)
	if err != nil {
		http.Error(w, "connection request not found", http.StatusNotFound)
		return
	}
	if req.ReceiverID != int64(userId) {
		http.Error(w, "not authorized to accept this request", http.StatusForbidden)
		return
	}

	err = acceptConnectionRequest(r.Context(), requestId)
	if err != nil {
		log.Printf("error accepting request %d: %v", requestId, err)
		http.Error(w, "could not accept connection request", http.StatusInternalServerError)
		return
	}

	ServeConnectionsContent(w, r)
}

// HandleRejectConnectionRequest deactivates a pending request without creating a connection.
func HandleRejectConnectionRequest(w http.ResponseWriter, r *http.Request) {
	requestId, err := strconv.ParseInt(r.PathValue("request_id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid request id", http.StatusBadRequest)
		return
	}

	userId := r.Context().Value("userID").(int)

	queries := database.GetQueries()
	req, err := queries.GetConnectionRequestById(r.Context(), requestId)
	if err != nil {
		http.Error(w, "connection request not found", http.StatusNotFound)
		return
	}
	if req.ReceiverID != int64(userId) {
		http.Error(w, "not authorized to reject this request", http.StatusForbidden)
		return
	}

	err = rejectConnectionRequest(r.Context(), requestId)
	if err != nil {
		log.Printf("error rejecting request %d: %v", requestId, err)
		http.Error(w, "could not reject connection request", http.StatusInternalServerError)
		return
	}

	ServeConnectionsContent(w, r)
}

// HandleUpdateActiveConnection sets a connection as the user's active connection
// and re-renders the active connection display.
func HandleUpdateActiveConnection(w http.ResponseWriter, r *http.Request) {
	connectionId, err := strconv.Atoi(r.PathValue("connection_id"))
	if err != nil {
		http.Error(w, "invalid connection id", http.StatusBadRequest)
		return
	}

	connectionRole := r.PathValue("connection_role")
	connectionUsername := r.PathValue("connection_username")

	userId := r.Context().Value("userID").(int)

	err = updateActiveConnection(r.Context(), userId, connectionId)
	if err != nil {
		log.Printf("error updating active connection for user %d: %v", userId, err)
		http.Error(w, "could not update active connection", http.StatusInternalServerError)
		return
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

// HandleDeleteConnection removes a connection and re-renders the connections page.
func HandleDeleteConnection(w http.ResponseWriter, r *http.Request) {
	connectionId, err := strconv.Atoi(r.PathValue("connection_id"))
	if err != nil {
		http.Error(w, "invalid connection id", http.StatusBadRequest)
		return
	}

	userId := r.Context().Value("userID").(int)

	err = deleteConnection(r.Context(), connectionId, userId)
	if err != nil {
		log.Printf("error deleting connection %d for user %d: %v", connectionId, userId, err)
		http.Error(w, "could not delete connection", http.StatusInternalServerError)
		return
	}

	ServeConnectionsContent(w, r)
}
