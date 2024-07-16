-- +goose Up
ALTER TABLE posts
ADD author VARCHAR(64) NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE posts
DROP author;