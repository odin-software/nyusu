-- name: CreateUser :one
INSERT INTO users (name)
VALUES (?)
RETURNING *;
