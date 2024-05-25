-- +goose Up
ALTER TABLE feeds
ADD last_fetched_at INTEGER;

-- +goose Down
ALTER TABLE feeds
DROP last_fetched_at;