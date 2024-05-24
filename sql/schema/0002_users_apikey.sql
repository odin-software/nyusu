-- +goose Up
ALTER TABLE users
ADD api_key VARCHAR(64) NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE users
DROP api_key;