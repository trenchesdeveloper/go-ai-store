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

// MockCartStore provides mocked store methods for CartService testing
type MockCartStore struct {
	mock.Mock
}

// Cart methods
func (m *MockCartStore) GetCartByUserID(ctx context.Context, userID int32) (db.Cart, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(db.Cart), args.Error(1)
}

func (m *MockCartStore) CreateCart(ctx context.Context, userID int32) (db.Cart, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(db.Cart), args.Error(1)
}

func (m *MockCartStore) ListCartItems(ctx context.Context, cartID int32) ([]db.CartItem, error) {
	args := m.Called(ctx, cartID)
	return args.Get(0).([]db.CartItem), args.Error(1)
}

func (m *MockCartStore) GetProductByID(ctx context.Context, id int32) (db.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Product), args.Error(1)
}

func (m *MockCartStore) UpsertCartItem(ctx context.Context, arg db.UpsertCartItemParams) (db.CartItem, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.CartItem), args.Error(1)
}

func (m *MockCartStore) GetCartItemByID(ctx context.Context, id int32) (db.CartItem, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.CartItem), args.Error(1)
}

func (m *MockCartStore) UpdateCartItemQuantity(ctx context.Context, arg db.UpdateCartItemQuantityParams) (db.CartItem, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.CartItem), args.Error(1)
}

func (m *MockCartStore) SoftDeleteCartItem(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCartStore) SoftDeleteCartItemsByCartID(ctx context.Context, cartID int32) error {
	args := m.Called(ctx, cartID)
	return args.Error(0)
}

func (m *MockCartStore) GetProductsByIDs(ctx context.Context, ids []int32) ([]db.Product, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]db.Product), args.Error(1)
}

func (m *MockCartStore) GetCategoriesByIDs(ctx context.Context, ids []int32) ([]db.Category, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]db.Category), args.Error(1)
}

func (m *MockCartStore) ListProductImagesByProductIDs(ctx context.Context, ids []int32) ([]db.ProductImage, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]db.ProductImage), args.Error(1)
}

func (m *MockCartStore) ExecTx(ctx context.Context, fn func(*db.Queries) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

// Helper to create test cart
func createTestCart() db.Cart {
	return db.Cart{
		ID:        1,
		UserID:    1,
		CreatedAt: pgtype.Timestamptz{Valid: true},
		UpdatedAt: pgtype.Timestamptz{Valid: true},
	}
}

// Helper to create test cart item
func createTestCartItem() db.CartItem {
	return db.CartItem{
		ID:        1,
		CartID:    1,
		ProductID: 1,
		Quantity:  2,
		CreatedAt: pgtype.Timestamptz{Valid: true},
		UpdatedAt: pgtype.Timestamptz{Valid: true},
	}
}

func TestCartService_GetCart(t *testing.T) {
	t.Parallel()

	testCart := createTestCart()

	tests := []struct {
		name      string
		userID    int32
		setupMock func(m *MockCartStore)
		wantErr   bool
	}{
		{
			name:   "success - existing cart found",
			userID: 1,
			setupMock: func(m *MockCartStore) {
				m.On("GetCartByUserID", mock.Anything, int32(1)).Return(testCart, nil)
				m.On("ListCartItems", mock.Anything, int32(1)).Return([]db.CartItem{}, nil)
			},
			wantErr: false,
		},
		{
			name:   "success - cart not found, creates new cart",
			userID: 2,
			setupMock: func(m *MockCartStore) {
				m.On("GetCartByUserID", mock.Anything, int32(2)).Return(db.Cart{}, pgx.ErrNoRows)
				m.On("CreateCart", mock.Anything, int32(2)).Return(db.Cart{ID: 2, UserID: 2}, nil)
				m.On("ListCartItems", mock.Anything, int32(2)).Return([]db.CartItem{}, nil)
			},
			wantErr: false,
		},
		{
			name:   "error - database error",
			userID: 1,
			setupMock: func(m *MockCartStore) {
				m.On("GetCartByUserID", mock.Anything, int32(1)).Return(db.Cart{}, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockCartStore)
			tt.setupMock(mockStore)

			service := &CartService{store: createCartStoreWrapper(mockStore)}

			resp, err := service.GetCart(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestCartService_AddToCart(t *testing.T) {
	t.Parallel()

	testCart := createTestCart()

	tests := []struct {
		name      string
		userID    int32
		req       dto.AddToCartRequest
		setupMock func(m *MockCartStore)
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "error - product not found",
			userID: 1,
			req: dto.AddToCartRequest{
				ProductID: 999,
				Quantity:  1,
			},
			setupMock: func(m *MockCartStore) {
				m.On("GetCartByUserID", mock.Anything, int32(1)).Return(testCart, nil)
				m.On("GetProductByID", mock.Anything, int32(999)).Return(db.Product{}, pgx.ErrNoRows)
			},
			wantErr: true,
			errMsg:  "product not found",
		},
		{
			name:   "error - insufficient stock",
			userID: 1,
			req: dto.AddToCartRequest{
				ProductID: 1,
				Quantity:  100,
			},
			setupMock: func(m *MockCartStore) {
				m.On("GetCartByUserID", mock.Anything, int32(1)).Return(testCart, nil)
				m.On("GetProductByID", mock.Anything, int32(1)).Return(db.Product{
					ID:       1,
					Stock:    pgtype.Int4{Int32: 5, Valid: true}, // Only 5 in stock
					IsActive: pgtype.Bool{Bool: true, Valid: true},
				}, nil)
			},
			wantErr: true,
			errMsg:  "insufficient stock",
		},
		{
			name:   "error - cart creation fails",
			userID: 999,
			req: dto.AddToCartRequest{
				ProductID: 1,
				Quantity:  1,
			},
			setupMock: func(m *MockCartStore) {
				m.On("GetCartByUserID", mock.Anything, int32(999)).Return(db.Cart{}, pgx.ErrNoRows)
				m.On("CreateCart", mock.Anything, int32(999)).Return(db.Cart{}, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name:   "error - get product fails",
			userID: 1,
			req: dto.AddToCartRequest{
				ProductID: 1,
				Quantity:  1,
			},
			setupMock: func(m *MockCartStore) {
				m.On("GetCartByUserID", mock.Anything, int32(1)).Return(testCart, nil)
				m.On("GetProductByID", mock.Anything, int32(1)).Return(db.Product{}, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockCartStore)
			tt.setupMock(mockStore)

			service := &CartService{store: createCartStoreWrapper(mockStore)}

			resp, err := service.AddToCart(context.Background(), tt.userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestCartService_UpdateCartItem(t *testing.T) {
	t.Parallel()

	testCart := createTestCart()
	testCartItem := createTestCartItem()

	tests := []struct {
		name      string
		userID    int32
		itemID    int32
		req       dto.UpdateCartItemRequest
		setupMock func(m *MockCartStore)
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "error - cart not found",
			userID: 999,
			itemID: 1,
			req:    dto.UpdateCartItemRequest{Quantity: 5},
			setupMock: func(m *MockCartStore) {
				m.On("GetCartByUserID", mock.Anything, int32(999)).Return(db.Cart{}, pgx.ErrNoRows)
			},
			wantErr: true,
			errMsg:  "cart not found",
		},
		{
			name:   "error - cart item not found",
			userID: 1,
			itemID: 999,
			req:    dto.UpdateCartItemRequest{Quantity: 5},
			setupMock: func(m *MockCartStore) {
				m.On("GetCartByUserID", mock.Anything, int32(1)).Return(testCart, nil)
				m.On("GetCartItemByID", mock.Anything, int32(999)).Return(db.CartItem{}, pgx.ErrNoRows)
			},
			wantErr: true,
			errMsg:  "cart item not found",
		},
		{
			name:   "error - item belongs to different cart",
			userID: 1,
			itemID: 1,
			req:    dto.UpdateCartItemRequest{Quantity: 5},
			setupMock: func(m *MockCartStore) {
				m.On("GetCartByUserID", mock.Anything, int32(1)).Return(testCart, nil)
				// Return cart item that belongs to a different cart
				wrongCartItem := testCartItem
				wrongCartItem.CartID = 999
				m.On("GetCartItemByID", mock.Anything, int32(1)).Return(wrongCartItem, nil)
			},
			wantErr: true,
			errMsg:  "cart item not found",
		},
		{
			name:   "error - insufficient stock",
			userID: 1,
			itemID: 1,
			req:    dto.UpdateCartItemRequest{Quantity: 100},
			setupMock: func(m *MockCartStore) {
				m.On("GetCartByUserID", mock.Anything, int32(1)).Return(testCart, nil)
				m.On("GetCartItemByID", mock.Anything, int32(1)).Return(testCartItem, nil)
				m.On("GetProductByID", mock.Anything, int32(1)).Return(db.Product{
					ID:    1,
					Stock: pgtype.Int4{Int32: 5, Valid: true}, // Only 5 in stock
				}, nil)
			},
			wantErr: true,
			errMsg:  "insufficient stock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockCartStore)
			tt.setupMock(mockStore)

			service := &CartService{store: createCartStoreWrapper(mockStore)}

			resp, err := service.UpdateCartItem(context.Background(), tt.userID, tt.itemID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestCartService_RemoveCartItem(t *testing.T) {
	t.Parallel()

	testCart := createTestCart()
	testCartItem := createTestCartItem()

	tests := []struct {
		name      string
		userID    int32
		itemID    int32
		setupMock func(m *MockCartStore)
		wantErr   bool
	}{
		{
			name:   "success - item removed",
			userID: 1,
			itemID: 1,
			setupMock: func(m *MockCartStore) {
				m.On("GetCartByUserID", mock.Anything, int32(1)).Return(testCart, nil)
				m.On("GetCartItemByID", mock.Anything, int32(1)).Return(testCartItem, nil)
				m.On("SoftDeleteCartItem", mock.Anything, int32(1)).Return(nil)
				m.On("ListCartItems", mock.Anything, int32(1)).Return([]db.CartItem{}, nil)
			},
			wantErr: false,
		},
		{
			name:   "error - item not found",
			userID: 1,
			itemID: 999,
			setupMock: func(m *MockCartStore) {
				m.On("GetCartByUserID", mock.Anything, int32(1)).Return(testCart, nil)
				m.On("GetCartItemByID", mock.Anything, int32(999)).Return(db.CartItem{}, pgx.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockCartStore)
			tt.setupMock(mockStore)

			service := &CartService{store: createCartStoreWrapper(mockStore)}

			resp, err := service.RemoveCartItem(context.Background(), tt.userID, tt.itemID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestCartService_ClearCart(t *testing.T) {
	t.Parallel()

	testCart := createTestCart()

	tests := []struct {
		name      string
		userID    int32
		setupMock func(m *MockCartStore)
		wantErr   bool
	}{
		{
			name:   "success - cart cleared",
			userID: 1,
			setupMock: func(m *MockCartStore) {
				m.On("GetCartByUserID", mock.Anything, int32(1)).Return(testCart, nil)
				m.On("SoftDeleteCartItemsByCartID", mock.Anything, int32(1)).Return(nil)
			},
			wantErr: false,
		},
		{
			name:   "success - no cart exists (nothing to clear)",
			userID: 999,
			setupMock: func(m *MockCartStore) {
				m.On("GetCartByUserID", mock.Anything, int32(999)).Return(db.Cart{}, pgx.ErrNoRows)
			},
			wantErr: false, // Per code line 229, returns nil if no cart
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockCartStore)
			tt.setupMock(mockStore)

			service := &CartService{store: createCartStoreWrapper(mockStore)}

			err := service.ClearCart(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			mockStore.AssertExpectations(t)
		})
	}
}

// cartStoreWrapper wraps MockCartStore to implement db.Store interface
type cartStoreWrapper struct {
	*MockCartStore
}

func createCartStoreWrapper(m *MockCartStore) db.Store {
	return &cartStoreWrapper{MockCartStore: m}
}

// Implement remaining db.Store methods as stubs
func (s *cartStoreWrapper) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	return db.User{}, nil
}
func (s *cartStoreWrapper) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return db.User{}, nil
}
func (s *cartStoreWrapper) GetUserByID(ctx context.Context, id int32) (db.User, error) {
	return db.User{}, nil
}
func (s *cartStoreWrapper) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	return db.User{}, nil
}
func (s *cartStoreWrapper) UpdateUserPassword(ctx context.Context, arg db.UpdateUserPasswordParams) error {
	return nil
}
func (s *cartStoreWrapper) UpdateUserRole(ctx context.Context, arg db.UpdateUserRoleParams) (db.User, error) {
	return db.User{}, nil
}
func (s *cartStoreWrapper) UpdateUserStatus(ctx context.Context, arg db.UpdateUserStatusParams) (db.User, error) {
	return db.User{}, nil
}
func (s *cartStoreWrapper) SoftDeleteUser(ctx context.Context, id int32) error { return nil }
func (s *cartStoreWrapper) ListUsers(ctx context.Context, arg db.ListUsersParams) ([]db.User, error) {
	return nil, nil
}
func (s *cartStoreWrapper) CountUsers(ctx context.Context) (int64, error) { return 0, nil }
func (s *cartStoreWrapper) CreateRefreshToken(ctx context.Context, arg db.CreateRefreshTokenParams) (db.RefreshToken, error) {
	return db.RefreshToken{}, nil
}
func (s *cartStoreWrapper) GetRefreshToken(ctx context.Context, token string) (db.RefreshToken, error) {
	return db.RefreshToken{}, nil
}
func (s *cartStoreWrapper) DeleteRefreshToken(ctx context.Context, token string) error { return nil }
func (s *cartStoreWrapper) DeleteRefreshTokensByUserID(ctx context.Context, userID int32) error {
	return nil
}
func (s *cartStoreWrapper) DeleteExpiredRefreshTokens(ctx context.Context) error { return nil }
func (s *cartStoreWrapper) GetRefreshTokensByUserID(ctx context.Context, userID int32) ([]db.RefreshToken, error) {
	return nil, nil
}
func (s *cartStoreWrapper) GetCartByID(ctx context.Context, id int32) (db.Cart, error) {
	return db.Cart{}, nil
}
func (s *cartStoreWrapper) UpdateCartTimestamp(ctx context.Context, id int32) (db.Cart, error) {
	return db.Cart{}, nil
}
func (s *cartStoreWrapper) SoftDeleteCart(ctx context.Context, id int32) error { return nil }
func (s *cartStoreWrapper) SoftDeleteCartByUserID(ctx context.Context, userID int32) error {
	return nil
}
func (s *cartStoreWrapper) CreateCartItem(ctx context.Context, arg db.CreateCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *cartStoreWrapper) GetCartItem(ctx context.Context, arg db.GetCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *cartStoreWrapper) RestoreCartItem(ctx context.Context, arg db.RestoreCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *cartStoreWrapper) CountCartItems(ctx context.Context, cartID int32) (int64, error) {
	return 0, nil
}
func (s *cartStoreWrapper) CreateCategory(ctx context.Context, arg db.CreateCategoryParams) (db.Category, error) {
	return db.Category{}, nil
}
func (s *cartStoreWrapper) GetCategoryByID(ctx context.Context, id int32) (db.Category, error) {
	return db.Category{}, nil
}
func (s *cartStoreWrapper) ListCategories(ctx context.Context, arg db.ListCategoriesParams) ([]db.Category, error) {
	return nil, nil
}
func (s *cartStoreWrapper) ListActiveCategories(ctx context.Context) ([]db.Category, error) {
	return nil, nil
}
func (s *cartStoreWrapper) UpdateCategory(ctx context.Context, arg db.UpdateCategoryParams) (db.Category, error) {
	return db.Category{}, nil
}
func (s *cartStoreWrapper) UpdateCategoryStatus(ctx context.Context, arg db.UpdateCategoryStatusParams) (db.Category, error) {
	return db.Category{}, nil
}
func (s *cartStoreWrapper) SoftDeleteCategory(ctx context.Context, id int32) error { return nil }
func (s *cartStoreWrapper) CountCategories(ctx context.Context) (int64, error)     { return 0, nil }
func (s *cartStoreWrapper) CreateProduct(ctx context.Context, arg db.CreateProductParams) (db.Product, error) {
	return db.Product{}, nil
}
func (s *cartStoreWrapper) GetProductByIDForUpdate(ctx context.Context, id int32) (db.Product, error) {
	return db.Product{}, nil
}
func (s *cartStoreWrapper) GetProductBySKU(ctx context.Context, sku string) (db.Product, error) {
	return db.Product{}, nil
}
func (s *cartStoreWrapper) GetProductsByIDsForUpdate(ctx context.Context, ids []int32) ([]db.Product, error) {
	return nil, nil
}
func (s *cartStoreWrapper) ListProducts(ctx context.Context, arg db.ListProductsParams) ([]db.Product, error) {
	return nil, nil
}
func (s *cartStoreWrapper) ListActiveProducts(ctx context.Context, arg db.ListActiveProductsParams) ([]db.Product, error) {
	return nil, nil
}
func (s *cartStoreWrapper) ListProductsByCategory(ctx context.Context, arg db.ListProductsByCategoryParams) ([]db.Product, error) {
	return nil, nil
}
func (s *cartStoreWrapper) UpdateProduct(ctx context.Context, arg db.UpdateProductParams) (db.Product, error) {
	return db.Product{}, nil
}
func (s *cartStoreWrapper) UpdateProductStatus(ctx context.Context, arg db.UpdateProductStatusParams) (db.Product, error) {
	return db.Product{}, nil
}
func (s *cartStoreWrapper) UpdateProductStock(ctx context.Context, arg db.UpdateProductStockParams) (db.Product, error) {
	return db.Product{}, nil
}
func (s *cartStoreWrapper) SoftDeleteProduct(ctx context.Context, id int32) error  { return nil }
func (s *cartStoreWrapper) CountProducts(ctx context.Context) (int64, error)       { return 0, nil }
func (s *cartStoreWrapper) CountActiveProducts(ctx context.Context) (int64, error) { return 0, nil }
func (s *cartStoreWrapper) CountProductsByCategory(ctx context.Context, categoryID int32) (int64, error) {
	return 0, nil
}
func (s *cartStoreWrapper) CreateProductImage(ctx context.Context, arg db.CreateProductImageParams) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *cartStoreWrapper) GetProductImageByID(ctx context.Context, id int32) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *cartStoreWrapper) GetPrimaryProductImage(ctx context.Context, productID int32) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *cartStoreWrapper) ListProductImages(ctx context.Context, productID int32) ([]db.ProductImage, error) {
	return nil, nil
}
func (s *cartStoreWrapper) UpdateProductImage(ctx context.Context, arg db.UpdateProductImageParams) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *cartStoreWrapper) SetPrimaryProductImage(ctx context.Context, arg db.SetPrimaryProductImageParams) error {
	return nil
}
func (s *cartStoreWrapper) SoftDeleteProductImage(ctx context.Context, id int32) error { return nil }
func (s *cartStoreWrapper) SoftDeleteProductImagesByProductID(ctx context.Context, productID int32) error {
	return nil
}
func (s *cartStoreWrapper) CreateOrder(ctx context.Context, arg db.CreateOrderParams) (db.Order, error) {
	return db.Order{}, nil
}
func (s *cartStoreWrapper) GetOrderByID(ctx context.Context, id int32) (db.Order, error) {
	return db.Order{}, nil
}
func (s *cartStoreWrapper) ListOrders(ctx context.Context, arg db.ListOrdersParams) ([]db.Order, error) {
	return nil, nil
}
func (s *cartStoreWrapper) ListOrdersByUserID(ctx context.Context, arg db.ListOrdersByUserIDParams) ([]db.Order, error) {
	return nil, nil
}
func (s *cartStoreWrapper) ListOrdersByStatus(ctx context.Context, arg db.ListOrdersByStatusParams) ([]db.Order, error) {
	return nil, nil
}
func (s *cartStoreWrapper) UpdateOrderStatus(ctx context.Context, arg db.UpdateOrderStatusParams) (db.Order, error) {
	return db.Order{}, nil
}
func (s *cartStoreWrapper) UpdateOrderTotal(ctx context.Context, arg db.UpdateOrderTotalParams) (db.Order, error) {
	return db.Order{}, nil
}
func (s *cartStoreWrapper) SoftDeleteOrder(ctx context.Context, id int32) error { return nil }
func (s *cartStoreWrapper) CountOrders(ctx context.Context) (int64, error)      { return 0, nil }
func (s *cartStoreWrapper) CountOrdersByUserID(ctx context.Context, userID int32) (int64, error) {
	return 0, nil
}
func (s *cartStoreWrapper) CountOrdersByStatus(ctx context.Context, status db.NullOrderStatus) (int64, error) {
	return 0, nil
}
func (s *cartStoreWrapper) GetOrderTotal(ctx context.Context, orderID int32) (pgtype.Numeric, error) {
	return pgtype.Numeric{}, nil
}
func (s *cartStoreWrapper) CreateOrderItem(ctx context.Context, arg db.CreateOrderItemParams) (db.OrderItem, error) {
	return db.OrderItem{}, nil
}
func (s *cartStoreWrapper) GetOrderItemByID(ctx context.Context, id int32) (db.OrderItem, error) {
	return db.OrderItem{}, nil
}
func (s *cartStoreWrapper) ListOrderItems(ctx context.Context, orderID int32) ([]db.OrderItem, error) {
	return nil, nil
}
func (s *cartStoreWrapper) SoftDeleteOrderItem(ctx context.Context, id int32) error { return nil }
func (s *cartStoreWrapper) SoftDeleteOrderItemsByOrderID(ctx context.Context, orderID int32) error {
	return nil
}
func (s *cartStoreWrapper) CountOrderItems(ctx context.Context, orderID int32) (int64, error) {
	return 0, nil
}
func (s *cartStoreWrapper) CreateIdempotencyKey(ctx context.Context, arg db.CreateIdempotencyKeyParams) (db.OrderIdempotencyKey, error) {
	return db.OrderIdempotencyKey{}, nil
}
func (s *cartStoreWrapper) GetIdempotencyKey(ctx context.Context, arg db.GetIdempotencyKeyParams) (db.OrderIdempotencyKey, error) {
	return db.OrderIdempotencyKey{}, nil
}
func (s *cartStoreWrapper) UpdateIdempotencyKeyOrderID(ctx context.Context, arg db.UpdateIdempotencyKeyOrderIDParams) error {
	return nil
}
