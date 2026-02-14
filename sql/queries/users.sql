-- name: GetOrCreateUserBySub :one
INSERT INTO users (name, email, sub)
VALUES ($1, $2, $3)
ON CONFLICT (sub) DO UPDATE SET
  name = EXCLUDED.name,
  email = EXCLUDED.email,
  updated_at = NOW()
RETURNING *;

-- name: GetUserById :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT *
FROM users
WHERE email = $1;

-- name: GetUserBySub :one
SELECT *
FROM users
WHERE sub = $1;
