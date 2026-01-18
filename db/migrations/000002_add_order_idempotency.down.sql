-- Drop order idempotency keys table
DROP INDEX IF EXISTS idx_order_idempotency_keys_user_key;
DROP TABLE IF EXISTS order_idempotency_keys;
