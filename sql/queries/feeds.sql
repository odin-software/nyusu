-- FEEDS TABLE
-- name: GetAllFeeds :many
SELECT id, name, url
FROM feeds
LIMIT $1
OFFSET $2;

-- name: GetNextFeedsToFetch :many
SELECT id, name, url
FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT $1;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET
	last_fetched_at = NOW(),
	updated_at = NOW()
WHERE id = $1;

-- name: GetFeedByUrl :one
SELECT *
FROM feeds
WHERE url = $1;

-- name: CreateFeed :one
INSERT INTO feeds (name, url, link, description, image_url, image_text, language, user_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- FEED FOLLOWS TABLE
-- name: GetFeedFollows :one
SELECT id
FROM feed_follows
WHERE feed_id = $1 AND user_id = $2;

-- name: GetFeedFollowsFromUser :many
SELECT id, user_id, feed_id
FROM feed_follows
WHERE user_id = $1;

-- name: CreateFeedFollows :one
INSERT INTO feed_follows (user_id, feed_id)
VALUES ($1, $2)
RETURNING *;

-- name: DeleteFeedFollows :exec
DELETE FROM feed_follows
WHERE id = $1;

-- name: GetAllFeedFollowsByEmail :many
SELECT f.id, f."name", f.url, f.link, f.description, f.created_at, ff.id AS feed_follow_id
FROM feed_follows ff
INNER JOIN feeds f ON ff.feed_id = f.id
INNER JOIN users u ON ff.user_id = u.id
WHERE u.email = $1
LIMIT $2
OFFSET $3;
