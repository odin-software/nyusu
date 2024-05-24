-- name: CreateFeed :one
INSERT INTO feeds (name, url, user_id)
VALUES (?, ?, ?)
RETURNING *;