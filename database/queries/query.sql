-- name: GetAllUsers :many
select * from users;

-- name: GetUserById :one
SELECT * from users WHERE user_id = ?;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;

-- name: UserIsMemberById :one
SELECT is_member FROM users WHERE user_id = ?;

-- name: CreateUser :one
INSERT INTO users (email, username, role, password_hash) VALUES (?, ?, ?, ?) returning user_id;

-- name: CreateSession :exec
INSERT INTO sessions(user_id, token, created_at, expires_at) VALUES(?, ?, ?, ?);

-- name: CreateRequest :exec
INSERT INTO connection_requests (sender_id, reciever_id) VALUES (?, ?);

-- name: GetUserPendingRequests :many
SELECT 
FROM connection_requests
WHERE is_active = true
AND
reciever_id = ?

    SELECT users.email
FROM connection_requests
JOIN users ON connection_requests.sender_user_id = users.user_id
WHERE connection_requests.is_active = 1
  AND connection_requests.receiver_user_id = ?;;
