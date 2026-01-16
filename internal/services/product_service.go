package services

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/trenchesdeveloper/go-ai-store/db/sqlc"
	"github.com/trenchesdeveloper/go-ai-store/internal/dto"
	"github.com/trenchesdeveloper/go-ai-store/internal/utils"
)

type ProductService struct {
	store db.Store
}

func NewProductService(store db.Store) *ProductService {
	return &ProductService{store: store}
}

func (s *ProductService) CreateCategory(ctx context.Context, req dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	category, err := s.store.CreateCategory(ctx, db.CreateCategoryParams{
		Name: req.Name,
		Description: pgtype.Text{
			String: req.Description,
			Valid:  true,
		},
	})
	if err != nil {
		return nil, err
	}
	return &dto.CategoryResponse{
		ID:   int64(category.ID),
		Name: category.Name,
	}, nil
}

func (s *ProductService) GetCategories(ctx context.Context) ([]dto.CategoryResponse, error) {
	categories, err := s.store.ListCategories(ctx, db.ListCategoriesParams{})
	if err != nil {
		return nil, err
	}

	categoryResponses := make([]dto.CategoryResponse, len(categories))
	for i, category := range categories {
		categoryResponses[i] = dto.CategoryResponse{
			ID:          int64(category.ID),
			Name:        category.Name,
			Description: category.Description.String,
			IsActive:    category.IsActive.Bool,
		}
	}

	return categoryResponses, nil
}

func (s *ProductService) UpdateCategory(ctx context.Context, req dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
	// Get existing category to preserve IsActive if not provided
	existing, err := s.store.GetCategoryByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	// Use existing IsActive if not provided in request
	isActive := existing.IsActive
	if req.IsActive != nil {
		isActive = pgtype.Bool{
			Bool:  *req.IsActive,
			Valid: true,
		}
	}

	category, err := s.store.UpdateCategory(ctx, db.UpdateCategoryParams{
		ID:   req.ID,
		Name: req.Name,
		Description: pgtype.Text{
			String: req.Description,
			Valid:  true,
		},
		IsActive: isActive,
	})
	if err != nil {
		return nil, err
	}

	return &dto.CategoryResponse{
		ID:          int64(category.ID),
		Name:        category.Name,
		Description: category.Description.String,
		IsActive:    category.IsActive.Bool,
	}, nil
}

func (s *ProductService) DeleteCategory(ctx context.Context, id uint) error {
	return s.store.SoftDeleteCategory(ctx, int32(id)) //#nosec G115 -- id from validated request
}

func (s *ProductService) CreateProduct(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	var price pgtype.Numeric
	if err := price.Scan(req.Price); err != nil {
		return nil, err
	}

	product, err := s.store.CreateProduct(ctx, db.CreateProductParams{
		Name: req.Name,
		Description: pgtype.Text{
			String: req.Description,
			Valid:  true,
		},
		Price: price,
		Stock: pgtype.Int4{
			Valid: true,
			Int32: int32(req.Stock), //#nosec G115 -- stock is validated
		},
		CategoryID: int32(req.CategoryID), //#nosec G115 -- category ID from validated request
		Sku:        req.SKU,
	})
	if err != nil {
		return nil, err
	}

	priceFloat, _ := product.Price.Float64Value()
	return &dto.ProductResponse{
		ID:          uint(product.ID), //#nosec G115 -- DB ID is always positive
		Name:        product.Name,
		Description: product.Description.String,
		Price:       priceFloat.Float64,
		Stock:       int(product.Stock.Int32),
		CategoryID:  uint(product.CategoryID), //#nosec G115 -- DB ID is always positive
	}, nil
}

func (s *ProductService) GetProducts(ctx context.Context, page, limit int) ([]dto.ProductResponse, *utils.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	// Get total count for pagination
	totalCount, err := s.store.CountActiveProducts(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Calculate total pages
	totalPages := int(totalCount) / limit
	if int(totalCount)%limit > 0 {
		totalPages++
	}

	paginationMeta := &utils.PaginationMeta{
		Page:       page,
		Limit:      limit,
		TotalCount: int(totalCount),
		TotalPages: totalPages,
	}

	products, err := s.store.ListActiveProducts(ctx, db.ListActiveProductsParams{
		Offset: int32((page - 1) * limit), //#nosec G115 -- pagination values are bounded
		Limit:  int32(limit),              //#nosec G115 -- pagination values are bounded
	})
	if err != nil {
		return nil, nil, err
	}

	if len(products) == 0 {
		return []dto.ProductResponse{}, paginationMeta, nil
	}

	// Collect unique category IDs and all product IDs
	categoryIDSet := make(map[int32]struct{})
	productIDs := make([]int32, len(products))
	for i, product := range products {
		categoryIDSet[product.CategoryID] = struct{}{}
		productIDs[i] = product.ID
	}

	// Convert category ID set to slice
	categoryIDs := make([]int32, 0, len(categoryIDSet))
	for id := range categoryIDSet {
		categoryIDs = append(categoryIDs, id)
	}

	// Batch fetch categories
	categories, err := s.store.GetCategoriesByIDs(ctx, categoryIDs)
	if err != nil {
		return nil, nil, err
	}

	// Create category lookup map
	categoryMap := make(map[int32]db.Category, len(categories))
	for _, cat := range categories {
		categoryMap[cat.ID] = cat
	}

	// Batch fetch images
	images, err := s.store.ListProductImagesByProductIDs(ctx, productIDs)
	if err != nil {
		return nil, nil, err
	}

	// Group images by product ID
	imageMap := make(map[int32][]db.ProductImage)
	for _, img := range images {
		imageMap[img.ProductID] = append(imageMap[img.ProductID], img)
	}

	// Build response
	productResponses := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		// Get category from lookup map
		category := categoryMap[product.CategoryID]

		// Get images from lookup map and convert to DTOs
		productImages := imageMap[product.ID]
		imageResponses := make([]dto.ProductImageResponse, len(productImages))
		for j, img := range productImages {
			imageResponses[j] = dto.ProductImageResponse{
				ID:        uint(img.ID), //#nosec G115 -- DB ID is always positive
				URL:       img.Url,
				AltText:   img.AltText.String,
				IsPrimary: img.IsPrimary.Bool,
			}
		}

		priceFloat, _ := product.Price.Float64Value()
		productResponses[i] = dto.ProductResponse{
			ID:          uint(product.ID), //#nosec G115 -- DB ID is always positive
			Name:        product.Name,
			Description: product.Description.String,
			Price:       priceFloat.Float64,
			Stock:       int(product.Stock.Int32),
			CategoryID:  uint(product.CategoryID), //#nosec G115 -- DB ID is always positive
			SKU:         product.Sku,
			IsActive:    product.IsActive.Bool,
			Category: dto.CategoryResponse{
				ID:          int64(category.ID),
				Name:        category.Name,
				Description: category.Description.String,
				IsActive:    category.IsActive.Bool,
			},
			Images: imageResponses,
		}
	}

	return productResponses, paginationMeta, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id uint) (*dto.ProductResponse, error) {
	product, err := s.store.GetProductByID(ctx, int32(id)) //#nosec G115 -- id from validated request
	if err != nil {
		return nil, err
	}

	images, err := s.store.ListProductImages(ctx, int32(id)) //#nosec G115 -- id from validated request
	if err != nil {
		return nil, err
	}

	return s.convertProductToProductResponse(product, images), nil
}

func (s *ProductService) UpdateProductByID(ctx context.Context, id uint, req *dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	// First, fetch the existing product
	existing, err := s.store.GetProductByID(ctx, int32(id)) //#nosec G115 -- id from validated request
	if err != nil {
		return nil, err
	}

	// Prepare price from request
	var price pgtype.Numeric
	if err := price.Scan(req.Price); err != nil {
		return nil, err
	}

	// Update the product with values from the request, using existing values as fallbacks
	product, err := s.store.UpdateProduct(ctx, db.UpdateProductParams{
		ID:   int32(id), //#nosec G115 -- id from validated request
		Name: req.Name,
		Description: pgtype.Text{
			String: req.Description,
			Valid:  true,
		},
		Price: price,
		Stock: pgtype.Int4{
			Valid: true,
			Int32: int32(req.Stock), //#nosec G115 -- stock is validated
		},
		CategoryID: int32(req.CategoryID), //#nosec G115 -- category ID from validated request
		Sku:        existing.Sku,          // Preserve existing SKU since it's not in UpdateProductRequest
	})
	if err != nil {
		return nil, err
	}

	// Handle optional IsActive update
	if req.IsActive != nil {
		product, err = s.store.UpdateProductStatus(ctx, db.UpdateProductStatusParams{
			ID: int32(id), //#nosec G115 -- id from validated request
			IsActive: pgtype.Bool{
				Bool:  *req.IsActive,
				Valid: true,
			},
		})
		if err != nil {
			return nil, err
		}
	}

	images, err := s.store.ListProductImages(ctx, int32(id)) //#nosec G115 -- id from validated request
	if err != nil {
		return nil, err
	}

	return s.convertProductToProductResponse(product, images), nil
}

func (s *ProductService) DeleteProductByID(ctx context.Context, id uint) error {
	return s.store.SoftDeleteProduct(ctx, int32(id)) //#nosec G115 -- id from validated request
}

func (s *ProductService) convertProductToProductResponse(product db.Product, images []db.ProductImage) *dto.ProductResponse {
	// Convert []db.ProductImage to []dto.ProductImageResponse
	imageResponses := make([]dto.ProductImageResponse, len(images))
	for i, img := range images {
		imageResponses[i] = dto.ProductImageResponse{
			ID:        uint(img.ID), //#nosec G115 -- DB ID is always positive
			URL:       img.Url,
			AltText:   img.AltText.String,
			IsPrimary: img.IsPrimary.Bool,
		}
	}

	priceFloat, _ := product.Price.Float64Value()
	return &dto.ProductResponse{
		ID:          uint(product.ID), //#nosec G115 -- DB ID is always positive
		Name:        product.Name,
		Description: product.Description.String,
		Price:       priceFloat.Float64,
		Stock:       int(product.Stock.Int32),
		CategoryID:  uint(product.CategoryID), //#nosec G115 -- DB ID is always positive
		Images:      imageResponses,
	}
}
