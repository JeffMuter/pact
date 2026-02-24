-- name: GetAllUsers :many
SELECT * FROM users;

-- name: GetUserById :one
SELECT * from users WHERE user_id = ?;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;

-- name: GetUsernameByUserId :one
SELECT username FROM users WHERE user_id = ?;

-- name: UserIsMemberById :one
SELECT is_member FROM users WHERE user_id = ?;

-- name: CreateUser :one
INSERT INTO users (email, username, password_hash) VALUES (?, ?, ?) returning user_id;

-- name: CreateSession :exec
INSERT INTO sessions(user_id, token, created_at, expires_at) VALUES(?, ?, ?, ?);

-- name: CreateRequest :exec
INSERT INTO connection_requests (sender_id, receiver_id, suggested_manager_id, suggested_worker_id) VALUES (?, ?, ?, ?);

-- name: GetUserPendingRequests :many
SELECT connection_requests.request_id, users.email, connection_requests.sender_id, connection_requests.receiver_id, connection_requests.suggested_manager_id, connection_requests.suggested_worker_id
FROM connection_requests
JOIN users ON connection_requests.sender_id = users.user_id
WHERE connection_requests.is_active = 1
AND connection_requests.receiver_id = ?;

-- name: DeleteConnectionRequestByUserIds :exec
DELETE FROM connection_requests 
WHERE (sender_id = ? AND receiver_id = ?) 
   OR (sender_id = ? AND receiver_id = ?);

-- name: CreateConnection :exec
INSERT INTO connections (manager_id, worker_id) VALUES (?, ?);

-- name: GetConnectionsById :many
SELECT 
    c.connection_id,
    c.manager_id,
    m.username AS manager_username,
    c.worker_id,
    w.username AS worker_username
FROM connections c
JOIN users m ON c.manager_id = m.user_id
JOIN users w ON c.worker_id = w.user_id
WHERE ? IN (c.manager_id, c.worker_id);

-- name: UpdateActiveConnection :exec
UPDATE users SET active_connection_id = ?;

-- name: GetActiveConnectionId :one
SELECT active_connection_id
FROM users
WHERE user_id= ?;

-- name: GetActiveConnectionDetails :one
SELECT worker_id, manager_id
FROM connections
WHERE connection_id = ?;

-- name: GetActiveConnectionUserDetails :one
SELECT 
  CASE WHEN c.manager_id = ? THEN c.worker_id ELSE c.manager_id END AS user_id,
  u.username,
  CASE WHEN c.manager_id = ? THEN 'worker' ELSE 'manager' END AS role
FROM connections c
JOIN users u ON u.user_id = CASE WHEN c.manager_id = ? THEN c.worker_id ELSE c.manager_id END
WHERE c.connection_id = (SELECT active_connection_id FROM users WHERE users.user_id = ?);
