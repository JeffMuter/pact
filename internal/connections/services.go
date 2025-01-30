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

	// add this new req to db.
	err = queries.CreateRequest(ctx, args)
	if err != nil {
		fmt.Println("error creating request")
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
	queries := database.GetQueries()
	ctx := context.Background()

	// TODO: use generated delete here!

	return nil
}
