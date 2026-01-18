package dto

type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateCategoryRequest struct {
	ID          int32  `json:"id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

type CategoryResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Stock       int     `json:"stock" binding:"min=0"`
	CategoryID  int64   `json:"category_id" binding:"required"`
	SKU         string  `json:"sku" binding:"required"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Stock       int     `json:"stock" binding:"min=0"`
	CategoryID  int64   `json:"category_id" binding:"required"`
	IsActive    *bool   `json:"is_active"`
}

type ProductResponse struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Price       float64                `json:"price"`
	Stock       int                    `json:"stock"`
	CategoryID  uint                   `json:"category_id"`
	SKU         string                 `json:"sku"`
	IsActive    bool                   `json:"is_active"`
	Category    CategoryResponse       `json:"category"`
	Images      []ProductImageResponse `json:"images"`
}

type ProductImageResponse struct {
	ID        uint   `json:"id"`
	URL       string `json:"url"`
	AltText   string `json:"alt_text"`
	IsPrimary bool   `json:"is_primary"`
}
