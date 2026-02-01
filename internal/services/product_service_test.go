package services

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	db "github.com/trenchesdeveloper/go-ai-store/db/sqlc"
	"github.com/trenchesdeveloper/go-ai-store/internal/dto"
)

// MockProductStore provides mocked store methods for ProductService testing
type MockProductStore struct {
	mock.Mock
}

// Category methods
func (m *MockProductStore) CreateCategory(ctx context.Context, arg db.CreateCategoryParams) (db.Category, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Category), args.Error(1)
}

func (m *MockProductStore) GetCategoryByID(ctx context.Context, id int32) (db.Category, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Category), args.Error(1)
}

func (m *MockProductStore) ListActiveCategories(ctx context.Context) ([]db.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.Category), args.Error(1)
}

func (m *MockProductStore) UpdateCategory(ctx context.Context, arg db.UpdateCategoryParams) (db.Category, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Category), args.Error(1)
}

func (m *MockProductStore) SoftDeleteCategory(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Product methods
func (m *MockProductStore) CreateProduct(ctx context.Context, arg db.CreateProductParams) (db.Product, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Product), args.Error(1)
}

func (m *MockProductStore) GetProductByID(ctx context.Context, id int32) (db.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Product), args.Error(1)
}

func (m *MockProductStore) GetProductBySKU(ctx context.Context, sku string) (db.Product, error) {
	args := m.Called(ctx, sku)
	return args.Get(0).(db.Product), args.Error(1)
}

func (m *MockProductStore) ListActiveProducts(ctx context.Context, arg db.ListActiveProductsParams) ([]db.Product, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.Product), args.Error(1)
}

func (m *MockProductStore) CountActiveProducts(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockProductStore) UpdateProduct(ctx context.Context, arg db.UpdateProductParams) (db.Product, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Product), args.Error(1)
}

func (m *MockProductStore) SoftDeleteProduct(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductStore) UpdateProductStatus(ctx context.Context, arg db.UpdateProductStatusParams) (db.Product, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Product), args.Error(1)
}

func (m *MockProductStore) ListProductImages(ctx context.Context, productID int32) ([]db.ProductImage, error) {
	args := m.Called(ctx, productID)
	return args.Get(0).([]db.ProductImage), args.Error(1)
}

func (m *MockProductStore) ListProductImagesByProductIDs(ctx context.Context, ids []int32) ([]db.ProductImage, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]db.ProductImage), args.Error(1)
}

func (m *MockProductStore) CreateProductImage(ctx context.Context, arg db.CreateProductImageParams) (db.ProductImage, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.ProductImage), args.Error(1)
}

func (m *MockProductStore) GetCategoriesByIDs(ctx context.Context, ids []int32) ([]db.Category, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]db.Category), args.Error(1)
}

func (m *MockProductStore) ExecTx(ctx context.Context, fn func(*db.Queries) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

// Helper to create test category
func createTestCategory() db.Category {
	return db.Category{
		ID:          1,
		Name:        "Electronics",
		Description: pgtype.Text{String: "Electronic devices", Valid: true},
		IsActive:    pgtype.Bool{Bool: true, Valid: true},
		CreatedAt:   pgtype.Timestamptz{Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Valid: true},
	}
}

// Helper to create test product
func createTestProduct() db.Product {
	return db.Product{
		ID:          1,
		Name:        "Test Product",
		Description: pgtype.Text{String: "A test product", Valid: true},
		Price:       pgtype.Numeric{Valid: true},
		Stock:       pgtype.Int4{Int32: 10, Valid: true},
		CategoryID:  1,
		Sku:         "TEST-001",
		IsActive:    pgtype.Bool{Bool: true, Valid: true},
		CreatedAt:   pgtype.Timestamptz{Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Valid: true},
	}
}

func TestProductService_CreateCategory(t *testing.T) {
	t.Parallel()

	testCategory := createTestCategory()

	tests := []struct {
		name      string
		req       dto.CreateCategoryRequest
		setupMock func(m *MockProductStore)
		wantErr   bool
	}{
		{
			name: "success - category created",
			req: dto.CreateCategoryRequest{
				Name:        "Electronics",
				Description: "Electronic devices",
			},
			setupMock: func(m *MockProductStore) {
				m.On("CreateCategory", mock.Anything, mock.MatchedBy(func(arg db.CreateCategoryParams) bool {
					return arg.Name == "Electronics"
				})).Return(testCategory, nil)
			},
			wantErr: false,
		},
		{
			name: "error - database error",
			req: dto.CreateCategoryRequest{
				Name:        "Electronics",
				Description: "Electronic devices",
			},
			setupMock: func(m *MockProductStore) {
				m.On("CreateCategory", mock.Anything, mock.Anything).Return(db.Category{}, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockProductStore)
			tt.setupMock(mockStore)

			service := &ProductService{store: createProductStoreWrapper(mockStore)}

			resp, err := service.CreateCategory(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, testCategory.Name, resp.Name)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestProductService_GetCategories(t *testing.T) {
	t.Parallel()

	testCategories := []db.Category{
		createTestCategory(),
		{
			ID:          2,
			Name:        "Books",
			Description: pgtype.Text{String: "Books and publications", Valid: true},
			IsActive:    pgtype.Bool{Bool: true, Valid: true},
		},
	}

	tests := []struct {
		name      string
		setupMock func(m *MockProductStore)
		wantLen   int
		wantErr   bool
	}{
		{
			name: "success - returns categories",
			setupMock: func(m *MockProductStore) {
				m.On("ListActiveCategories", mock.Anything).Return(testCategories, nil)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "success - empty list",
			setupMock: func(m *MockProductStore) {
				m.On("ListActiveCategories", mock.Anything).Return([]db.Category{}, nil)
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "error - database error",
			setupMock: func(m *MockProductStore) {
				m.On("ListActiveCategories", mock.Anything).Return([]db.Category{}, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockProductStore)
			tt.setupMock(mockStore)

			service := &ProductService{store: createProductStoreWrapper(mockStore)}

			resp, err := service.GetCategories(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, resp, tt.wantLen)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestProductService_GetCategoryByID(t *testing.T) {
	t.Parallel()

	testCategory := createTestCategory()

	tests := []struct {
		name      string
		id        uint
		setupMock func(m *MockProductStore)
		wantErr   bool
	}{
		{
			name: "success - category found",
			id:   1,
			setupMock: func(m *MockProductStore) {
				m.On("GetCategoryByID", mock.Anything, int32(1)).Return(testCategory, nil)
			},
			wantErr: false,
		},
		{
			name: "error - category not found",
			id:   999,
			setupMock: func(m *MockProductStore) {
				m.On("GetCategoryByID", mock.Anything, int32(999)).Return(db.Category{}, pgx.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockProductStore)
			tt.setupMock(mockStore)

			service := &ProductService{store: createProductStoreWrapper(mockStore)}

			resp, err := service.GetCategoryByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, testCategory.Name, resp.Name)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestProductService_DeleteCategory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		id        uint
		setupMock func(m *MockProductStore)
		wantErr   bool
	}{
		{
			name: "success - category deleted",
			id:   1,
			setupMock: func(m *MockProductStore) {
				m.On("SoftDeleteCategory", mock.Anything, int32(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error - category not found",
			id:   999,
			setupMock: func(m *MockProductStore) {
				m.On("SoftDeleteCategory", mock.Anything, int32(999)).Return(pgx.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockProductStore)
			tt.setupMock(mockStore)

			service := &ProductService{store: createProductStoreWrapper(mockStore)}

			err := service.DeleteCategory(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestProductService_GetProductByID(t *testing.T) {
	t.Parallel()

	testProduct := createTestProduct()

	tests := []struct {
		name      string
		id        uint
		setupMock func(m *MockProductStore)
		wantErr   bool
	}{
		{
			name: "success - product found",
			id:   1,
			setupMock: func(m *MockProductStore) {
				m.On("GetProductByID", mock.Anything, int32(1)).Return(testProduct, nil)
				m.On("ListProductImages", mock.Anything, int32(1)).Return([]db.ProductImage{}, nil)
			},
			wantErr: false,
		},
		{
			name: "error - product not found",
			id:   999,
			setupMock: func(m *MockProductStore) {
				m.On("GetProductByID", mock.Anything, int32(999)).Return(db.Product{}, pgx.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockProductStore)
			tt.setupMock(mockStore)

			service := &ProductService{store: createProductStoreWrapper(mockStore)}

			resp, err := service.GetProductByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, testProduct.Name, resp.Name)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestProductService_DeleteProductByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		id        uint
		setupMock func(m *MockProductStore)
		wantErr   bool
	}{
		{
			name: "success - product deleted",
			id:   1,
			setupMock: func(m *MockProductStore) {
				m.On("SoftDeleteProduct", mock.Anything, int32(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error - product not found",
			id:   999,
			setupMock: func(m *MockProductStore) {
				m.On("SoftDeleteProduct", mock.Anything, int32(999)).Return(pgx.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockProductStore)
			tt.setupMock(mockStore)

			service := &ProductService{store: createProductStoreWrapper(mockStore)}

			err := service.DeleteProductByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestProductService_CreateProduct(t *testing.T) {
	t.Parallel()

	testProduct := createTestProduct()
	testCategory := createTestCategory()

	tests := []struct {
		name      string
		req       dto.CreateProductRequest
		setupMock func(m *MockProductStore)
		wantErr   bool
	}{
		{
			name: "success - product created",
			req: dto.CreateProductRequest{
				Name:        "Test Product",
				Description: "A test product",
				Price:       99.99,
				Stock:       10,
				CategoryID:  1,
				SKU:         "TEST-001",
			},
			setupMock: func(m *MockProductStore) {
				m.On("CreateProduct", mock.Anything, mock.Anything).Return(testProduct, nil)
				m.On("GetCategoryByID", mock.Anything, int32(1)).Return(testCategory, nil)
			},
			wantErr: false,
		},
		{
			name: "error - database error",
			req: dto.CreateProductRequest{
				Name:        "Test Product",
				Description: "A test product",
				Price:       99.99,
				Stock:       10,
				CategoryID:  1,
				SKU:         "TEST-001",
			},
			setupMock: func(m *MockProductStore) {
				m.On("CreateProduct", mock.Anything, mock.Anything).Return(db.Product{}, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockProductStore)
			tt.setupMock(mockStore)

			service := &ProductService{store: createProductStoreWrapper(mockStore)}

			resp, err := service.CreateProduct(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, testProduct.Name, resp.Name)
		})
	}
}

func TestProductService_UpdateCategory(t *testing.T) {
	t.Parallel()

	testCategory := createTestCategory()

	tests := []struct {
		name      string
		req       dto.UpdateCategoryRequest
		setupMock func(m *MockProductStore)
		wantErr   bool
	}{
		{
			name: "success - category updated",
			req: dto.UpdateCategoryRequest{
				ID:          1,
				Name:        "Updated Category",
				Description: "Updated description",
			},
			setupMock: func(m *MockProductStore) {
				m.On("GetCategoryByID", mock.Anything, int32(1)).Return(testCategory, nil)
				updatedCat := testCategory
				updatedCat.Name = "Updated Category"
				m.On("UpdateCategory", mock.Anything, mock.Anything).Return(updatedCat, nil)
			},
			wantErr: false,
		},
		{
			name: "success - category updated with IsActive",
			req: dto.UpdateCategoryRequest{
				ID:          1,
				Name:        "Updated Category",
				Description: "Updated description",
				IsActive:    boolPtr(false),
			},
			setupMock: func(m *MockProductStore) {
				m.On("GetCategoryByID", mock.Anything, int32(1)).Return(testCategory, nil)
				updatedCat := testCategory
				updatedCat.Name = "Updated Category"
				updatedCat.IsActive = pgtype.Bool{Bool: false, Valid: true}
				m.On("UpdateCategory", mock.Anything, mock.Anything).Return(updatedCat, nil)
			},
			wantErr: false,
		},
		{
			name: "error - category not found",
			req: dto.UpdateCategoryRequest{
				ID:   999,
				Name: "Updated Category",
			},
			setupMock: func(m *MockProductStore) {
				m.On("GetCategoryByID", mock.Anything, int32(999)).Return(db.Category{}, pgx.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockProductStore)
			tt.setupMock(mockStore)

			service := &ProductService{store: createProductStoreWrapper(mockStore)}

			resp, err := service.UpdateCategory(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}

// Helper for bool pointer
func boolPtr(b bool) *bool {
	return &b
}

func TestProductService_GetProducts(t *testing.T) {
	t.Parallel()

	testProduct := createTestProduct()
	testCategory := createTestCategory()

	tests := []struct {
		name      string
		page      int
		limit     int
		setupMock func(m *MockProductStore)
		wantLen   int
		wantErr   bool
	}{
		{
			name:  "success - returns products",
			page:  1,
			limit: 10,
			setupMock: func(m *MockProductStore) {
				m.On("CountActiveProducts", mock.Anything).Return(int64(2), nil)
				m.On("ListActiveProducts", mock.Anything, mock.Anything).Return([]db.Product{testProduct, testProduct}, nil)
				m.On("GetCategoriesByIDs", mock.Anything, mock.Anything).Return([]db.Category{testCategory}, nil)
				m.On("ListProductImagesByProductIDs", mock.Anything, mock.Anything).Return([]db.ProductImage{}, nil)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:  "success - empty products",
			page:  1,
			limit: 10,
			setupMock: func(m *MockProductStore) {
				m.On("CountActiveProducts", mock.Anything).Return(int64(0), nil)
				m.On("ListActiveProducts", mock.Anything, mock.Anything).Return([]db.Product{}, nil)
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:  "success - default page and limit",
			page:  0,
			limit: 0,
			setupMock: func(m *MockProductStore) {
				m.On("CountActiveProducts", mock.Anything).Return(int64(0), nil)
				m.On("ListActiveProducts", mock.Anything, mock.Anything).Return([]db.Product{}, nil)
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:  "error - database error",
			page:  1,
			limit: 10,
			setupMock: func(m *MockProductStore) {
				m.On("CountActiveProducts", mock.Anything).Return(int64(0), errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockProductStore)
			tt.setupMock(mockStore)

			service := &ProductService{store: createProductStoreWrapper(mockStore)}

			resp, meta, err := service.GetProducts(context.Background(), tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, resp, tt.wantLen)
			assert.NotNil(t, meta)
		})
	}
}

func TestProductService_UpdateProductByID(t *testing.T) {
	t.Parallel()

	testProduct := createTestProduct()

	tests := []struct {
		name      string
		id        uint
		req       *dto.UpdateProductRequest
		setupMock func(m *MockProductStore)
		wantErr   bool
	}{
		{
			name: "success - product updated",
			id:   1,
			req: &dto.UpdateProductRequest{
				Name:        "Updated Product",
				Description: "Updated description",
				Price:       199.99,
				Stock:       20,
				CategoryID:  1,
			},
			setupMock: func(m *MockProductStore) {
				m.On("GetProductByID", mock.Anything, int32(1)).Return(testProduct, nil)
				updatedProduct := testProduct
				updatedProduct.Name = "Updated Product"
				m.On("UpdateProduct", mock.Anything, mock.Anything).Return(updatedProduct, nil)
				m.On("ListProductImages", mock.Anything, int32(1)).Return([]db.ProductImage{}, nil)
			},
			wantErr: false,
		},
		{
			name: "success - product updated with IsActive",
			id:   1,
			req: &dto.UpdateProductRequest{
				Name:        "Updated Product",
				Description: "Updated description",
				Price:       199.99,
				Stock:       20,
				CategoryID:  1,
				IsActive:    boolPtr(false),
			},
			setupMock: func(m *MockProductStore) {
				m.On("GetProductByID", mock.Anything, int32(1)).Return(testProduct, nil)
				updatedProduct := testProduct
				updatedProduct.Name = "Updated Product"
				m.On("UpdateProduct", mock.Anything, mock.Anything).Return(updatedProduct, nil)
				m.On("UpdateProductStatus", mock.Anything, mock.Anything).Return(updatedProduct, nil)
				m.On("ListProductImages", mock.Anything, int32(1)).Return([]db.ProductImage{}, nil)
			},
			wantErr: false,
		},
		{
			name: "error - product not found",
			id:   999,
			req: &dto.UpdateProductRequest{
				Name:        "Updated Product",
				Description: "Updated description",
				Price:       199.99,
				Stock:       20,
				CategoryID:  1,
			},
			setupMock: func(m *MockProductStore) {
				m.On("GetProductByID", mock.Anything, int32(999)).Return(db.Product{}, pgx.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "error - update fails",
			id:   1,
			req: &dto.UpdateProductRequest{
				Name:        "Updated Product",
				Description: "Updated description",
				Price:       199.99,
				Stock:       20,
				CategoryID:  1,
			},
			setupMock: func(m *MockProductStore) {
				m.On("GetProductByID", mock.Anything, int32(1)).Return(testProduct, nil)
				m.On("UpdateProduct", mock.Anything, mock.Anything).Return(db.Product{}, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockProductStore)
			tt.setupMock(mockStore)

			service := &ProductService{store: createProductStoreWrapper(mockStore)}

			resp, err := service.UpdateProductByID(context.Background(), tt.id, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}

// productStoreWrapper wraps MockProductStore to implement db.Store interface
type productStoreWrapper struct {
	*MockProductStore
}

func createProductStoreWrapper(m *MockProductStore) db.Store {
	return &productStoreWrapper{MockProductStore: m}
}

// Implement remaining db.Store methods as stubs
func (s *productStoreWrapper) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	return db.User{}, nil
}
func (s *productStoreWrapper) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return db.User{}, nil
}
func (s *productStoreWrapper) GetUserByID(ctx context.Context, id int32) (db.User, error) {
	return db.User{}, nil
}
func (s *productStoreWrapper) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	return db.User{}, nil
}
func (s *productStoreWrapper) UpdateUserPassword(ctx context.Context, arg db.UpdateUserPasswordParams) error {
	return nil
}
func (s *productStoreWrapper) UpdateUserRole(ctx context.Context, arg db.UpdateUserRoleParams) (db.User, error) {
	return db.User{}, nil
}
func (s *productStoreWrapper) UpdateUserStatus(ctx context.Context, arg db.UpdateUserStatusParams) (db.User, error) {
	return db.User{}, nil
}
func (s *productStoreWrapper) SoftDeleteUser(ctx context.Context, id int32) error { return nil }
func (s *productStoreWrapper) ListUsers(ctx context.Context, arg db.ListUsersParams) ([]db.User, error) {
	return nil, nil
}
func (s *productStoreWrapper) CountUsers(ctx context.Context) (int64, error) { return 0, nil }
func (s *productStoreWrapper) CreateRefreshToken(ctx context.Context, arg db.CreateRefreshTokenParams) (db.RefreshToken, error) {
	return db.RefreshToken{}, nil
}
func (s *productStoreWrapper) GetRefreshToken(ctx context.Context, token string) (db.RefreshToken, error) {
	return db.RefreshToken{}, nil
}
func (s *productStoreWrapper) DeleteRefreshToken(ctx context.Context, token string) error { return nil }
func (s *productStoreWrapper) DeleteRefreshTokensByUserID(ctx context.Context, userID int32) error {
	return nil
}
func (s *productStoreWrapper) DeleteExpiredRefreshTokens(ctx context.Context) error { return nil }
func (s *productStoreWrapper) GetRefreshTokensByUserID(ctx context.Context, userID int32) ([]db.RefreshToken, error) {
	return nil, nil
}
func (s *productStoreWrapper) CreateCart(ctx context.Context, userID int32) (db.Cart, error) {
	return db.Cart{}, nil
}
func (s *productStoreWrapper) GetCartByUserID(ctx context.Context, userID int32) (db.Cart, error) {
	return db.Cart{}, nil
}
func (s *productStoreWrapper) GetCartByID(ctx context.Context, id int32) (db.Cart, error) {
	return db.Cart{}, nil
}
func (s *productStoreWrapper) UpdateCartTimestamp(ctx context.Context, id int32) (db.Cart, error) {
	return db.Cart{}, nil
}
func (s *productStoreWrapper) SoftDeleteCart(ctx context.Context, id int32) error { return nil }
func (s *productStoreWrapper) SoftDeleteCartByUserID(ctx context.Context, userID int32) error {
	return nil
}
func (s *productStoreWrapper) CreateCartItem(ctx context.Context, arg db.CreateCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *productStoreWrapper) GetCartItem(ctx context.Context, arg db.GetCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *productStoreWrapper) GetCartItemByID(ctx context.Context, id int32) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *productStoreWrapper) ListCartItems(ctx context.Context, cartID int32) ([]db.CartItem, error) {
	return nil, nil
}
func (s *productStoreWrapper) UpdateCartItemQuantity(ctx context.Context, arg db.UpdateCartItemQuantityParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *productStoreWrapper) UpsertCartItem(ctx context.Context, arg db.UpsertCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *productStoreWrapper) RestoreCartItem(ctx context.Context, arg db.RestoreCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *productStoreWrapper) SoftDeleteCartItem(ctx context.Context, id int32) error { return nil }
func (s *productStoreWrapper) SoftDeleteCartItemsByCartID(ctx context.Context, cartID int32) error {
	return nil
}
func (s *productStoreWrapper) CountCartItems(ctx context.Context, cartID int32) (int64, error) {
	return 0, nil
}
func (s *productStoreWrapper) ListCategories(ctx context.Context, arg db.ListCategoriesParams) ([]db.Category, error) {
	return nil, nil
}
func (s *productStoreWrapper) UpdateCategoryStatus(ctx context.Context, arg db.UpdateCategoryStatusParams) (db.Category, error) {
	return db.Category{}, nil
}
func (s *productStoreWrapper) CountCategories(ctx context.Context) (int64, error) { return 0, nil }
func (s *productStoreWrapper) GetProductByIDForUpdate(ctx context.Context, id int32) (db.Product, error) {
	return db.Product{}, nil
}
func (s *productStoreWrapper) GetProductsByIDs(ctx context.Context, ids []int32) ([]db.Product, error) {
	return nil, nil
}
func (s *productStoreWrapper) GetProductsByIDsForUpdate(ctx context.Context, ids []int32) ([]db.Product, error) {
	return nil, nil
}
func (s *productStoreWrapper) ListProducts(ctx context.Context, arg db.ListProductsParams) ([]db.Product, error) {
	return nil, nil
}
func (s *productStoreWrapper) ListProductsByCategory(ctx context.Context, arg db.ListProductsByCategoryParams) ([]db.Product, error) {
	return nil, nil
}
func (s *productStoreWrapper) UpdateProductStatus(ctx context.Context, arg db.UpdateProductStatusParams) (db.Product, error) {
	return db.Product{}, nil
}
func (s *productStoreWrapper) UpdateProductStock(ctx context.Context, arg db.UpdateProductStockParams) (db.Product, error) {
	return db.Product{}, nil
}
func (s *productStoreWrapper) CountProducts(ctx context.Context) (int64, error) { return 0, nil }
func (s *productStoreWrapper) CountProductsByCategory(ctx context.Context, categoryID int32) (int64, error) {
	return 0, nil
}
func (s *productStoreWrapper) GetProductImageByID(ctx context.Context, id int32) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *productStoreWrapper) GetPrimaryProductImage(ctx context.Context, productID int32) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *productStoreWrapper) UpdateProductImage(ctx context.Context, arg db.UpdateProductImageParams) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *productStoreWrapper) SetPrimaryProductImage(ctx context.Context, arg db.SetPrimaryProductImageParams) error {
	return nil
}
func (s *productStoreWrapper) SoftDeleteProductImage(ctx context.Context, id int32) error { return nil }
func (s *productStoreWrapper) SoftDeleteProductImagesByProductID(ctx context.Context, productID int32) error {
	return nil
}
func (s *productStoreWrapper) CreateOrder(ctx context.Context, arg db.CreateOrderParams) (db.Order, error) {
	return db.Order{}, nil
}
func (s *productStoreWrapper) GetOrderByID(ctx context.Context, id int32) (db.Order, error) {
	return db.Order{}, nil
}
func (s *productStoreWrapper) ListOrders(ctx context.Context, arg db.ListOrdersParams) ([]db.Order, error) {
	return nil, nil
}
func (s *productStoreWrapper) ListOrdersByUserID(ctx context.Context, arg db.ListOrdersByUserIDParams) ([]db.Order, error) {
	return nil, nil
}
func (s *productStoreWrapper) ListOrdersByStatus(ctx context.Context, arg db.ListOrdersByStatusParams) ([]db.Order, error) {
	return nil, nil
}
func (s *productStoreWrapper) UpdateOrderStatus(ctx context.Context, arg db.UpdateOrderStatusParams) (db.Order, error) {
	return db.Order{}, nil
}
func (s *productStoreWrapper) UpdateOrderTotal(ctx context.Context, arg db.UpdateOrderTotalParams) (db.Order, error) {
	return db.Order{}, nil
}
func (s *productStoreWrapper) SoftDeleteOrder(ctx context.Context, id int32) error { return nil }
func (s *productStoreWrapper) CountOrders(ctx context.Context) (int64, error)      { return 0, nil }
func (s *productStoreWrapper) CountOrdersByUserID(ctx context.Context, userID int32) (int64, error) {
	return 0, nil
}
func (s *productStoreWrapper) CountOrdersByStatus(ctx context.Context, status db.NullOrderStatus) (int64, error) {
	return 0, nil
}
func (s *productStoreWrapper) GetOrderTotal(ctx context.Context, orderID int32) (pgtype.Numeric, error) {
	return pgtype.Numeric{}, nil
}
func (s *productStoreWrapper) CreateOrderItem(ctx context.Context, arg db.CreateOrderItemParams) (db.OrderItem, error) {
	return db.OrderItem{}, nil
}
func (s *productStoreWrapper) GetOrderItemByID(ctx context.Context, id int32) (db.OrderItem, error) {
	return db.OrderItem{}, nil
}
func (s *productStoreWrapper) ListOrderItems(ctx context.Context, orderID int32) ([]db.OrderItem, error) {
	return nil, nil
}
func (s *productStoreWrapper) SoftDeleteOrderItem(ctx context.Context, id int32) error { return nil }
func (s *productStoreWrapper) SoftDeleteOrderItemsByOrderID(ctx context.Context, orderID int32) error {
	return nil
}
func (s *productStoreWrapper) CountOrderItems(ctx context.Context, orderID int32) (int64, error) {
	return 0, nil
}
func (s *productStoreWrapper) CreateIdempotencyKey(ctx context.Context, arg db.CreateIdempotencyKeyParams) (db.OrderIdempotencyKey, error) {
	return db.OrderIdempotencyKey{}, nil
}
func (s *productStoreWrapper) GetIdempotencyKey(ctx context.Context, arg db.GetIdempotencyKeyParams) (db.OrderIdempotencyKey, error) {
	return db.OrderIdempotencyKey{}, nil
}
func (s *productStoreWrapper) UpdateIdempotencyKeyOrderID(ctx context.Context, arg db.UpdateIdempotencyKeyOrderIDParams) error {
	return nil
}
