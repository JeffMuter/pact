-- name: GetAllUsers :many
select * from users;

-- name: GetUserById :one
SELECT * from users WHERE user_id = ?;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = ?;
