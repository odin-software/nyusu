-- +goose Up
CREATE TABLE sessions (
  id INTEGER PRIMARY KEY,
  token VARCHAR(64) NOT NULL UNIQUE,
  user_id INTEGER NOT NULL,
  created_at INTEGER NOT NULL DEFAULT (unixepoch()),
  expires_at INTEGER NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index for faster token lookups
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- +goose Down
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_token;
DROP TABLE sessions;
