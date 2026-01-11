package services

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/trenchesdeveloper/go-ai-store/db/sqlc"
	"github.com/trenchesdeveloper/go-ai-store/internal/dto"
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

func (s *ProductService) GetProducts(ctx context.Context) ([]dto.ProductResponse, error) {
	products, err := s.store.ListProducts(ctx, db.ListProductsParams{})
	if err != nil {
		return nil, err
	}

	productResponses := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		priceFloat, _ := product.Price.Float64Value()
		productResponses[i] = dto.ProductResponse{
			ID:          uint(product.ID),
			Name:        product.Name,
			Description: product.Description.String,
			Price:       priceFloat.Float64,
			Stock:       int(product.Stock.Int32),
			CategoryID:  uint(product.CategoryID),
		}
	}

	return productResponses, nil
}
