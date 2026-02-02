-- Add tsvector column for full-text search (PostgreSQL 12+)
-- Weights: A (highest) = name, sku | B = description
-- Note: Existing rows are automatically populated with this generated column
ALTER TABLE products ADD COLUMN search_vector tsvector
  GENERATED ALWAYS AS (
    setweight(to_tsvector('english', coalesce(name, '')), 'A') ||
    setweight(to_tsvector('simple', coalesce(sku, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(description, '')), 'B')
  ) STORED;

-- Create GIN index for fast full-text search
CREATE INDEX idx_products_search ON products USING GIN (search_vector);

-- Add column comment for documentation
COMMENT ON COLUMN products.search_vector IS 'Full-text search vector combining name (weight A), sku (weight A), and description (weight B)';
