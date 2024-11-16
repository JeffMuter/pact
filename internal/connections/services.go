package connections

import (
	"context"
	"fmt"
	"pact/database"
)

// AddRequest takes in the current users id, and the email they submitted to
// attempt sending a request. If it fails, we do not error, the user does not
// need to know that email is or isnt in our database
func AddRequest(userId int, email string) {

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
