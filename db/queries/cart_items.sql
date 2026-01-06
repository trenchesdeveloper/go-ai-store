-- name: CreateCartItem :one
INSERT INTO cart_items (cart_id, product_id, quantity)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetCartItemByID :one
SELECT * FROM cart_items
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetCartItem :one
SELECT * FROM cart_items
WHERE cart_id = $1 AND product_id = $2 AND deleted_at IS NULL;

-- name: ListCartItems :many
SELECT * FROM cart_items
WHERE cart_id = $1 AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: UpdateCartItemQuantity :one
UPDATE cart_items
SET quantity = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteCartItem :exec
UPDATE cart_items
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteCartItemsByCartID :exec
UPDATE cart_items
SET deleted_at = CURRENT_TIMESTAMP
WHERE cart_id = $1 AND deleted_at IS NULL;

-- name: CountCartItems :one
SELECT COUNT(*) FROM cart_items WHERE cart_id = $1 AND deleted_at IS NULL;
