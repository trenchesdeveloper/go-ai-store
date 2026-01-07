-- name: CreateProduct :one
INSERT INTO products (category_id, name, description, price, stock, sku)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetProductByID :one
SELECT * FROM products
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetProductBySKU :one
SELECT * FROM products
WHERE sku = $1 AND deleted_at IS NULL;

-- name: ListProducts :many
SELECT * FROM products
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListActiveProducts :many
SELECT * FROM products
WHERE is_active = true AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListProductsByCategory :many
SELECT * FROM products
WHERE category_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateProduct :one
UPDATE products
SET category_id = $2, name = $3, description = $4, price = $5, stock = $6, sku = $7, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateProductStock :one
UPDATE products
SET stock = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateProductStatus :one
UPDATE products
SET is_active = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteProduct :exec
UPDATE products
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: CountProducts :one
SELECT COUNT(*) FROM products WHERE deleted_at IS NULL;

-- name: CountProductsByCategory :one
SELECT COUNT(*) FROM products WHERE category_id = $1 AND deleted_at IS NULL;
