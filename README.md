go run main.go starts the web server on port 8080 --> localhost:8080/
there is a nix-shell here, although not fleshed out properly at the latest write of the readme.

this project requires authentication, in dev, you can create accounts, requires  accessing the db to turn accounts into active members with subscriptions. 

users can send friend requests by email identification, requesting that user accept the role as either manager or worker.

Managers assign tasks, create and save lists of tasks with different points as a reward for the worker. Can also create rewards, and punishments for when tasks are not submitted correctly on time. Tasks can have requirements from time requirements, required to submit an image, or texts with word counts. Users upload said text, image, or video files to aws S3 bucket, where it's stored for 24 hours after review. Managers also review completed task submissions, and either approve, or reject them. 

Workers can view assigned tasks, and move them to an in-review phase for the manager to review. Workers can submit image, text, or video to the task in order to mark the task as complete. Workers can also collect points by completing tasks, then use the points to purchase 'rewards', another kind of task from the manager.

User can subscribe via Stripe.

Users can have multiple 'connections', but only one connection can be selected as the 'active' connection. As a user may be the manager, or the worker for different connections. Users can have 2 connections to the same other user, one where they're the manager, the other where they're the worker.

Tools:

Golang,
HTMX,
TailwindCSS,
DaisyUI,
SQLite,
JavaScript,
SQLC,
