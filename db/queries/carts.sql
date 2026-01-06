-- name: CreateCart :one
INSERT INTO carts (user_id)
VALUES ($1)
RETURNING *;

-- name: GetCartByID :one
SELECT * FROM carts
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetCartByUserID :one
SELECT * FROM carts
WHERE user_id = $1 AND deleted_at IS NULL;

-- name: UpdateCartTimestamp :one
UPDATE carts
SET updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteCart :exec
UPDATE carts
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteCartByUserID :exec
UPDATE carts
SET deleted_at = CURRENT_TIMESTAMP
WHERE user_id = $1 AND deleted_at IS NULL;
