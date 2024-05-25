-- +goose Up
CREATE TABLE feed_follows (
  id INTEGER PRIMARY KEY,
  user_id INTEGER NOT NULL,
  feed_id INTEGER NOT NULL,
  created_at INTEGER NOT NULL DEFAULT (unixepoch()),
  updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
  FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
  FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feed_follows;