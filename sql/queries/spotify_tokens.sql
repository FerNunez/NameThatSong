-- name: CreateSpotifyToken :one
INSERT INTO spotify_tokens (refresh_token, created_at, updated_at, access_token, token_type, scope, expires_at, user_id)
VALUES (
  $1,
  NOW(),
  NOW(),
  $2,
  $3,
  $4,
  $5,
  $6
)
RETURNING *;

-- name: GetSpotifyTokenByID :one
SELECT * FROM spotify_tokens
WHERE user_id = $1;

-- name: UpdateSpotifyAccessToken :exec
UPDATE spotify_tokens
SET access_token = $1,
    expires_at = $2,
    updated_at = NOW()
WHERE user_id = $3;


