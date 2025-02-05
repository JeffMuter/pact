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

	recieverUser, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		fmt.Println("couldnt get user by email")
		return fmt.Errorf("error couldnt find existing user via their email: %w\n", err)
	}

	var suggestedManager, suggestedWorker int

	if senderRole == "manager" {
		suggestedManager = userId
		suggestedWorker = int(recieverUser.UserID)
	}
	if senderRole == "worker" {
		suggestedManager = int(recieverUser.UserID)
		suggestedWorker = userId
	}

	var args database.CreateRequestParams
	args.SenderID = int64(userId)
	args.RecieverID = recieverUser.UserID
	args.SuggestedManagerID = int64(suggestedManager)
	args.SuggestedWorkerID = int64(suggestedWorker)

	// add this new req to db.
	err = queries.CreateRequest(ctx, args)
	if err != nil {
		fmt.Printf("error creating request from email senderID: %d, recieverId: %d\n", args.SenderID, args.RecieverID)
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

// deleteConnectionRequest uses senderId and recieverId to delete any pending requests matching the sender and reciever fields of connection_requests table
func deleteConnectionRequest(senderId, recieverId int) error {
	fmt.Printf("senderId: %v... recieverId: %v...\n", senderId, recieverId)

	queries := database.GetQueries()
	ctx := context.Background()

	var args database.DeleteConnectionRequestByUserIdsParams
	args.SenderID = int64(senderId)
	args.RecieverID = int64(recieverId)

	err := queries.DeleteConnectionRequestByUserIds(ctx, args)
	if err != nil {
		return fmt.Errorf("deleting connection request didnt go so good: %w\n", err)
	}

	return nil
}

// createConnection uses a sender and reciever id to create a new connection row in the connections table by gathering info using those details to gather info from other tables
func createConnection(senderRole string, senderId, recieverId int) error {
	// TODO: my understanding is, the were currently sending 2 userid, without knowing which one is the manager and which is the worker in this connection.

	fmt.Println("begin to create connection from user ids")

	var managerUserId, workerUserId int

	queries := database.GetQueries()
	ctx := context.Background()

	if senderRole == "manager" {

	} else if senderRole == "worker" {

	} else {
		return fmt.Errorf("senderRole neither manager or worker.")
	}

	var args database.CreateConnectionParams
	args.ManagerID = int64(managerUserId)
	args.WorkerID = int64(workerUserId)

	// create connection
	err = queries.CreateConnection(ctx, args)
	if err != nil {
		return fmt.Errorf("create connection query failed: %w", err)
	}

	// this code wont run unless connection creation success
	// delete connection request
	err = deleteConnectionRequest(senderId, recieverId)
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
		fmt.Errorf("update active connection failed with id: %d. error: %w", connectionId, err)
	}

	return nil
}

func getActiveConnectionDetails(userId int) (string, string, error) {
	var acUsername, acRole string

	queries := database.GetQueries()
	ctx := context.Background()

	row, err := queries.GetActiveConnectionDetails(ctx, int64(userId))
	if err != nil {
		fmt.Printf("could not get active connection details from db using userId: %s, err: %v\n", userId, err)
	}

	// using the manager and worker id from the returned row, figure out this user's role in the active connection
	managerId, workerId := int(row.ManagerID), int(row.WorkerID)

	if userId == workerId {
		acRole = "manager"
		acRole, err = queries.GetUsernameByUserId(ctx, int64(managerId))
	} else if userId == managerId {
		acRole = "worker"
		acRole, err = queries.GetUsernameByUserId(ctx, int64(workerId))
	}

	return acUsername, acRole, nil
}
