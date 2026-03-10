package router

import (
	"net/http"
	"pact/internal/auth"
	"pact/internal/buckets"
	"pact/internal/connections"
	"pact/internal/pages"
	"pact/internal/storage"
	"pact/internal/stripe"
)

func Router() *http.ServeMux {

	mux := http.NewServeMux()

	// home page: page seen by logged in users
	mux.HandleFunc("GET /", auth.AuthMiddleware(pages.ServeBucketsPage))

	// guest page: the page non-logged-in users see
	mux.HandleFunc("GET /description", auth.OptionalAuthMiddleware(pages.ServeDescriptionPage))

	// log in
	mux.HandleFunc("GET /loginPage", pages.ServeLoginPage)
	mux.HandleFunc("POST /login", pages.LoginFormHandler)

	// registration
	mux.HandleFunc("GET /registerPage", pages.ServeRegistrationPage)
	mux.HandleFunc("POST /register", pages.RegisterHandler)

	mux.HandleFunc("GET /logout", auth.Logout)

	// navbars for the different types of user authorization.
	mux.HandleFunc("GET /guestNavbar", pages.ServeGuestNavbar)
	mux.HandleFunc("GET /registeredNavbar", auth.AuthMiddleware(pages.ServeRegisteredNavbar))
	mux.HandleFunc("GET /memberNavbar", auth.AuthMiddleware(pages.ServeMemberNavbar))

	// account page
	mux.HandleFunc("GET /account", auth.AuthMiddleware(pages.ServeAccountPage))
	mux.HandleFunc("DELETE /deleteAccount", auth.AuthMiddleware(pages.DeleteAccountHandler))

	// stripe/membership
	mux.HandleFunc("GET /stripe", auth.AuthMiddleware(stripe.ServeMembershipPage))
	mux.HandleFunc("POST /createSession", auth.AuthMiddleware(stripe.HandleCreateCheckoutSession))

	// buckets/home pages
	mux.HandleFunc("GET /buckets", auth.AuthMiddleware(pages.ServeBucketsPage))
	mux.HandleFunc("GET /bucketsContent", auth.AuthMiddleware(buckets.ServeBucketsContent))
	mux.HandleFunc("GET /home", auth.AuthMiddleware(pages.ServeHomePage))

	// task management (manager)
	mux.HandleFunc("POST /task/create", auth.AuthMiddleware(buckets.HandleCreateTask))
	mux.HandleFunc("POST /task/assign/{task_id}", auth.AuthMiddleware(buckets.HandleAssignSavedTask))
	mux.HandleFunc("DELETE /task/{task_id}", auth.AuthMiddleware(buckets.HandleDeleteTask))
	mux.HandleFunc("DELETE /assigned-task/{assigned_task_id}", auth.AuthMiddleware(buckets.HandleDeleteAssignedTask))
	mux.HandleFunc("PUT /assigned-task/{assigned_task_id}", auth.AuthMiddleware(buckets.HandleEditAssignedTask))
	mux.HandleFunc("POST /reward/create", auth.AuthMiddleware(buckets.HandleCreateReward))
	mux.HandleFunc("PUT /reward/{task_id}", auth.AuthMiddleware(buckets.HandleUpdateReward))
	mux.HandleFunc("DELETE /reward/{task_id}", auth.AuthMiddleware(buckets.HandleDeleteReward))

	// task actions (worker)
	mux.HandleFunc("POST /task/save/{assigned_task_id}", auth.AuthMiddleware(buckets.HandleSaveSubmission))
	mux.HandleFunc("POST /task/submit/{assigned_task_id}", auth.AuthMiddleware(buckets.HandleSubmitTask))
	mux.HandleFunc("POST /reward/purchase/{task_id}", auth.AuthMiddleware(buckets.HandlePurchaseReward))

	// task review (manager)
	mux.HandleFunc("POST /task/approve/{assigned_task_id}", auth.AuthMiddleware(buckets.HandleApproveTask))
	mux.HandleFunc("POST /task/disapprove/{assigned_task_id}", auth.AuthMiddleware(buckets.HandleDisapproveTask))

	// connections
	mux.HandleFunc("GET /connectionsContent", auth.AuthMiddleware(connections.ServeConnectionsContent))
	mux.HandleFunc("POST /createConnectionRequest", auth.AuthMiddleware(connections.HandleCreateConnectionRequest))
	mux.HandleFunc("POST /acceptConnectionRequest/{request_id}", auth.AuthMiddleware(connections.HandleAcceptConnectionRequest))
	mux.HandleFunc("POST /rejectConnectionRequest/{request_id}", auth.AuthMiddleware(connections.HandleRejectConnectionRequest))
	mux.HandleFunc("PUT /updateActiveConnection/{connection_id}/{connection_username}/{connection_role}", auth.AuthMiddleware(connections.HandleUpdateActiveConnection))
	mux.HandleFunc("DELETE /connection/{connection_id}", auth.AuthMiddleware(connections.HandleDeleteConnection))

	// Serve static files
	fileServer := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))
	mux.Handle("GET /js/", http.StripPrefix("/js/", fileServer))
	mux.Handle("GET /images/", http.StripPrefix("/images/", fileServer))

	// Serve uploaded media files with manager authorization
	mux.HandleFunc("GET /uploads/", auth.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userID").(int)
		storage.ServeUploadedFile(w, r, userId)
	}))

	return mux
}
