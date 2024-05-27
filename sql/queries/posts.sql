-- name: CreatePost :one
INSERT INTO posts (title, url, description, feed_id, published_at)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetPostsByUser :many
SELECT p.title, p.url, p.published_at
FROM feed_follows ff
INNER JOIN feeds f ON ff.feed_id = f.id
INNER JOIN posts p ON p.feed_id = f.id
WHERE ff.user_id = ?
ORDER BY p.published_at DESC
LIMIT ?
OFFSET ?;
