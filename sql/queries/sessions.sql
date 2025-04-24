-- name: CreateSession :one
INSERT INTO sessions (id, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
  $1,
  NOW(),
  NOW(),
  $2,
  $3,
  NULL
)
RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions WHERE id = $1;

-- name: UpdateSession :exec
UPDATE sessions
SET revoked_at = $1,
    updated_at = $2
WHERE id = $3;
