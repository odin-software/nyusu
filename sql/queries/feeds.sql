-- FEEDS TABLE
-- name: GetAllFeeds :many
SELECT id, name, url
FROM feeds;

-- name: GetNextFeedsToFetch :many
SELECT id, name, url 
FROM feeds
ORDER BY last_fetched_at ASC
LIMIT ?;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET
	last_fetched_at = unixepoch(),
	updated_at = unixepoch()
WHERE id = ?;

-- name: CreateFeed :one
INSERT INTO feeds (name, url, description, image_url, image_text, language, user_id)
VALUES (?, ?, ?, ?, ?, ?, ?)
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