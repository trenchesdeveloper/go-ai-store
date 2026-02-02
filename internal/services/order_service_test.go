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
)

// MockOrderStore provides mocked store methods for OrderService testing
type MockOrderStore struct {
	mock.Mock
}

// Order methods
func (m *MockOrderStore) GetOrderByID(ctx context.Context, id int32) (db.Order, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Order), args.Error(1)
}

func (m *MockOrderStore) ListOrdersByUserID(ctx context.Context, arg db.ListOrdersByUserIDParams) ([]db.Order, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.Order), args.Error(1)
}

func (m *MockOrderStore) CountOrdersByUserID(ctx context.Context, userID int32) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderStore) UpdateOrderStatus(ctx context.Context, arg db.UpdateOrderStatusParams) (db.Order, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Order), args.Error(1)
}

func (m *MockOrderStore) ListOrderItems(ctx context.Context, orderID int32) ([]db.OrderItem, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).([]db.OrderItem), args.Error(1)
}

func (m *MockOrderStore) GetProductsByIDs(ctx context.Context, ids []int32) ([]db.Product, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]db.Product), args.Error(1)
}

func (m *MockOrderStore) GetCategoriesByIDs(ctx context.Context, ids []int32) ([]db.Category, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]db.Category), args.Error(1)
}

func (m *MockOrderStore) ListProductImagesByProductIDs(ctx context.Context, ids []int32) ([]db.ProductImage, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]db.ProductImage), args.Error(1)
}

func (m *MockOrderStore) ExecTx(ctx context.Context, fn func(*db.Queries) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

// Helper to create test order
func createTestOrder() db.Order {
	return db.Order{
		ID:          1,
		UserID:      1,
		Status:      db.NullOrderStatus{OrderStatus: db.OrderStatusPending, Valid: true},
		TotalAmount: pgtype.Numeric{Valid: true},
		CreatedAt:   pgtype.Timestamptz{Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Valid: true},
	}
}

func TestOrderService_GetOrderByID(t *testing.T) {
	t.Parallel()

	testOrder := createTestOrder()

	tests := []struct {
		name      string
		userID    int32
		orderID   int32
		isAdmin   bool
		setupMock func(m *MockOrderStore)
		wantErr   bool
		errMsg    string
	}{
		{
			name:    "success - owner retrieves order",
			userID:  1,
			orderID: 1,
			isAdmin: false,
			setupMock: func(m *MockOrderStore) {
				m.On("GetOrderByID", mock.Anything, int32(1)).Return(testOrder, nil)
				m.On("ListOrderItems", mock.Anything, int32(1)).Return([]db.OrderItem{}, nil)
			},
			wantErr: false,
		},
		{
			name:    "success - admin retrieves any order",
			userID:  999, // Different user
			orderID: 1,
			isAdmin: true,
			setupMock: func(m *MockOrderStore) {
				m.On("GetOrderByID", mock.Anything, int32(1)).Return(testOrder, nil)
				m.On("ListOrderItems", mock.Anything, int32(1)).Return([]db.OrderItem{}, nil)
			},
			wantErr: false,
		},
		{
			name:    "error - order not found",
			userID:  1,
			orderID: 999,
			isAdmin: false,
			setupMock: func(m *MockOrderStore) {
				m.On("GetOrderByID", mock.Anything, int32(999)).Return(db.Order{}, pgx.ErrNoRows)
			},
			wantErr: true,
			errMsg:  "order not found",
		},
		{
			name:    "error - unauthorized access",
			userID:  2, // Different user
			orderID: 1,
			isAdmin: false,
			setupMock: func(m *MockOrderStore) {
				m.On("GetOrderByID", mock.Anything, int32(1)).Return(testOrder, nil) // Order belongs to user 1
			},
			wantErr: true,
			errMsg:  "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockOrderStore)
			tt.setupMock(mockStore)

			service := &OrderService{store: createOrderStoreWrapper(mockStore)}

			resp, err := service.GetOrderByID(context.Background(), tt.userID, tt.orderID, tt.isAdmin)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}

func TestOrderService_UpdateOrderStatus(t *testing.T) {
	t.Parallel()

	updatedOrder := createTestOrder()
	updatedOrder.Status = db.NullOrderStatus{OrderStatus: db.OrderStatusConfirmed, Valid: true}

	tests := []struct {
		name      string
		orderID   int32
		status    string
		setupMock func(m *MockOrderStore)
		wantErr   bool
		errMsg    string
	}{
		{
			name:    "success - status updated to confirmed",
			orderID: 1,
			status:  "confirmed",
			setupMock: func(m *MockOrderStore) {
				m.On("UpdateOrderStatus", mock.Anything, mock.MatchedBy(func(arg db.UpdateOrderStatusParams) bool {
					return arg.ID == 1 && arg.Status.OrderStatus == db.OrderStatusConfirmed
				})).Return(updatedOrder, nil)
				m.On("ListOrderItems", mock.Anything, int32(1)).Return([]db.OrderItem{}, nil)
			},
			wantErr: false,
		},
		{
			name:    "success - status updated to shipped",
			orderID: 1,
			status:  "shipped",
			setupMock: func(m *MockOrderStore) {
				shippedOrder := createTestOrder()
				shippedOrder.Status = db.NullOrderStatus{OrderStatus: db.OrderStatusShipped, Valid: true}
				m.On("UpdateOrderStatus", mock.Anything, mock.Anything).Return(shippedOrder, nil)
				m.On("ListOrderItems", mock.Anything, int32(1)).Return([]db.OrderItem{}, nil)
			},
			wantErr: false,
		},
		{
			name:    "error - invalid status",
			orderID: 1,
			status:  "invalid_status",
			setupMock: func(m *MockOrderStore) {
				// No mock needed - validation fails first
			},
			wantErr: true,
			errMsg:  "invalid order status",
		},
		{
			name:    "error - order not found",
			orderID: 999,
			status:  "confirmed",
			setupMock: func(m *MockOrderStore) {
				m.On("UpdateOrderStatus", mock.Anything, mock.Anything).Return(db.Order{}, pgx.ErrNoRows)
			},
			wantErr: true,
			errMsg:  "order not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockOrderStore)
			tt.setupMock(mockStore)

			service := &OrderService{store: createOrderStoreWrapper(mockStore)}

			resp, err := service.UpdateOrderStatus(context.Background(), tt.orderID, tt.status)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}

func TestOrderService_CancelOrder(t *testing.T) {
	t.Parallel()

	pendingOrder := createTestOrder()
	confirmedOrder := createTestOrder()
	confirmedOrder.Status = db.NullOrderStatus{OrderStatus: db.OrderStatusConfirmed, Valid: true}

	tests := []struct {
		name      string
		userID    int32
		orderID   int32
		setupMock func(m *MockOrderStore)
		wantErr   bool
		errMsg    string
	}{
		{
			name:    "success - pending order cancelled",
			userID:  1,
			orderID: 1,
			setupMock: func(m *MockOrderStore) {
				m.On("GetOrderByID", mock.Anything, int32(1)).Return(pendingOrder, nil)
				cancelledOrder := pendingOrder
				cancelledOrder.Status = db.NullOrderStatus{OrderStatus: db.OrderStatusCancelled, Valid: true}
				m.On("UpdateOrderStatus", mock.Anything, mock.Anything).Return(cancelledOrder, nil)
				m.On("ListOrderItems", mock.Anything, int32(1)).Return([]db.OrderItem{}, nil)
			},
			wantErr: false,
		},
		{
			name:    "error - order not found",
			userID:  1,
			orderID: 999,
			setupMock: func(m *MockOrderStore) {
				m.On("GetOrderByID", mock.Anything, int32(999)).Return(db.Order{}, pgx.ErrNoRows)
			},
			wantErr: true,
			errMsg:  "order not found",
		},
		{
			name:    "error - not owner",
			userID:  2, // Different user
			orderID: 1,
			setupMock: func(m *MockOrderStore) {
				m.On("GetOrderByID", mock.Anything, int32(1)).Return(pendingOrder, nil)
			},
			wantErr: true,
			errMsg:  "unauthorized",
		},
		{
			name:    "error - order not pending",
			userID:  1,
			orderID: 1,
			setupMock: func(m *MockOrderStore) {
				m.On("GetOrderByID", mock.Anything, int32(1)).Return(confirmedOrder, nil)
			},
			wantErr: true,
			errMsg:  "cannot be cancelled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockOrderStore)
			tt.setupMock(mockStore)

			service := &OrderService{store: createOrderStoreWrapper(mockStore)}

			resp, err := service.CancelOrder(context.Background(), tt.userID, tt.orderID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
		})
	}
}

func TestOrderService_GetUserOrders(t *testing.T) {
	t.Parallel()

	testOrder := createTestOrder()

	tests := []struct {
		name      string
		userID    int32
		page      int
		limit     int
		setupMock func(m *MockOrderStore)
		wantLen   int
		wantErr   bool
	}{
		{
			name:   "success - returns orders with pagination",
			userID: 1,
			page:   1,
			limit:  10,
			setupMock: func(m *MockOrderStore) {
				m.On("CountOrdersByUserID", mock.Anything, int32(1)).Return(int64(2), nil)
				m.On("ListOrdersByUserID", mock.Anything, mock.Anything).Return([]db.Order{testOrder, testOrder}, nil)
				m.On("ListOrderItems", mock.Anything, mock.Anything).Return([]db.OrderItem{}, nil)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:   "success - empty list",
			userID: 999,
			page:   1,
			limit:  10,
			setupMock: func(m *MockOrderStore) {
				m.On("CountOrdersByUserID", mock.Anything, int32(999)).Return(int64(0), nil)
				m.On("ListOrdersByUserID", mock.Anything, mock.Anything).Return([]db.Order{}, nil)
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:   "error - database error",
			userID: 1,
			page:   1,
			limit:  10,
			setupMock: func(m *MockOrderStore) {
				m.On("CountOrdersByUserID", mock.Anything, int32(1)).Return(int64(0), errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockOrderStore)
			tt.setupMock(mockStore)

			service := &OrderService{store: createOrderStoreWrapper(mockStore)}

			resp, meta, err := service.GetUserOrders(context.Background(), tt.userID, tt.page, tt.limit)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, resp, tt.wantLen)
			if tt.wantLen > 0 {
				assert.NotNil(t, meta)
			}
		})
	}
}

// orderStoreWrapper wraps MockOrderStore to implement db.Store interface
type orderStoreWrapper struct {
	*MockOrderStore
}

func createOrderStoreWrapper(m *MockOrderStore) db.Store {
	return &orderStoreWrapper{MockOrderStore: m}
}

// Implement remaining db.Store methods as stubs
func (s *orderStoreWrapper) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	return db.User{}, nil
}
func (s *orderStoreWrapper) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	return db.User{}, nil
}
func (s *orderStoreWrapper) GetUserByID(ctx context.Context, id int32) (db.User, error) {
	return db.User{}, nil
}
func (s *orderStoreWrapper) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	return db.User{}, nil
}
func (s *orderStoreWrapper) UpdateUserPassword(ctx context.Context, arg db.UpdateUserPasswordParams) error {
	return nil
}
func (s *orderStoreWrapper) UpdateUserRole(ctx context.Context, arg db.UpdateUserRoleParams) (db.User, error) {
	return db.User{}, nil
}
func (s *orderStoreWrapper) UpdateUserStatus(ctx context.Context, arg db.UpdateUserStatusParams) (db.User, error) {
	return db.User{}, nil
}
func (s *orderStoreWrapper) SoftDeleteUser(ctx context.Context, id int32) error { return nil }
func (s *orderStoreWrapper) ListUsers(ctx context.Context, arg db.ListUsersParams) ([]db.User, error) {
	return nil, nil
}
func (s *orderStoreWrapper) CountUsers(ctx context.Context) (int64, error) { return 0, nil }
func (s *orderStoreWrapper) CreateRefreshToken(ctx context.Context, arg db.CreateRefreshTokenParams) (db.RefreshToken, error) {
	return db.RefreshToken{}, nil
}
func (s *orderStoreWrapper) GetRefreshToken(ctx context.Context, token string) (db.RefreshToken, error) {
	return db.RefreshToken{}, nil
}
func (s *orderStoreWrapper) DeleteRefreshToken(ctx context.Context, token string) error { return nil }
func (s *orderStoreWrapper) DeleteRefreshTokensByUserID(ctx context.Context, userID int32) error {
	return nil
}
func (s *orderStoreWrapper) DeleteExpiredRefreshTokens(ctx context.Context) error { return nil }
func (s *orderStoreWrapper) GetRefreshTokensByUserID(ctx context.Context, userID int32) ([]db.RefreshToken, error) {
	return nil, nil
}
func (s *orderStoreWrapper) CreateCart(ctx context.Context, userID int32) (db.Cart, error) {
	return db.Cart{}, nil
}
func (s *orderStoreWrapper) GetCartByUserID(ctx context.Context, userID int32) (db.Cart, error) {
	return db.Cart{}, nil
}
func (s *orderStoreWrapper) GetCartByID(ctx context.Context, id int32) (db.Cart, error) {
	return db.Cart{}, nil
}
func (s *orderStoreWrapper) UpdateCartTimestamp(ctx context.Context, id int32) (db.Cart, error) {
	return db.Cart{}, nil
}
func (s *orderStoreWrapper) SoftDeleteCart(ctx context.Context, id int32) error { return nil }
func (s *orderStoreWrapper) SoftDeleteCartByUserID(ctx context.Context, userID int32) error {
	return nil
}
func (s *orderStoreWrapper) CreateCartItem(ctx context.Context, arg db.CreateCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *orderStoreWrapper) GetCartItem(ctx context.Context, arg db.GetCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *orderStoreWrapper) GetCartItemByID(ctx context.Context, id int32) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *orderStoreWrapper) ListCartItems(ctx context.Context, cartID int32) ([]db.CartItem, error) {
	return nil, nil
}
func (s *orderStoreWrapper) UpdateCartItemQuantity(ctx context.Context, arg db.UpdateCartItemQuantityParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *orderStoreWrapper) UpsertCartItem(ctx context.Context, arg db.UpsertCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *orderStoreWrapper) RestoreCartItem(ctx context.Context, arg db.RestoreCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *orderStoreWrapper) SoftDeleteCartItem(ctx context.Context, id int32) error { return nil }
func (s *orderStoreWrapper) SoftDeleteCartItemsByCartID(ctx context.Context, cartID int32) error {
	return nil
}
func (s *orderStoreWrapper) CountCartItems(ctx context.Context, cartID int32) (int64, error) {
	return 0, nil
}
func (s *orderStoreWrapper) CreateCategory(ctx context.Context, arg db.CreateCategoryParams) (db.Category, error) {
	return db.Category{}, nil
}
func (s *orderStoreWrapper) GetCategoryByID(ctx context.Context, id int32) (db.Category, error) {
	return db.Category{}, nil
}
func (s *orderStoreWrapper) ListCategories(ctx context.Context, arg db.ListCategoriesParams) ([]db.Category, error) {
	return nil, nil
}
func (s *orderStoreWrapper) ListActiveCategories(ctx context.Context) ([]db.Category, error) {
	return nil, nil
}
func (s *orderStoreWrapper) UpdateCategory(ctx context.Context, arg db.UpdateCategoryParams) (db.Category, error) {
	return db.Category{}, nil
}
func (s *orderStoreWrapper) UpdateCategoryStatus(ctx context.Context, arg db.UpdateCategoryStatusParams) (db.Category, error) {
	return db.Category{}, nil
}
func (s *orderStoreWrapper) SoftDeleteCategory(ctx context.Context, id int32) error { return nil }
func (s *orderStoreWrapper) CountCategories(ctx context.Context) (int64, error)     { return 0, nil }
func (s *orderStoreWrapper) CreateProduct(ctx context.Context, arg db.CreateProductParams) (db.Product, error) {
	return db.Product{}, nil
}
func (s *orderStoreWrapper) GetProductByID(ctx context.Context, id int32) (db.Product, error) {
	return db.Product{}, nil
}
func (s *orderStoreWrapper) GetProductByIDForUpdate(ctx context.Context, id int32) (db.Product, error) {
	return db.Product{}, nil
}
func (s *orderStoreWrapper) GetProductBySKU(ctx context.Context, sku string) (db.Product, error) {
	return db.Product{}, nil
}
func (s *orderStoreWrapper) GetProductsByIDsForUpdate(ctx context.Context, ids []int32) ([]db.Product, error) {
	return nil, nil
}
func (s *orderStoreWrapper) ListProducts(ctx context.Context, arg db.ListProductsParams) ([]db.Product, error) {
	return nil, nil
}
func (s *orderStoreWrapper) ListActiveProducts(ctx context.Context, arg db.ListActiveProductsParams) ([]db.Product, error) {
	return nil, nil
}
func (s *orderStoreWrapper) ListProductsByCategory(ctx context.Context, arg db.ListProductsByCategoryParams) ([]db.Product, error) {
	return nil, nil
}
func (s *orderStoreWrapper) UpdateProduct(ctx context.Context, arg db.UpdateProductParams) (db.Product, error) {
	return db.Product{}, nil
}
func (s *orderStoreWrapper) UpdateProductStatus(ctx context.Context, arg db.UpdateProductStatusParams) (db.Product, error) {
	return db.Product{}, nil
}
func (s *orderStoreWrapper) UpdateProductStock(ctx context.Context, arg db.UpdateProductStockParams) (db.Product, error) {
	return db.Product{}, nil
}
func (s *orderStoreWrapper) SoftDeleteProduct(ctx context.Context, id int32) error  { return nil }
func (s *orderStoreWrapper) CountProducts(ctx context.Context) (int64, error)       { return 0, nil }
func (s *orderStoreWrapper) CountActiveProducts(ctx context.Context) (int64, error) { return 0, nil }
func (s *orderStoreWrapper) CountProductsByCategory(ctx context.Context, categoryID int32) (int64, error) {
	return 0, nil
}
func (s *orderStoreWrapper) CreateProductImage(ctx context.Context, arg db.CreateProductImageParams) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *orderStoreWrapper) GetProductImageByID(ctx context.Context, id int32) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *orderStoreWrapper) GetPrimaryProductImage(ctx context.Context, productID int32) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *orderStoreWrapper) ListProductImages(ctx context.Context, productID int32) ([]db.ProductImage, error) {
	return nil, nil
}
func (s *orderStoreWrapper) UpdateProductImage(ctx context.Context, arg db.UpdateProductImageParams) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *orderStoreWrapper) SetPrimaryProductImage(ctx context.Context, arg db.SetPrimaryProductImageParams) error {
	return nil
}
func (s *orderStoreWrapper) SoftDeleteProductImage(ctx context.Context, id int32) error { return nil }
func (s *orderStoreWrapper) SoftDeleteProductImagesByProductID(ctx context.Context, productID int32) error {
	return nil
}
func (s *orderStoreWrapper) CreateOrder(ctx context.Context, arg db.CreateOrderParams) (db.Order, error) {
	return db.Order{}, nil
}
func (s *orderStoreWrapper) ListOrders(ctx context.Context, arg db.ListOrdersParams) ([]db.Order, error) {
	return nil, nil
}
func (s *orderStoreWrapper) ListOrdersByStatus(ctx context.Context, arg db.ListOrdersByStatusParams) ([]db.Order, error) {
	return nil, nil
}
func (s *orderStoreWrapper) UpdateOrderTotal(ctx context.Context, arg db.UpdateOrderTotalParams) (db.Order, error) {
	return db.Order{}, nil
}
func (s *orderStoreWrapper) SoftDeleteOrder(ctx context.Context, id int32) error { return nil }
func (s *orderStoreWrapper) CountOrders(ctx context.Context) (int64, error)      { return 0, nil }
func (s *orderStoreWrapper) CountOrdersByStatus(ctx context.Context, status db.NullOrderStatus) (int64, error) {
	return 0, nil
}
func (s *orderStoreWrapper) GetOrderTotal(ctx context.Context, orderID int32) (pgtype.Numeric, error) {
	return pgtype.Numeric{}, nil
}
func (s *orderStoreWrapper) CreateOrderItem(ctx context.Context, arg db.CreateOrderItemParams) (db.OrderItem, error) {
	return db.OrderItem{}, nil
}
func (s *orderStoreWrapper) GetOrderItemByID(ctx context.Context, id int32) (db.OrderItem, error) {
	return db.OrderItem{}, nil
}
func (s *orderStoreWrapper) SoftDeleteOrderItem(ctx context.Context, id int32) error { return nil }
func (s *orderStoreWrapper) SoftDeleteOrderItemsByOrderID(ctx context.Context, orderID int32) error {
	return nil
}
func (s *orderStoreWrapper) CountOrderItems(ctx context.Context, orderID int32) (int64, error) {
	return 0, nil
}
func (s *orderStoreWrapper) CreateIdempotencyKey(ctx context.Context, arg db.CreateIdempotencyKeyParams) (db.OrderIdempotencyKey, error) {
	return db.OrderIdempotencyKey{}, nil
}
func (s *orderStoreWrapper) GetIdempotencyKey(ctx context.Context, arg db.GetIdempotencyKeyParams) (db.OrderIdempotencyKey, error) {
	return db.OrderIdempotencyKey{}, nil
}
func (s *orderStoreWrapper) UpdateIdempotencyKeyOrderID(ctx context.Context, arg db.UpdateIdempotencyKeyOrderIDParams) error {
	return nil
}
func (s *orderStoreWrapper) SearchProducts(ctx context.Context, arg db.SearchProductsParams) ([]db.SearchProductsRow, error) {
	return nil, nil
}
func (s *orderStoreWrapper) CountSearchProducts(ctx context.Context, arg db.CountSearchProductsParams) (int64, error) {
	return 0, nil
}
