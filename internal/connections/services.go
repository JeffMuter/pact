package connections

import (
	"context"
	"fmt"
	"pact/database"
)

// AddRequest takes in the current users id, and the email they submitted to
// attempt sending a request. If it fails, we do not error, the user does not
// need to know that email is or isnt in our database
func CreateConnectionRequest(userId int, email string) error {

	queries := database.GetQueries()
	ctx := context.Background()

	user, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		fmt.Println("couldnt get user by email")
		return fmt.Errorf("error couldnt find existing user via their email: %w\n", err)
	}

	var args database.CreateRequestParams
	args.SenderID = int64(userId)
	args.RecieverID = user.UserID
	args.

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

func createConnection(senderId, recieverId int) error {
	fmt.Println("begin to create connection from user ids")

	var managerUserId, workerUserId int

	queries := database.GetQueries()
	ctx := context.Background()

	// get sender and reciever user values from the db from the ids recieved from params
	senderUser, err := queries.GetUserById(ctx, int64(senderId))
	if err != nil {
		return fmt.Errorf("could not find connection request sender by their id in db: %w", err)
	}

	recieverUser, err := queries.GetUserById(ctx, int64(recieverId))
	if err != nil {
		return fmt.Errorf("user could not be found in db from the connection request recievers id: %w", err)
	}

	// set manager & worker id values, and do some error detection
	if senderUser.Role == recieverUser.Role {
		return fmt.Errorf("connection request sender & reciever are the same role. A role combination we do no support.")
	}
	if senderUser.Role == "manager" {
		managerUserId = senderId
	} else if senderUser.Role == "worker" {
		workerUserId = senderId
	} else {
		return fmt.Errorf("sender user was found, but their role type is invalid...")
	}

	if recieverUser.Role == "manager" && managerUserId == 0 {
		managerUserId = recieverId
	} else if recieverUser.Role == "worker" && workerUserId == 0 {
		workerUserId = recieverId
	} else {
		return fmt.Errorf("reciever and sender users were found. reciever role was probably invalid")
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
func getConnectionsByUserId(userId int) ([]database.Connection, error) {

	queries := database.GetQueries()
	ctx := context.Background()

	rows, err := queries.GetConnectionsById(ctx, int64(userId))
	if err != nil {
		return nil, fmt.Errorf("could not get connections by userId: %w", err)
	}

	return rows, nil
}
