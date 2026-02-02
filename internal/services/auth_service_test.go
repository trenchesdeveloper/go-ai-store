package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	db "github.com/trenchesdeveloper/go-ai-store/db/sqlc"
	"github.com/trenchesdeveloper/go-ai-store/internal/config"
	"github.com/trenchesdeveloper/go-ai-store/internal/dto"
	"github.com/trenchesdeveloper/go-ai-store/internal/utils"
)

// MockEventPublisher mocks the event publisher interface
type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) Publish(ctx context.Context, eventType string, data interface{}, headers map[string]string) error {
	args := m.Called(ctx, eventType, data, headers)
	return args.Error(0)
}

func (m *MockEventPublisher) Close() error {
	return nil
}

// MockAuthStore provides mocked store methods for AuthService
type MockAuthStore struct {
	mock.Mock
}

func (m *MockAuthStore) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockAuthStore) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockAuthStore) CreateCart(ctx context.Context, userID int32) (db.Cart, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(db.Cart), args.Error(1)
}

func (m *MockAuthStore) CreateRefreshToken(ctx context.Context, arg db.CreateRefreshTokenParams) (db.RefreshToken, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.RefreshToken), args.Error(1)
}

func (m *MockAuthStore) GetRefreshToken(ctx context.Context, token string) (db.RefreshToken, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(db.RefreshToken), args.Error(1)
}

func (m *MockAuthStore) DeleteRefreshToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockAuthStore) GetUserByID(ctx context.Context, id int32) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockAuthStore) ExecTx(ctx context.Context, fn func(*db.Queries) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

// Helper function to create a test config
func newAuthTestConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:                "test-secret-key-for-testing-purposes",
			ExpiresIn:             time.Hour,
			RefreshTokenExpiresIn: 24 * time.Hour,
		},
	}
}

// Helper to create test user
func createAuthTestUser(hashedPassword string) db.User {
	return db.User{
		ID:        1,
		Email:     "test@example.com",
		Password:  hashedPassword,
		FirstName: "John",
		LastName:  "Doe",
		Phone:     pgtype.Text{String: "1234567890", Valid: true},
		Role:      db.NullUserRole{UserRole: db.UserRoleCustomer, Valid: true},
		IsActive:  pgtype.Bool{Bool: true, Valid: true},
		CreatedAt: pgtype.Timestamptz{Valid: true},
		UpdatedAt: pgtype.Timestamptz{Valid: true},
	}
}

func TestAuthService_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		req       dto.RegisterRequest
		setupMock func(m *MockAuthStore)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "success - new user registered",
			req: dto.RegisterRequest{
				Email:     "newuser@example.com",
				Password:  "password123",
				FirstName: "New",
				LastName:  "User",
				Phone:     "1234567890",
			},
			setupMock: func(m *MockAuthStore) {
				// User doesn't exist
				m.On("GetUserByEmail", mock.Anything, "newuser@example.com").Return(db.User{}, pgx.ErrNoRows)
				// Create user succeeds
				m.On("CreateUser", mock.Anything, mock.MatchedBy(func(arg db.CreateUserParams) bool {
					return arg.Email == "newuser@example.com" && arg.FirstName == "New"
				})).Return(db.User{
					ID:        1,
					Email:     "newuser@example.com",
					FirstName: "New",
					LastName:  "User",
					Phone:     pgtype.Text{String: "1234567890", Valid: true},
					Role:      db.NullUserRole{UserRole: db.UserRoleCustomer, Valid: true},
					IsActive:  pgtype.Bool{Bool: true, Valid: true},
					CreatedAt: pgtype.Timestamptz{Valid: true},
					UpdatedAt: pgtype.Timestamptz{Valid: true},
				}, nil)
				// Create cart succeeds
				m.On("CreateCart", mock.Anything, int32(1)).Return(db.Cart{ID: 1, UserID: 1}, nil)
				// Create refresh token succeeds
				m.On("CreateRefreshToken", mock.Anything, mock.Anything).Return(db.RefreshToken{}, nil)
			},
			wantErr: false,
		},
		{
			name: "error - user already exists",
			req: dto.RegisterRequest{
				Email:     "existing@example.com",
				Password:  "password123",
				FirstName: "Existing",
				LastName:  "User",
			},
			setupMock: func(m *MockAuthStore) {
				m.On("GetUserByEmail", mock.Anything, "existing@example.com").Return(db.User{
					ID:    1,
					Email: "existing@example.com",
				}, nil)
			},
			wantErr: true,
			errMsg:  "user already exists",
		},
		{
			name: "error - database error during check",
			req: dto.RegisterRequest{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "Test",
				LastName:  "User",
			},
			setupMock: func(m *MockAuthStore) {
				m.On("GetUserByEmail", mock.Anything, "test@example.com").Return(db.User{}, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "error - create user fails",
			req: dto.RegisterRequest{
				Email:     "newuser@example.com",
				Password:  "password123",
				FirstName: "New",
				LastName:  "User",
			},
			setupMock: func(m *MockAuthStore) {
				m.On("GetUserByEmail", mock.Anything, "newuser@example.com").Return(db.User{}, pgx.ErrNoRows)
				m.On("CreateUser", mock.Anything, mock.Anything).Return(db.User{}, errors.New("create failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockAuthStore)
			mockPublisher := new(MockEventPublisher)
			tt.setupMock(mockStore)

			cfg := newAuthTestConfig()
			service := &AuthService{
				db:  createAuthStoreWrapper(mockStore),
				cfg: cfg,
				pub: mockPublisher,
			}

			resp, err := service.Register(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, resp.AccessToken)
			assert.NotEmpty(t, resp.RefreshToken)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	t.Parallel()

	// Create a hashed password for testing
	hashedPassword, _ := utils.HashPassword("correctpassword")
	testUser := createAuthTestUser(hashedPassword)

	tests := []struct {
		name      string
		req       dto.LoginRequest
		setupMock func(m *MockAuthStore, pub *MockEventPublisher)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "success - valid credentials",
			req: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "correctpassword",
			},
			setupMock: func(m *MockAuthStore, pub *MockEventPublisher) {
				m.On("GetUserByEmail", mock.Anything, "test@example.com").Return(testUser, nil)
				m.On("CreateRefreshToken", mock.Anything, mock.Anything).Return(db.RefreshToken{}, nil)
				pub.On("Publish", mock.Anything, "user_logged_in", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "error - user not found",
			req: dto.LoginRequest{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockAuthStore, pub *MockEventPublisher) {
				m.On("GetUserByEmail", mock.Anything, "notfound@example.com").Return(db.User{}, pgx.ErrNoRows)
			},
			wantErr: true,
			errMsg:  "invalid email or password",
		},
		{
			name: "error - wrong password",
			req: dto.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			setupMock: func(m *MockAuthStore, pub *MockEventPublisher) {
				m.On("GetUserByEmail", mock.Anything, "test@example.com").Return(testUser, nil)
			},
			wantErr: true,
			errMsg:  "invalid email or password",
		},
		{
			name: "error - inactive user",
			req: dto.LoginRequest{
				Email:    "inactive@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockAuthStore, pub *MockEventPublisher) {
				inactiveUser := testUser
				inactiveUser.IsActive = pgtype.Bool{Bool: false, Valid: true}
				m.On("GetUserByEmail", mock.Anything, "inactive@example.com").Return(inactiveUser, nil)
			},
			wantErr: true,
			errMsg:  "user is not active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockAuthStore)
			mockPublisher := new(MockEventPublisher)
			tt.setupMock(mockStore, mockPublisher)

			cfg := newAuthTestConfig()
			service := &AuthService{
				db:  createAuthStoreWrapper(mockStore),
				cfg: cfg,
				pub: mockPublisher,
			}

			resp, err := service.Login(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, resp.AccessToken)
			assert.NotEmpty(t, resp.RefreshToken)
			assert.Equal(t, testUser.Email, resp.User.Email)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		refreshToken string
		setupMock    func(m *MockAuthStore)
		wantErr      bool
	}{
		{
			name:         "success - token deleted",
			refreshToken: "valid-refresh-token",
			setupMock: func(m *MockAuthStore) {
				m.On("DeleteRefreshToken", mock.Anything, "valid-refresh-token").Return(nil)
			},
			wantErr: false,
		},
		{
			name:         "error - delete fails",
			refreshToken: "invalid-token",
			setupMock: func(m *MockAuthStore) {
				m.On("DeleteRefreshToken", mock.Anything, "invalid-token").Return(pgx.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockAuthStore)
			tt.setupMock(mockStore)

			cfg := newAuthTestConfig()
			service := &AuthService{
				db:  createAuthStoreWrapper(mockStore),
				cfg: cfg,
			}

			err := service.Logout(context.Background(), tt.refreshToken)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	t.Parallel()

	cfg := newAuthTestConfig()
	// Generate a valid refresh token for testing
	_, validRefreshToken, _ := utils.GenerateTokenPair(cfg, 1, "test@example.com", "customer")

	testUser := db.User{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Role:      db.NullUserRole{UserRole: db.UserRoleCustomer, Valid: true},
		IsActive:  pgtype.Bool{Bool: true, Valid: true},
	}

	tests := []struct {
		name      string
		req       dto.RefreshTokenRequest
		setupMock func(m *MockAuthStore)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "success - token refreshed",
			req:  dto.RefreshTokenRequest{RefreshToken: validRefreshToken},
			setupMock: func(m *MockAuthStore) {
				m.On("GetRefreshToken", mock.Anything, validRefreshToken).Return(db.RefreshToken{
					Token:     validRefreshToken,
					UserID:    1,
					ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
				}, nil)
				m.On("GetUserByID", mock.Anything, int32(1)).Return(testUser, nil)
				m.On("DeleteRefreshToken", mock.Anything, validRefreshToken).Return(nil)
				m.On("CreateRefreshToken", mock.Anything, mock.Anything).Return(db.RefreshToken{}, nil)
			},
			wantErr: false,
		},
		{
			name: "error - invalid JWT token",
			req:  dto.RefreshTokenRequest{RefreshToken: "invalid-jwt"},
			setupMock: func(m *MockAuthStore) {
				// No mock needed - JWT validation fails first
			},
			wantErr: true,
			errMsg:  "invalid refresh token",
		},
		{
			name: "error - token not found in database",
			req:  dto.RefreshTokenRequest{RefreshToken: validRefreshToken},
			setupMock: func(m *MockAuthStore) {
				m.On("GetRefreshToken", mock.Anything, validRefreshToken).Return(db.RefreshToken{}, pgx.ErrNoRows)
			},
			wantErr: true,
			errMsg:  "invalid refresh token",
		},
		{
			name: "error - token expired",
			req:  dto.RefreshTokenRequest{RefreshToken: validRefreshToken},
			setupMock: func(m *MockAuthStore) {
				m.On("GetRefreshToken", mock.Anything, validRefreshToken).Return(db.RefreshToken{
					Token:     validRefreshToken,
					UserID:    1,
					ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour), Valid: true}, // Expired
				}, nil)
			},
			wantErr: true,
			errMsg:  "invalid refresh token",
		},
		{
			name: "error - user not found",
			req:  dto.RefreshTokenRequest{RefreshToken: validRefreshToken},
			setupMock: func(m *MockAuthStore) {
				m.On("GetRefreshToken", mock.Anything, validRefreshToken).Return(db.RefreshToken{
					Token:     validRefreshToken,
					UserID:    1,
					ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
				}, nil)
				m.On("GetUserByID", mock.Anything, int32(1)).Return(db.User{}, pgx.ErrNoRows)
			},
			wantErr: true,
			errMsg:  "user not found",
		},
		{
			name: "error - user inactive",
			req:  dto.RefreshTokenRequest{RefreshToken: validRefreshToken},
			setupMock: func(m *MockAuthStore) {
				m.On("GetRefreshToken", mock.Anything, validRefreshToken).Return(db.RefreshToken{
					Token:     validRefreshToken,
					UserID:    1,
					ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
				}, nil)
				inactiveUser := testUser
				inactiveUser.IsActive = pgtype.Bool{Bool: false, Valid: true}
				m.On("GetUserByID", mock.Anything, int32(1)).Return(inactiveUser, nil)
			},
			wantErr: true,
			errMsg:  "user is not active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStore := new(MockAuthStore)
			mockPublisher := new(MockEventPublisher)
			tt.setupMock(mockStore)

			service := &AuthService{
				db:  createAuthStoreWrapper(mockStore),
				cfg: cfg,
				pub: mockPublisher,
			}

			resp, err := service.RefreshToken(context.Background(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, resp.AccessToken)
			assert.NotEmpty(t, resp.RefreshToken)
		})
	}
}

// authStoreWrapper wraps MockAuthStore to implement db.Store interface
type authStoreWrapper struct {
	*MockAuthStore
}

func createAuthStoreWrapper(m *MockAuthStore) db.Store {
	return &authStoreWrapper{MockAuthStore: m}
}

// Implement remaining db.Store methods as stubs
func (s *authStoreWrapper) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	return db.User{}, nil
}
func (s *authStoreWrapper) UpdateUserPassword(ctx context.Context, arg db.UpdateUserPasswordParams) error {
	return nil
}
func (s *authStoreWrapper) UpdateUserRole(ctx context.Context, arg db.UpdateUserRoleParams) (db.User, error) {
	return db.User{}, nil
}
func (s *authStoreWrapper) UpdateUserStatus(ctx context.Context, arg db.UpdateUserStatusParams) (db.User, error) {
	return db.User{}, nil
}
func (s *authStoreWrapper) SoftDeleteUser(ctx context.Context, id int32) error { return nil }
func (s *authStoreWrapper) ListUsers(ctx context.Context, arg db.ListUsersParams) ([]db.User, error) {
	return nil, nil
}
func (s *authStoreWrapper) CountUsers(ctx context.Context) (int64, error) { return 0, nil }
func (s *authStoreWrapper) DeleteRefreshTokensByUserID(ctx context.Context, u int32) error {
	return nil
}
func (s *authStoreWrapper) DeleteExpiredRefreshTokens(ctx context.Context) error { return nil }
func (s *authStoreWrapper) GetRefreshTokensByUserID(ctx context.Context, u int32) ([]db.RefreshToken, error) {
	return nil, nil
}
func (s *authStoreWrapper) GetCartByUserID(ctx context.Context, u int32) (db.Cart, error) {
	return db.Cart{}, nil
}
func (s *authStoreWrapper) GetCartByID(ctx context.Context, id int32) (db.Cart, error) {
	return db.Cart{}, nil
}
func (s *authStoreWrapper) UpdateCartTimestamp(ctx context.Context, id int32) (db.Cart, error) {
	return db.Cart{}, nil
}
func (s *authStoreWrapper) SoftDeleteCart(ctx context.Context, id int32) error        { return nil }
func (s *authStoreWrapper) SoftDeleteCartByUserID(ctx context.Context, u int32) error { return nil }
func (s *authStoreWrapper) CreateCartItem(ctx context.Context, arg db.CreateCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *authStoreWrapper) GetCartItem(ctx context.Context, arg db.GetCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *authStoreWrapper) GetCartItemByID(ctx context.Context, id int32) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *authStoreWrapper) ListCartItems(ctx context.Context, c int32) ([]db.CartItem, error) {
	return nil, nil
}
func (s *authStoreWrapper) UpdateCartItemQuantity(ctx context.Context, arg db.UpdateCartItemQuantityParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *authStoreWrapper) UpsertCartItem(ctx context.Context, arg db.UpsertCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *authStoreWrapper) RestoreCartItem(ctx context.Context, arg db.RestoreCartItemParams) (db.CartItem, error) {
	return db.CartItem{}, nil
}
func (s *authStoreWrapper) SoftDeleteCartItem(ctx context.Context, id int32) error { return nil }
func (s *authStoreWrapper) SoftDeleteCartItemsByCartID(ctx context.Context, c int32) error {
	return nil
}
func (s *authStoreWrapper) CountCartItems(ctx context.Context, c int32) (int64, error) { return 0, nil }
func (s *authStoreWrapper) CreateCategory(ctx context.Context, arg db.CreateCategoryParams) (db.Category, error) {
	return db.Category{}, nil
}
func (s *authStoreWrapper) GetCategoryByID(ctx context.Context, id int32) (db.Category, error) {
	return db.Category{}, nil
}
func (s *authStoreWrapper) GetCategoriesByIDs(ctx context.Context, ids []int32) ([]db.Category, error) {
	return nil, nil
}
func (s *authStoreWrapper) ListCategories(ctx context.Context, arg db.ListCategoriesParams) ([]db.Category, error) {
	return nil, nil
}
func (s *authStoreWrapper) ListActiveCategories(ctx context.Context) ([]db.Category, error) {
	return nil, nil
}
func (s *authStoreWrapper) UpdateCategory(ctx context.Context, arg db.UpdateCategoryParams) (db.Category, error) {
	return db.Category{}, nil
}
func (s *authStoreWrapper) UpdateCategoryStatus(ctx context.Context, arg db.UpdateCategoryStatusParams) (db.Category, error) {
	return db.Category{}, nil
}
func (s *authStoreWrapper) SoftDeleteCategory(ctx context.Context, id int32) error { return nil }
func (s *authStoreWrapper) CountCategories(ctx context.Context) (int64, error)     { return 0, nil }
func (s *authStoreWrapper) CreateProduct(ctx context.Context, arg db.CreateProductParams) (db.Product, error) {
	return db.Product{}, nil
}
func (s *authStoreWrapper) GetProductByID(ctx context.Context, id int32) (db.Product, error) {
	return db.Product{}, nil
}
func (s *authStoreWrapper) GetProductByIDForUpdate(ctx context.Context, id int32) (db.Product, error) {
	return db.Product{}, nil
}
func (s *authStoreWrapper) GetProductBySKU(ctx context.Context, sku string) (db.Product, error) {
	return db.Product{}, nil
}
func (s *authStoreWrapper) GetProductsByIDs(ctx context.Context, ids []int32) ([]db.Product, error) {
	return nil, nil
}
func (s *authStoreWrapper) GetProductsByIDsForUpdate(ctx context.Context, ids []int32) ([]db.Product, error) {
	return nil, nil
}
func (s *authStoreWrapper) ListProducts(ctx context.Context, arg db.ListProductsParams) ([]db.Product, error) {
	return nil, nil
}
func (s *authStoreWrapper) ListActiveProducts(ctx context.Context, arg db.ListActiveProductsParams) ([]db.Product, error) {
	return nil, nil
}
func (s *authStoreWrapper) ListProductsByCategory(ctx context.Context, arg db.ListProductsByCategoryParams) ([]db.Product, error) {
	return nil, nil
}
func (s *authStoreWrapper) UpdateProduct(ctx context.Context, arg db.UpdateProductParams) (db.Product, error) {
	return db.Product{}, nil
}
func (s *authStoreWrapper) UpdateProductStatus(ctx context.Context, arg db.UpdateProductStatusParams) (db.Product, error) {
	return db.Product{}, nil
}
func (s *authStoreWrapper) UpdateProductStock(ctx context.Context, arg db.UpdateProductStockParams) (db.Product, error) {
	return db.Product{}, nil
}
func (s *authStoreWrapper) SoftDeleteProduct(ctx context.Context, id int32) error  { return nil }
func (s *authStoreWrapper) CountProducts(ctx context.Context) (int64, error)       { return 0, nil }
func (s *authStoreWrapper) CountActiveProducts(ctx context.Context) (int64, error) { return 0, nil }
func (s *authStoreWrapper) CountProductsByCategory(ctx context.Context, c int32) (int64, error) {
	return 0, nil
}
func (s *authStoreWrapper) CreateProductImage(ctx context.Context, arg db.CreateProductImageParams) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *authStoreWrapper) GetProductImageByID(ctx context.Context, id int32) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *authStoreWrapper) GetPrimaryProductImage(ctx context.Context, p int32) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *authStoreWrapper) ListProductImages(ctx context.Context, p int32) ([]db.ProductImage, error) {
	return nil, nil
}
func (s *authStoreWrapper) ListProductImagesByProductIDs(ctx context.Context, ids []int32) ([]db.ProductImage, error) {
	return nil, nil
}
func (s *authStoreWrapper) UpdateProductImage(ctx context.Context, arg db.UpdateProductImageParams) (db.ProductImage, error) {
	return db.ProductImage{}, nil
}
func (s *authStoreWrapper) SetPrimaryProductImage(ctx context.Context, arg db.SetPrimaryProductImageParams) error {
	return nil
}
func (s *authStoreWrapper) SoftDeleteProductImage(ctx context.Context, id int32) error { return nil }
func (s *authStoreWrapper) SoftDeleteProductImagesByProductID(ctx context.Context, p int32) error {
	return nil
}
func (s *authStoreWrapper) CreateOrder(ctx context.Context, arg db.CreateOrderParams) (db.Order, error) {
	return db.Order{}, nil
}
func (s *authStoreWrapper) GetOrderByID(ctx context.Context, id int32) (db.Order, error) {
	return db.Order{}, nil
}
func (s *authStoreWrapper) ListOrders(ctx context.Context, arg db.ListOrdersParams) ([]db.Order, error) {
	return nil, nil
}
func (s *authStoreWrapper) ListOrdersByUserID(ctx context.Context, arg db.ListOrdersByUserIDParams) ([]db.Order, error) {
	return nil, nil
}
func (s *authStoreWrapper) ListOrdersByStatus(ctx context.Context, arg db.ListOrdersByStatusParams) ([]db.Order, error) {
	return nil, nil
}
func (s *authStoreWrapper) UpdateOrderStatus(ctx context.Context, arg db.UpdateOrderStatusParams) (db.Order, error) {
	return db.Order{}, nil
}
func (s *authStoreWrapper) UpdateOrderTotal(ctx context.Context, arg db.UpdateOrderTotalParams) (db.Order, error) {
	return db.Order{}, nil
}
func (s *authStoreWrapper) SoftDeleteOrder(ctx context.Context, id int32) error { return nil }
func (s *authStoreWrapper) CountOrders(ctx context.Context) (int64, error)      { return 0, nil }
func (s *authStoreWrapper) CountOrdersByUserID(ctx context.Context, u int32) (int64, error) {
	return 0, nil
}
func (s *authStoreWrapper) CountOrdersByStatus(ctx context.Context, s2 db.NullOrderStatus) (int64, error) {
	return 0, nil
}
func (s *authStoreWrapper) GetOrderTotal(ctx context.Context, o int32) (pgtype.Numeric, error) {
	return pgtype.Numeric{}, nil
}
func (s *authStoreWrapper) CreateOrderItem(ctx context.Context, arg db.CreateOrderItemParams) (db.OrderItem, error) {
	return db.OrderItem{}, nil
}
func (s *authStoreWrapper) GetOrderItemByID(ctx context.Context, id int32) (db.OrderItem, error) {
	return db.OrderItem{}, nil
}
func (s *authStoreWrapper) ListOrderItems(ctx context.Context, o int32) ([]db.OrderItem, error) {
	return nil, nil
}
func (s *authStoreWrapper) SoftDeleteOrderItem(ctx context.Context, id int32) error { return nil }
func (s *authStoreWrapper) SoftDeleteOrderItemsByOrderID(ctx context.Context, o int32) error {
	return nil
}
func (s *authStoreWrapper) CountOrderItems(ctx context.Context, o int32) (int64, error) {
	return 0, nil
}
func (s *authStoreWrapper) CreateIdempotencyKey(ctx context.Context, arg db.CreateIdempotencyKeyParams) (db.OrderIdempotencyKey, error) {
	return db.OrderIdempotencyKey{}, nil
}
func (s *authStoreWrapper) GetIdempotencyKey(ctx context.Context, arg db.GetIdempotencyKeyParams) (db.OrderIdempotencyKey, error) {
	return db.OrderIdempotencyKey{}, nil
}
func (s *authStoreWrapper) UpdateIdempotencyKeyOrderID(ctx context.Context, arg db.UpdateIdempotencyKeyOrderIDParams) error {
	return nil
}
func (s *authStoreWrapper) SearchProducts(ctx context.Context, arg db.SearchProductsParams) ([]db.SearchProductsRow, error) {
	return nil, nil
}
func (s *authStoreWrapper) CountSearchProducts(ctx context.Context, arg db.CountSearchProductsParams) (int64, error) {
	return 0, nil
}
