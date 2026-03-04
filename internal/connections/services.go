package connections

import (
	"context"
	"fmt"
	"pact/database"
)

// AddRequest takes in the current users id, and the email they submitted to
// attempt sending a request. If it fails, we do not error, the user does not
// need to know that email is or isnt in our database
func CreateConnectionRequest(userId int, senderRole, email string) error {

	queries := database.GetQueries()
	ctx := context.Background()

	receiverUser, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		fmt.Println("couldnt get user by email")
		return fmt.Errorf("error couldnt find existing user via their email: %w\n", err)
	}

	var suggestedManager, suggestedWorker int

	if senderRole == "manager" {
		suggestedManager = userId
		suggestedWorker = int(receiverUser.UserID)
	}
	if senderRole == "worker" {
		suggestedManager = int(receiverUser.UserID)
		suggestedWorker = userId
	}

	var args database.CreateRequestParams
	args.SenderID = int64(userId)
	args.ReceiverID = receiverUser.UserID
	args.SuggestedManagerID = int64(suggestedManager)
	args.SuggestedWorkerID = int64(suggestedWorker)

	// add this new req to db.
	err = queries.CreateRequest(ctx, args)
	if err != nil {
		fmt.Printf("error creating request from email senderID: %d, receiverId: %d\n", args.SenderID, args.ReceiverID)
		return fmt.Errorf("error couldnt find existing user via their email: %w\n", err)
	}

	return nil
}

// getUsersPendingConnectionRequests fetches all the information of open/ non accepted or declined connections to return
func getUsersPendingConnectionRequests(userId int) ([]database.GetUserPendingRequestsRow, error) {
	queries := database.GetQueries()
	ctx := context.Background()

	pendingRequestData, err := queries.GetUserPendingRequests(ctx, int64(userId))
	if err != nil {
		return pendingRequestData, fmt.Errorf("error getting pending request info by the given userId: %d, %w\n", userId, err)
	}
	for _, row := range pendingRequestData {
		fmt.Printf("rowdata: %v\n", row)
	}

	return pendingRequestData, nil
}

// deleteConnectionRequest uses senderId and receiverId to delete any pending requests matching the sender and receiver fields of connection_requests table
func deleteConnectionRequest(senderId, receiverId int) error {
	fmt.Printf("senderId: %v... receiverId: %v...\n", senderId, receiverId)

	queries := database.GetQueries()
	ctx := context.Background()

	var args database.DeleteConnectionRequestByUserIdsParams
	args.SenderID = int64(senderId)
	args.ReceiverID = int64(receiverId)

	err := queries.DeleteConnectionRequestByUserIds(ctx, args)
	if err != nil {
		return fmt.Errorf("deleting connection request didnt go so good: %w\n", err)
	}

	return nil
}

// createConnection is used to create a new connection in the db
func createConnection(managerId, workerId int) error {
	fmt.Println("begin to create connection from user ids")

	queries := database.GetQueries()
	ctx := context.Background()

	var args database.CreateConnectionParams
	args.ManagerID = int64(managerId)
	args.WorkerID = int64(workerId)

	// create connection
	err := queries.CreateConnection(ctx, args)
	if err != nil {
		return fmt.Errorf("create connection query failed: %w", err)
	}

	// this code wont run unless connection creation success
	// delete connection request
	err = deleteConnectionRequest(managerId, workerId)
	if err != nil {
		fmt.Printf("deleting connection request failed, after connection successfully created: %v\n", err)
	}
	fmt.Println("connection created. request successfully deleted")

	return nil
}

// getConnectionsById takes a users id, and returns all active connections for this user
func getConnectionsByUserId(userId int) ([]database.GetConnectionsByIdRow, error) {

	queries := database.GetQueries()
	ctx := context.Background()

	rows, err := queries.GetConnectionsById(ctx, int64(userId))
	if err != nil {
		return nil, fmt.Errorf("could not get connections by userId: %w", err)
	}

	return rows, nil
}

// updateActiveConnection using connectionId and the connectionRole to update users user on the user table, update the connectionId
func updateActiveConnection(connectionId int) error {
	queries := database.GetQueries()
	ctx := context.Background()

	err := queries.UpdateActiveConnection(ctx, int64(connectionId))
	if err != nil {
		return fmt.Errorf("update active connection failed with id: %d. error: %w", connectionId, err)
	}

	return nil
}

// getActiveConnectionDetails takes current user Id, uses that to get details on the user of the current connection. their Id, their role, and username
func getActiveConnectionDetails(userId int) (int, string, string, error) {
	// data about person current user's connected to
	var acUsername, acRole string
	var acId int

	queries := database.GetQueries()
	ctx := context.Background()

	// start by getting the active
	var params database.GetActiveConnectionUserDetailsParams
	params.ManagerID = int64(userId)
	params.ManagerID_2 = int64(userId)
	params.UserID = int64(userId)
	
	row, err := queries.GetActiveConnectionUserDetails(ctx, params)
	if err != nil {
		return acId, acUsername, acRole, fmt.Errorf("error: getting active connection details from db, %w", err)
	}

	userIDInt64, ok := row.UserID.(int64)
	if !ok {
		return acId, acUsername, acRole, fmt.Errorf("error: userID type assertion failed")
	}
	acId = int(userIDInt64)
	acUsername = row.Username
	acRole = row.Role

	return acId, acUsername, acRole, nil
}
