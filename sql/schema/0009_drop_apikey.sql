-- +goose Up
ALTER TABLE users
DROP api_key;

-- +goose Down
ALTER TABLE users
ADD api_key VARCHAR(64) NOT NULL DEFAULT '';