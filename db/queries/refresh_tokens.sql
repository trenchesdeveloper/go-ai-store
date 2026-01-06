-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (user_id, token, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1 AND deleted_at IS NULL;

-- name: GetRefreshTokensByUserID :many
SELECT * FROM refresh_tokens
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: DeleteRefreshToken :exec
UPDATE refresh_tokens
SET deleted_at = CURRENT_TIMESTAMP
WHERE token = $1 AND deleted_at IS NULL;

-- name: DeleteRefreshTokensByUserID :exec
UPDATE refresh_tokens
SET deleted_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND deleted_at IS NULL;

-- name: DeleteExpiredRefreshTokens :exec
UPDATE refresh_tokens
SET deleted_at = CURRENT_TIMESTAMP
WHERE expires_at < CURRENT_TIMESTAMP AND deleted_at IS NULL;
