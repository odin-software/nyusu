-- name: CreatePost :one
INSERT INTO posts (title, url, description, author, feed_id, published_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetPostsByUser :many
SELECT DISTINCT p.id, f.name, p.title, p.author, p.url, p.published_at
FROM feed_follows ff
INNER JOIN users u ON ff.user_id = u.id
INNER JOIN feeds f ON ff.feed_id = f.id
INNER JOIN posts p ON p.feed_id = f.id
WHERE u.email = ?
ORDER BY p.published_at DESC
LIMIT ?
OFFSET ?;

-- name: GetPostsByUserAndFeed :many
SELECT p.id, p.title, f.name, p.url, p.published_at
FROM feed_follows ff
INNER JOIN feeds f ON ff.feed_id = f.id
INNER JOIN posts p ON p.feed_id = f.id
INNER JOIN users u ON ff.user_id = u.id
WHERE u.email = ? AND f.id = ?
ORDER BY p.published_at DESC
LIMIT ?
OFFSET ?;

-- name: GetBookmarkedPostsByPublished :many
SELECT DISTINCT p.id, p.title, p.url, p.published_at
FROM users_bookmarks ub
INNER JOIN posts p ON p.id = ub.post_id
WHERE ub.user_id = ?
ORDER BY p.published_at DESC
LIMIT ?
OFFSET ?;

-- name: GetBookmarkedPostsByDate :many
SELECT DISTINCT p.id, p.title, p.url, p.published_at, f.name
FROM users_bookmarks ub
INNER JOIN posts p ON p.id = ub.post_id
INNER JOIN feeds f ON p.feed_id = f.id
WHERE ub.user_id = ?
ORDER BY ub.created_at DESC
LIMIT ?
OFFSET ?;

-- name: BookmarkPost :exec
INSERT INTO users_bookmarks (user_id, post_id)
VALUES (?, ?);

-- name: UnbookmarkPost :exec
DELETE FROM users_bookmarks
WHERE user_id = ? AND post_id = ?;

-- name: GetPostsByUserWithBookmarks :many
SELECT DISTINCT p.id, f.name, p.title, p.author, p.url, p.published_at,
       CASE WHEN ub.post_id IS NOT NULL THEN 1 ELSE 0 END as is_bookmarked
FROM feed_follows ff
INNER JOIN users u ON ff.user_id = u.id
INNER JOIN feeds f ON ff.feed_id = f.id
INNER JOIN posts p ON p.feed_id = f.id
LEFT JOIN users_bookmarks ub ON ub.post_id = p.id AND ub.user_id = u.id
WHERE u.email = ?
ORDER BY p.published_at DESC
LIMIT ?
OFFSET ?;

-- name: GetPostsByUserAndFeedWithBookmarks :many
SELECT p.id, p.title, f.name, p.url, p.published_at,
       CASE WHEN ub.post_id IS NOT NULL THEN 1 ELSE 0 END as is_bookmarked
FROM feed_follows ff
INNER JOIN feeds f ON ff.feed_id = f.id
INNER JOIN posts p ON p.feed_id = f.id
INNER JOIN users u ON ff.user_id = u.id
LEFT JOIN users_bookmarks ub ON ub.post_id = p.id AND ub.user_id = u.id
WHERE u.email = ? AND f.id = ?
ORDER BY p.published_at DESC
LIMIT ?
OFFSET ?;
