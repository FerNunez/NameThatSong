-- Users
-- name: CreateUser :one
INSERT INTO users (id, email, hashed_password, email_verified, display_name, avatar_url, last_login_at, created_at, updated_at)
VALUES ( $1, $2, $3, $4, $5, $6, NULL, NOW(), NOW())
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: VerifyUserEmail :exec
UPDATE users
SET email_verified = true, updated_at = $2
WHERE id = $1
RETURNING *;

-- name: UpdateLastLogin :exec
UPDATE users
SET last_login_at = $2, updated_at = $3
WHERE id = $1
RETURNING *;

-- name: UpdateUserProfile :exec
UPDATE users
SET display_name = $2, avatar_url = $3, updated_at = $4
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET hashed_password = $2, updated_at = NOW()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: ResetUsers :exec
TRUNCATE TABLE users RESTART IDENTITY CASCADE;

-- Email verification token
-- name: CreateEmailVerificationTokens :one
INSERT INTO email_verification_tokens (id, user_id, token, expires_at, used_at, created_at)
VALUES ( $1, $2, $3, $4, NULL, NOW())
RETURNING *;

-- name: GetEmailVerificationTokensByUserID :one
SELECT * FROM email_verification_tokens WHERE user_id = $1;

-- name: GetEmailVerificationTokensByToken :one
SELECT * FROM email_verification_tokens WHERE token = $1;

-- name: UpdateEmailVerificationTokensUsedAtByToken :one
UPDATE email_verification_tokens
SET used_at = NOW()
WHERE token = $1
RETURNING *;

-- name: CleanupExpiredEmailVerificationTokens :exec
DELETE FROM email_verification_tokens WHERE expires_at < NOW();

-- Password Reset Tokens
-- name: CreatePasswordResetTokens :one
INSERT INTO password_reset_tokens (id, user_id, token, ip_address, user_agent, expires_at, used_at, created_at)
VALUES ( $1, $2, $3, $4, $5, $6, NULL, NOW())
RETURNING *;

-- name: GetPasswordResetTokensByToken :one
SELECT * FROM password_reset_tokens WHERE token = $1;

-- name: GetPasswordResetTokensByUserID :one
SELECT * FROM password_reset_tokens WHERE user_id = $1;

-- name: UpdatePasswordResetTokensUsedAtByToken :exec
UPDATE password_reset_tokens
SET used_at = NOW()
WHERE token = $1
RETURNING *;

-- name: DeletePasswordResetTokens :exec
DELETE FROM password_reset_tokens WHERE id = $1;

-- name: CleanupExpiredPasswordResetTokens :exec
DELETE FROM password_reset_tokens WHERE expires_at < NOW();

-- User Sessions
-- name: CreateSession :one
INSERT INTO user_sessions (id,  user_id, expires_at, revoked_at, created_at, updated_at)
VALUES (
  $1,
  $2,
  $3,
  NULL,
  NOW(),
  NOW()
)
RETURNING *;

-- name: GetSession :one
SELECT * FROM user_sessions WHERE id = $1;

-- name: UpdateSession :exec
UPDATE user_sessions
SET revoked_at = $1,
    updated_at = $2
WHERE id = $3;

-- name: RevokeUserSessionsByUserID :exec
UPDATE user_sessions 
SET revoked_at = NOW(), updated_at = NOW()
WHERE user_id = $1
RETURNING *;

-- name: CleanupExpiredSessions :exec
DELETE FROM user_sessions WHERE expires_at < NOW();
