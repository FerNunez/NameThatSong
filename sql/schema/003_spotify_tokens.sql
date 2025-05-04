-- +goose Up
CREATE TABLE spotify_tokens(
  refresh_token TEXT PRIMARY KEY,
  access_token TEXT NOT NULL,
  token_type TEXT NOT NULL,
  scope TEXT NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  user_id UUID NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE spotify_tokens;
