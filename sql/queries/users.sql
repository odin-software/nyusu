-- name: CreateUser :one
INSERT INTO users (name, email, password)
VALUES ("", ?, ?)
RETURNING *;

-- name: GetUserById :one
SELECT * 
FROM users
WHERE id = ?;

-- name: GetUserByEmail :one
SELECT * 
FROM users
WHERE email = ?;