-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, product_id, quantity, price)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetOrderItemByID :one
SELECT * FROM order_items
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListOrderItems :many
SELECT * FROM order_items
WHERE order_id = $1 AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: SoftDeleteOrderItem :exec
UPDATE order_items
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteOrderItemsByOrderID :exec
UPDATE order_items
SET deleted_at = CURRENT_TIMESTAMP
WHERE order_id = $1 AND deleted_at IS NULL;

-- name: CountOrderItems :one
SELECT COUNT(*) FROM order_items WHERE order_id = $1 AND deleted_at IS NULL;

-- name: GetOrderTotal :one
SELECT COALESCE(SUM(quantity * price), 0)::DECIMAL(10,2) as total
FROM order_items
WHERE order_id = $1 AND deleted_at IS NULL;
