-- +goose Up
ALTER TABLE feeds ADD COLUMN link VARCHAR(255);

-- +goose Down
ALTER TABLE feeds DROP COLUMN link;
