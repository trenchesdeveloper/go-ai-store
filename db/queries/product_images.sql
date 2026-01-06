-- name: CreateProductImage :one
INSERT INTO product_images (product_id, url, alt_text, is_primary)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetProductImageByID :one
SELECT * FROM product_images
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListProductImages :many
SELECT * FROM product_images
WHERE product_id = $1 AND deleted_at IS NULL
ORDER BY is_primary DESC, created_at ASC;

-- name: GetPrimaryProductImage :one
SELECT * FROM product_images
WHERE product_id = $1 AND is_primary = true AND deleted_at IS NULL;

-- name: SetPrimaryProductImage :exec
UPDATE product_images
SET is_primary = CASE WHEN id = $2 THEN true ELSE false END
WHERE product_id = $1 AND deleted_at IS NULL;

-- name: UpdateProductImage :one
UPDATE product_images
SET url = $2, alt_text = $3, is_primary = $4
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteProductImage :exec
UPDATE product_images
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteProductImagesByProductID :exec
UPDATE product_images
SET deleted_at = CURRENT_TIMESTAMP
WHERE product_id = $1 AND deleted_at IS NULL;
