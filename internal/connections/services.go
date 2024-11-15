package connections

import (
	"context"
	"fmt"
	"pact/database"
)

func AddRequest(userId int, email string) {
	// try to get the user based on the email.
	// if cant find the user, idc just log it
	// add the new details about both users to the connection_requests table

	queries := database.GetQueries()
	ctx := context.Background()

	user, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		fmt.Println("couldnt get user by email")
		return
	}

	var args database.CreateRequestParams
	args.SenderID = int64(userId)
	args.RecieverID = user.UserID

	// add this new req to db.
	queries.CreateRequest(ctx, args)

}
