-- name: CreateUser :one
INSERT INTO users (name, api_key)
VALUES (?, ?)
RETURNING *;

-- name: GetUserByApiKey :one
SELECT id, name, created_at, updated_at, api_key
FROM users
WHERE api_key = ?;