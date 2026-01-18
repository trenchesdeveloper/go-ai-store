-- name: CreateIdempotencyKey :one
INSERT INTO order_idempotency_keys (user_id, idempotency_key, order_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetIdempotencyKey :one
SELECT * FROM order_idempotency_keys
WHERE user_id = $1 AND idempotency_key = $2;

-- name: UpdateIdempotencyKeyOrderID :exec
UPDATE order_idempotency_keys
SET order_id = $3
WHERE user_id = $1 AND idempotency_key = $2;
