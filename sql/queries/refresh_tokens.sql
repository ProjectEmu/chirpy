-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES ($1, NOW(), NOW(), $2, $3, NULL)
RETURNING token, created_at, updated_at, user_id, expires_at, revoked_at;

-- name: GetRefreshToken :one
SELECT token, created_at, updated_at, user_id, expires_at, revoked_at
FROM refresh_tokens
WHERE token = $1
LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET updated_at = NOW(), revoked_at = NOW()
WHERE token = $1;

-- name: DeleteAllRefreshTokens :exec
DELETE FROM refresh_tokens;