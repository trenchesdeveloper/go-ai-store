-- name: CreateOrder :one
INSERT INTO orders (user_id, total_amount)
VALUES ($1, $2)
RETURNING *;

-- name: GetOrderByID :one
SELECT * FROM orders
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListOrdersByUserID :many
SELECT * FROM orders
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListOrders :many
SELECT * FROM orders
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListOrdersByStatus :many
SELECT * FROM orders
WHERE status = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateOrderStatus :one
UPDATE orders
SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateOrderTotal :one
UPDATE orders
SET total_amount = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteOrder :exec
UPDATE orders
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: CountOrders :one
SELECT COUNT(*) FROM orders WHERE deleted_at IS NULL;

-- name: CountOrdersByUserID :one
SELECT COUNT(*) FROM orders WHERE user_id = $1 AND deleted_at IS NULL;

-- name: CountOrdersByStatus :one
SELECT COUNT(*) FROM orders WHERE status = $1 AND deleted_at IS NULL;
