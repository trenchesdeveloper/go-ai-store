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

// MockUserStore is a minimal mock for UserService testing
type MockUserStore struct {
	mock.Mock
}

// Implement required methods from db.Store interface
func (m *MockUserStore) GetUserByID(ctx context.Context, id int32) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockUserStore) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}

// ExecTx is required by db.Store interface
func (m *MockUserStore) ExecTx(ctx context.Context, fn func(*db.Queries) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

// Helper to create a test user
func createTestUser() db.User {
	return db.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Phone:     pgtype.Text{String: "1234567890", Valid: true},
		Role:      db.NullUserRole{UserRole: db.UserRoleCustomer, Valid: true},
		IsActive:  pgtype.Bool{Bool: true, Valid: true},
		CreatedAt: pgtype.Timestamptz{Valid: true},
		UpdatedAt: pgtype.Timestamptz{Valid: true},
	}
}

func TestUserService_GetProfile(t *testing.T) {
	t.Parallel()

	testUser := createTestUser()

	tests := []struct {
		name       string
		userID     uint
		setupMock  func(m *MockUserStore)
		wantErr    bool
		errMessage string
	}{
		{
			name:   "success - user found",
			userID: 1,
			setupMock: func(m *MockUserStore) {
				m.On("GetUserByID", mock.Anything, int32(1)).Return(testUser, nil)
			},
			wantErr: false,
		},
		{
			name:   "error - user not found",
			userID: 999,
			setupMock: func(m *MockUserStore) {
				m.On("GetUserByID", mock.Anything, int32(999)).Return(db.User{}, pgx.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name:   "error - database error",
			userID: 1,
			setupMock: func(m *MockUserStore) {
				m.On("GetUserByID", mock.Anything, int32(1)).Return(db.User{}, errors.New("database connection error"))
			},
			wantErr: true,
		},
		{
			name:       "error - invalid user ID (too large)",
			userID:     uint(1) << 32, // Larger than MaxInt32
			setupMock:  func(m *MockUserStore) {},
			wantErr:    true,
			errMessage: "invalid user ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockUserStore)
			tt.setupMock(mockStore)

			// Create UserService with mock store
			// Note: We need to use the actual store interface, so we create a wrapper
			service := &UserService{store: createStoreWrapper(mockStore)}

			resp, err := service.GetProfile(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMessage != "" {
					assert.Contains(t, err.Error(), tt.errMessage)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, testUser.Email, resp.Email)
			assert.Equal(t, testUser.FirstName, resp.FirstName)
			assert.Equal(t, testUser.LastName, resp.LastName)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestUserService_UpdateProfile(t *testing.T) {
	t.Parallel()

	testUser := createTestUser()
	updatedUser := testUser
	updatedUser.FirstName = "Jane"
	updatedUser.LastName = "Smith"
	updatedUser.Phone = pgtype.Text{String: "9876543210", Valid: true}

	tests := []struct {
		name      string
		userID    uint
		req       dto.UpdateProfileRequest
		setupMock func(m *MockUserStore)
		wantErr   bool
	}{
		{
			name:   "success - profile updated",
			userID: 1,
			req: dto.UpdateProfileRequest{
				FirstName: "Jane",
				LastName:  "Smith",
				Phone:     "9876543210",
			},
			setupMock: func(m *MockUserStore) {
				m.On("GetUserByID", mock.Anything, int32(1)).Return(testUser, nil)
				m.On("UpdateUser", mock.Anything, mock.MatchedBy(func(arg db.UpdateUserParams) bool {
					return arg.ID == testUser.ID && arg.FirstName == "Jane"
				})).Return(updatedUser, nil)
			},
			wantErr: false,
		},
		{
			name:   "error - user not found",
			userID: 999,
			req: dto.UpdateProfileRequest{
				FirstName: "Jane",
				LastName:  "Smith",
			},
			setupMock: func(m *MockUserStore) {
				m.On("GetUserByID", mock.Anything, int32(999)).Return(db.User{}, pgx.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name:   "error - update fails",
			userID: 1,
			req: dto.UpdateProfileRequest{
				FirstName: "Jane",
				LastName:  "Smith",
			},
			setupMock: func(m *MockUserStore) {
				m.On("GetUserByID", mock.Anything, int32(1)).Return(testUser, nil)
				m.On("UpdateUser", mock.Anything, mock.Anything).Return(db.User{}, errors.New("update failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockUserStore)
			tt.setupMock(mockStore)

			service := &UserService{store: createStoreWrapper(mockStore)}

			resp, err := service.UpdateProfile(context.Background(), tt.userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, "Jane", resp.FirstName)
			assert.Equal(t, "Smith", resp.LastName)
			mockStore.AssertExpectations(t)
		})
	}
}

// storeWrapper wraps MockUserStore to implement the full db.Store interface
type storeWrapper struct {
	*MockUserStore
}

func createStoreWrapper(m *MockUserStore) db.Store {
	return &storeWrapper{MockUserStore: m}
}

// Implement remaining db.Store methods with stubs (not used in UserService)
func (s *storeWrapper) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	return db.User{}, nil
}
func (s *storeWrapper) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return db.User{}, nil
}
func (s *storeWrapper) UpdateUserPassword(ctx context.Context, arg db.UpdateUserPasswordParams) error {
	return nil
}
func (s *storeWrapper) UpdateUserRole(ctx context.Context, arg db.UpdateUserRoleParams) (db.User, error) {
	return db.User{}, nil
}
func (s *storeWrapper) UpdateUserStatus(ctx context.Context, arg db.UpdateUserStatusParams) (db.User, error) {
	return db.User{}, nil
}
func (s *storeWrapper) SoftDeleteUser(ctx context.Context, id int32) error { return nil }
func (s *storeWrapper) ListUsers(ctx context.Context, arg db.ListUsersParams) ([]db.User, error) {
	return nil, nil
}
func (s *storeWrapper) CountUsers(ctx context.Context) (int64, error) { return 0, nil }
func (s *storeWrapper) CreateRefreshToken(ctx context.Context, arg db.CreateRefreshTokenParams) (db.RefreshToken, error) {
	return db.RefreshToken{}, nil
}
func (s *storeWrapper) GetRefreshToken(ctx context.Context, token string) (db.RefreshToken, error) {
	return db.RefreshToken{}, nil
}
func (s *storeWrapper) DeleteRefreshToken(ctx context.Context, token string) error { return nil }
func (s *storeWrapper) DeleteRefreshTokensByUserID(ctx context.Context, userID int32) error {
	return nil
}
func (s *storeWrapper) DeleteExpiredRefreshTokens(ctx context.Context) error { return nil }
func (s *storeWrapper) GetRefreshTokensByUserID(ctx context.Context, userID int32) ([]db.RefreshToken, error) {
	return nil, nil
}
func (s *storeWrapper) CreateCart(ctx context.Context, userID int32) (db.Cart, error) {
	return db.Cart{}, nil
}
func (s *storeWrapper) GetCartByUserID(ctx context.Context, userID int32) (db.Cart, error) {
	return db.Cart{}, nil
}
func (s *storeWrapper) GetCartByID(ctx context.Context, id int32) (db.Cart, error) {
	return db.Cart{}, nil
}
func (s *storeWrapper) UpdateCartTimestamp(ctx context.Context, id int32) (db.Cart, error) {
	return db.Cart{}, nil
}
func (s *storeWrapper) SoftDeleteCart(ctx context.Context, id int32) error             { return nil }
func (s *storeWrapper) SoftDeleteCartByUserID(ctx context.Context, userID int32) error { return nil }
func (s *storeWrapper) CreateCartItem(ctx context.Context, arg db.CreateCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *storeWrapper) GetCartItem(ctx context.Context, arg db.GetCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *storeWrapper) GetCartItemByID(ctx context.Context, id int32) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *storeWrapper) ListCartItems(ctx context.Context, cartID int32) ([]db.CartItem, error) {
	return nil, nil
}
func (s *storeWrapper) UpdateCartItemQuantity(ctx context.Context, arg db.UpdateCartItemQuantityParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *storeWrapper) UpsertCartItem(ctx context.Context, arg db.UpsertCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *storeWrapper) RestoreCartItem(ctx context.Context, arg db.RestoreCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *storeWrapper) SoftDeleteCartItem(ctx context.Context, id int32) error { return nil }
func (s *storeWrapper) SoftDeleteCartItemsByCartID(ctx context.Context, cartID int32) error {
	return nil
}
func (s *storeWrapper) CountCartItems(ctx context.Context, cartID int32) (int64, error) {
	return 0, nil
}
func (s *storeWrapper) CreateCategory(ctx context.Context, arg db.CreateCategoryParams) (db.Category, error) {
	return db.Category{}, nil
}
func (s *storeWrapper) GetCategoryByID(ctx context.Context, id int32) (db.Category, error) {
	return db.Category{}, nil
}
func (s *storeWrapper) GetCategoriesByIDs(ctx context.Context, ids []int32) ([]db.Category, error) {
	return nil, nil
}
func (s *storeWrapper) ListCategories(ctx context.Context, arg db.ListCategoriesParams) ([]db.Category, error) {
	return nil, nil
}
func (s *storeWrapper) ListActiveCategories(ctx context.Context) ([]db.Category, error) {
	return nil, nil
}
func (s *storeWrapper) UpdateCategory(ctx context.Context, arg db.UpdateCategoryParams) (db.Category, error) {
	return db.Category{}, nil
}
func (s *storeWrapper) UpdateCategoryStatus(ctx context.Context, arg db.UpdateCategoryStatusParams) (db.Category, error) {
	return db.Category{}, nil
}
func (s *storeWrapper) SoftDeleteCategory(ctx context.Context, id int32) error { return nil }
func (s *storeWrapper) CountCategories(ctx context.Context) (int64, error)     { return 0, nil }
func (s *storeWrapper) CreateProduct(ctx context.Context, arg db.CreateProductParams) (db.Product, error) {
	return db.Product{}, nil
}
func (s *storeWrapper) GetProductByID(ctx context.Context, id int32) (db.Product, error) {
	return db.Product{}, nil
}
func (s *storeWrapper) GetProductByIDForUpdate(ctx context.Context, id int32) (db.Product, error) {
	return db.Product{}, nil
}
func (s *storeWrapper) GetProductBySKU(ctx context.Context, sku string) (db.Product, error) {
	return db.Product{}, nil
}
func (s *storeWrapper) GetProductsByIDs(ctx context.Context, ids []int32) ([]db.Product, error) {
	return nil, nil
}
func (s *storeWrapper) GetProductsByIDsForUpdate(ctx context.Context, ids []int32) ([]db.Product, error) {
	return nil, nil
}
func (s *storeWrapper) ListProducts(ctx context.Context, arg db.ListProductsParams) ([]db.Product, error) {
	return nil, nil
}
func (s *storeWrapper) ListActiveProducts(ctx context.Context, arg db.ListActiveProductsParams) ([]db.Product, error) {
	return nil, nil
}
func (s *storeWrapper) ListProductsByCategory(ctx context.Context, arg db.ListProductsByCategoryParams) ([]db.Product, error) {
	return nil, nil
}
func (s *storeWrapper) UpdateProduct(ctx context.Context, arg db.UpdateProductParams) (db.Product, error) {
	return db.Product{}, nil
}
func (s *storeWrapper) UpdateProductStatus(ctx context.Context, arg db.UpdateProductStatusParams) (db.Product, error) {
	return db.Product{}, nil
}
func (s *storeWrapper) UpdateProductStock(ctx context.Context, arg db.UpdateProductStockParams) (db.Product, error) {
	return db.Product{}, nil
}
func (s *storeWrapper) SoftDeleteProduct(ctx context.Context, id int32) error  { return nil }
func (s *storeWrapper) CountProducts(ctx context.Context) (int64, error)       { return 0, nil }
func (s *storeWrapper) CountActiveProducts(ctx context.Context) (int64, error) { return 0, nil }
func (s *storeWrapper) CountProductsByCategory(ctx context.Context, categoryID int32) (int64, error) {
	return 0, nil
}
func (s *storeWrapper) CreateProductImage(ctx context.Context, arg db.CreateProductImageParams) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *storeWrapper) GetProductImageByID(ctx context.Context, id int32) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *storeWrapper) GetPrimaryProductImage(ctx context.Context, productID int32) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *storeWrapper) ListProductImages(ctx context.Context, productID int32) ([]db.ProductImage, error) {
	return nil, nil
}
func (s *storeWrapper) ListProductImagesByProductIDs(ctx context.Context, ids []int32) ([]db.ProductImage, error) {
	return nil, nil
}
func (s *storeWrapper) UpdateProductImage(ctx context.Context, arg db.UpdateProductImageParams) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *storeWrapper) SetPrimaryProductImage(ctx context.Context, arg db.SetPrimaryProductImageParams) error {
	return nil
}
func (s *storeWrapper) SoftDeleteProductImage(ctx context.Context, id int32) error { return nil }
func (s *storeWrapper) SoftDeleteProductImagesByProductID(ctx context.Context, productID int32) error {
	return nil
}
func (s *storeWrapper) CreateOrder(ctx context.Context, arg db.CreateOrderParams) (db.Order, error) {
	return db.Order{}, nil
}
func (s *storeWrapper) GetOrderByID(ctx context.Context, id int32) (db.Order, error) {
	return db.Order{}, nil
}
func (s *storeWrapper) ListOrders(ctx context.Context, arg db.ListOrdersParams) ([]db.Order, error) {
	return nil, nil
}
func (s *storeWrapper) ListOrdersByUserID(ctx context.Context, arg db.ListOrdersByUserIDParams) ([]db.Order, error) {
	return nil, nil
}
func (s *storeWrapper) ListOrdersByStatus(ctx context.Context, arg db.ListOrdersByStatusParams) ([]db.Order, error) {
	return nil, nil
}
func (s *storeWrapper) UpdateOrderStatus(ctx context.Context, arg db.UpdateOrderStatusParams) (db.Order, error) {
	return db.Order{}, nil
}
func (s *storeWrapper) UpdateOrderTotal(ctx context.Context, arg db.UpdateOrderTotalParams) (db.Order, error) {
	return db.Order{}, nil
}
func (s *storeWrapper) SoftDeleteOrder(ctx context.Context, id int32) error { return nil }
func (s *storeWrapper) CountOrders(ctx context.Context) (int64, error)      { return 0, nil }
func (s *storeWrapper) CountOrdersByUserID(ctx context.Context, userID int32) (int64, error) {
	return 0, nil
}
func (s *storeWrapper) CountOrdersByStatus(ctx context.Context, status db.NullOrderStatus) (int64, error) {
	return 0, nil
}
func (s *storeWrapper) GetOrderTotal(ctx context.Context, orderID int32) (pgtype.Numeric, error) {
	return pgtype.Numeric{}, nil
}
func (s *storeWrapper) CreateOrderItem(ctx context.Context, arg db.CreateOrderItemParams) (db.OrderItem, error) {
	return db.OrderItem{}, nil
}
func (s *storeWrapper) GetOrderItemByID(ctx context.Context, id int32) (db.OrderItem, error) {
	return db.OrderItem{}, nil
}
func (s *storeWrapper) ListOrderItems(ctx context.Context, orderID int32) ([]db.OrderItem, error) {
	return nil, nil
}
func (s *storeWrapper) SoftDeleteOrderItem(ctx context.Context, id int32) error { return nil }
func (s *storeWrapper) SoftDeleteOrderItemsByOrderID(ctx context.Context, orderID int32) error {
	return nil
}
func (s *storeWrapper) CountOrderItems(ctx context.Context, orderID int32) (int64, error) {
	return 0, nil
}
func (s *storeWrapper) CreateIdempotencyKey(ctx context.Context, arg db.CreateIdempotencyKeyParams) (db.OrderIdempotencyKey, error) {
	return db.OrderIdempotencyKey{}, nil
}
func (s *storeWrapper) GetIdempotencyKey(ctx context.Context, arg db.GetIdempotencyKeyParams) (db.OrderIdempotencyKey, error) {
	return db.OrderIdempotencyKey{}, nil
}
func (s *storeWrapper) UpdateIdempotencyKeyOrderID(ctx context.Context, arg db.UpdateIdempotencyKeyOrderIDParams) error {
	return nil
}
