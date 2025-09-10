-- name: CreateSession :one
INSERT INTO sessions (token, user_id, expires_at)
VALUES (?, ?, ?)
RETURNING *;

-- name: GetSessionByToken :one
SELECT s.id, s.token, s.user_id, s.created_at, s.expires_at,
       u.id, u.name, u.email, u.password, u.created_at, u.updated_at
FROM sessions s
INNER JOIN users u ON s.user_id = u.id
WHERE s.token = ? AND s.expires_at > unixepoch();

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE token = ?;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at <= unixepoch();

-- name: DeleteUserSessions :exec
DELETE FROM sessions
WHERE user_id = ?;
