-- name: CreateCategory :one
INSERT INTO categories (name, description)
VALUES ($1, $2)
RETURNING *;

-- name: GetCategoryByID :one
SELECT * FROM categories
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListCategories :many
SELECT * FROM categories
WHERE deleted_at IS NULL
ORDER BY name ASC
LIMIT $1 OFFSET $2;

-- name: ListActiveCategories :many
SELECT * FROM categories
WHERE is_active = true AND deleted_at IS NULL
ORDER BY name ASC;

-- name: UpdateCategory :one
UPDATE categories
SET name = $2, description = $3, is_active = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateCategoryStatus :one
UPDATE categories
SET is_active = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteCategory :exec
UPDATE categories
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: CountCategories :one
SELECT COUNT(*) FROM categories WHERE deleted_at IS NULL;

-- name: GetCategoriesByIDs :many
SELECT * FROM categories
WHERE id = ANY($1::int[]) AND deleted_at IS NULL;
