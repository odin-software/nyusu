// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: posts.sql

package database

import (
	"context"
	"database/sql"
)

const bookmarkPost = `-- name: BookmarkPost :exec
INSERT INTO users_bookmarks (user_id, post_id)
VALUES (?, ?)
`

type BookmarkPostParams struct {
	UserID int64 `json:"user_id"`
	PostID int64 `json:"post_id"`
}

func (q *Queries) BookmarkPost(ctx context.Context, arg BookmarkPostParams) error {
	_, err := q.db.ExecContext(ctx, bookmarkPost, arg.UserID, arg.PostID)
	return err
}

const createPost = `-- name: CreatePost :one
INSERT INTO posts (title, url, description, feed_id, published_at)
VALUES (?, ?, ?, ?, ?)
RETURNING id, title, url, description, feed_id, created_at, updated_at, published_at, content
`

type CreatePostParams struct {
	Title       string         `json:"title"`
	Url         string         `json:"url"`
	Description sql.NullString `json:"description"`
	FeedID      int64          `json:"feed_id"`
	PublishedAt int64          `json:"published_at"`
}

func (q *Queries) CreatePost(ctx context.Context, arg CreatePostParams) (Post, error) {
	row := q.db.QueryRowContext(ctx, createPost,
		arg.Title,
		arg.Url,
		arg.Description,
		arg.FeedID,
		arg.PublishedAt,
	)
	var i Post
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Url,
		&i.Description,
		&i.FeedID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.PublishedAt,
		&i.Content,
	)
	return i, err
}

const getBookmarkedPostsByDate = `-- name: GetBookmarkedPostsByDate :many
SELECT DISTINCT p.id, p.title, p.url, p.published_at
FROM users_bookmarks ub
INNER JOIN posts p ON p.id = ub.post_id
WHERE ub.user_id = ? 
ORDER BY ub.created_at DESC
LIMIT ?
OFFSET ?
`

type GetBookmarkedPostsByDateParams struct {
	UserID int64 `json:"user_id"`
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}

type GetBookmarkedPostsByDateRow struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Url         string `json:"url"`
	PublishedAt int64  `json:"published_at"`
}

func (q *Queries) GetBookmarkedPostsByDate(ctx context.Context, arg GetBookmarkedPostsByDateParams) ([]GetBookmarkedPostsByDateRow, error) {
	rows, err := q.db.QueryContext(ctx, getBookmarkedPostsByDate, arg.UserID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetBookmarkedPostsByDateRow
	for rows.Next() {
		var i GetBookmarkedPostsByDateRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Url,
			&i.PublishedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getBookmarkedPostsByPublished = `-- name: GetBookmarkedPostsByPublished :many
SELECT DISTINCT p.id, p.title, p.url, p.published_at
FROM users_bookmarks ub
INNER JOIN posts p ON p.id = ub.post_id
WHERE ub.user_id = ? 
ORDER BY p.published_at DESC
LIMIT ?
OFFSET ?
`

type GetBookmarkedPostsByPublishedParams struct {
	UserID int64 `json:"user_id"`
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}

type GetBookmarkedPostsByPublishedRow struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Url         string `json:"url"`
	PublishedAt int64  `json:"published_at"`
}

func (q *Queries) GetBookmarkedPostsByPublished(ctx context.Context, arg GetBookmarkedPostsByPublishedParams) ([]GetBookmarkedPostsByPublishedRow, error) {
	rows, err := q.db.QueryContext(ctx, getBookmarkedPostsByPublished, arg.UserID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetBookmarkedPostsByPublishedRow
	for rows.Next() {
		var i GetBookmarkedPostsByPublishedRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Url,
			&i.PublishedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPostsByUser = `-- name: GetPostsByUser :many
SELECT p.id, p.title, p.url, p.published_at
FROM feed_follows ff
INNER JOIN feeds f ON ff.feed_id = f.id
INNER JOIN posts p ON p.feed_id = f.id
WHERE ff.user_id = ?
ORDER BY p.published_at DESC
LIMIT ?
OFFSET ?
`

type GetPostsByUserParams struct {
	UserID int64 `json:"user_id"`
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}

type GetPostsByUserRow struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Url         string `json:"url"`
	PublishedAt int64  `json:"published_at"`
}

func (q *Queries) GetPostsByUser(ctx context.Context, arg GetPostsByUserParams) ([]GetPostsByUserRow, error) {
	rows, err := q.db.QueryContext(ctx, getPostsByUser, arg.UserID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPostsByUserRow
	for rows.Next() {
		var i GetPostsByUserRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Url,
			&i.PublishedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPostsByUserAndFeed = `-- name: GetPostsByUserAndFeed :many
SELECT p.id, p.title, p.url, p.published_at
FROM feed_follows ff
INNER JOIN feeds f ON ff.feed_id = f.id
INNER JOIN posts p ON p.feed_id = f.id
WHERE ff.user_id = ? AND f.id = ?
ORDER BY p.published_at DESC
LIMIT ?
OFFSET ?
`

type GetPostsByUserAndFeedParams struct {
	UserID int64 `json:"user_id"`
	ID     int64 `json:"id"`
	Limit  int64 `json:"limit"`
	Offset int64 `json:"offset"`
}

type GetPostsByUserAndFeedRow struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Url         string `json:"url"`
	PublishedAt int64  `json:"published_at"`
}

func (q *Queries) GetPostsByUserAndFeed(ctx context.Context, arg GetPostsByUserAndFeedParams) ([]GetPostsByUserAndFeedRow, error) {
	rows, err := q.db.QueryContext(ctx, getPostsByUserAndFeed,
		arg.UserID,
		arg.ID,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetPostsByUserAndFeedRow
	for rows.Next() {
		var i GetPostsByUserAndFeedRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Url,
			&i.PublishedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const unbookmarkPost = `-- name: UnbookmarkPost :exec
DELETE FROM users_bookmarks
WHERE user_id = ? AND post_id = ?
`

type UnbookmarkPostParams struct {
	UserID int64 `json:"user_id"`
	PostID int64 `json:"post_id"`
}

func (q *Queries) UnbookmarkPost(ctx context.Context, arg UnbookmarkPostParams) error {
	_, err := q.db.ExecContext(ctx, unbookmarkPost, arg.UserID, arg.PostID)
	return err
}
