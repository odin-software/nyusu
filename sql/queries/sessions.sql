-- name: CreateSession :one
INSERT INTO sessions (token, user_id, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSessionByToken :one
SELECT s.id, s.token, s.user_id, s.created_at, s.expires_at,
       u.id AS user_id_2, u.name, u.email, u.sub, u.created_at AS user_created_at, u.updated_at AS user_updated_at
FROM sessions s
INNER JOIN users u ON s.user_id = u.id
WHERE s.token = $1 AND s.expires_at > NOW();

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE token = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at <= NOW();

-- name: DeleteUserSessions :exec
DELETE FROM sessions
WHERE user_id = $1;
