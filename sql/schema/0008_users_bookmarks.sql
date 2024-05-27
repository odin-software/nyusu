-- +goose Up
CREATE TABLE users_bookmarks (
  id INTEGER PRIMARY KEY,
  post_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE
  FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
)

-- +goose Down
DROP TABLE users_bookmarks;