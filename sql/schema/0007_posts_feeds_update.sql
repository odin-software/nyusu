-- +goose Up
ALTER TABLE feeds
ADD COLUMN description TEXT;
ALTER TABLE feeds
ADD COLUMN image_url TEXT;
ALTER TABLE feeds
ADD COLUMN image_text TEXT;
ALTER TABLE feeds
ADD COLUMN language VARCHAR(50);
ALTER TABLE posts
ADD COLUMN content TEXT;

-- +goose Down
ALTER TABLE feeds
DROP description;
ALTER TABLE feeds
DROP image_url;
ALTER TABLE feeds
DROP image_text;
ALTER TABLE feeds
DROP language;
ALTER TABLE posts
DROP content;