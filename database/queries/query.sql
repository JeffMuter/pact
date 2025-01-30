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
SELECT connection_requests.request_id, users.email, connection_requests.sender_id, connection_requests.reciever_id
FROM connection_requests
JOIN users ON connection_requests.sender_id = users.user_id
WHERE connection_requests.is_active = 1
AND connection_requests.reciever_id = ?;
