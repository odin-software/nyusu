-- FEEDS TABLE
-- name: GetAllFeeds :many
SELECT id, name, url
FROM feeds;

-- name: CreateFeed :one
INSERT INTO feeds (name, url, user_id)
VALUES (?, ?, ?)
RETURNING *;

-- FEED FOLLOWS TABLE
-- name: GetFeedFollows :one
SELECT id
FROM feed_follows
WHERE feed_id = ? AND user_id = ?;

-- name: GetFeedFollowsFromUser :many
SELECT id, user_id, feed_id
FROM feed_follows
WHERE user_id = ?;

-- name: CreateFeedFollows :one
INSERT INTO feed_follows (user_id, feed_id)
VALUES (?, ?) 
RETURNING *;

-- name: DeleteFeedFollows :exec
DELETE FROM feed_follows
WHERE id = ?;