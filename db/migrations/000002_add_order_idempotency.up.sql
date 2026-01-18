-- Order idempotency keys table to prevent duplicate order submissions
CREATE TABLE order_idempotency_keys (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    idempotency_key VARCHAR(255) NOT NULL,
    order_id INTEGER REFERENCES orders(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, idempotency_key)
);

-- Index for fast lookups
CREATE INDEX idx_order_idempotency_keys_user_key ON order_idempotency_keys(user_id, idempotency_key);
