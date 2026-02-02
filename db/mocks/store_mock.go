package mocks

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
	db "github.com/trenchesdeveloper/go-ai-store/db/sqlc"
)

// MockStore is a mock implementation of db.Store for testing
type MockStore struct {
	mock.Mock
}

// Ensure MockStore implements db.Store
var _ db.Store = (*MockStore)(nil)

// ExecTx mocks the transaction execution
func (m *MockStore) ExecTx(ctx context.Context, fn func(*db.Queries) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

// User methods
func (m *MockStore) CreateUser(ctx context.Context, arg db.CreateUserParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockStore) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockStore) GetUserByID(ctx context.Context, id int32) (db.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockStore) UpdateUser(ctx context.Context, arg db.UpdateUserParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockStore) UpdateUserPassword(ctx context.Context, arg db.UpdateUserPasswordParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockStore) UpdateUserRole(ctx context.Context, arg db.UpdateUserRoleParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockStore) UpdateUserStatus(ctx context.Context, arg db.UpdateUserStatusParams) (db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.User), args.Error(1)
}

func (m *MockStore) SoftDeleteUser(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) ListUsers(ctx context.Context, arg db.ListUsersParams) ([]db.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.User), args.Error(1)
}

func (m *MockStore) CountUsers(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// Auth/Token methods
func (m *MockStore) CreateRefreshToken(ctx context.Context, arg db.CreateRefreshTokenParams) (db.RefreshToken, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.RefreshToken), args.Error(1)
}

func (m *MockStore) GetRefreshToken(ctx context.Context, token string) (db.RefreshToken, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(db.RefreshToken), args.Error(1)
}

func (m *MockStore) DeleteRefreshToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockStore) DeleteRefreshTokensByUserID(ctx context.Context, userID int32) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockStore) DeleteExpiredRefreshTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockStore) GetRefreshTokensByUserID(ctx context.Context, userID int32) ([]db.RefreshToken, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]db.RefreshToken), args.Error(1)
}

// Cart methods
func (m *MockStore) CreateCart(ctx context.Context, userID int32) (db.Cart, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(db.Cart), args.Error(1)
}

func (m *MockStore) GetCartByUserID(ctx context.Context, userID int32) (db.Cart, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(db.Cart), args.Error(1)
}

func (m *MockStore) GetCartByID(ctx context.Context, id int32) (db.Cart, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Cart), args.Error(1)
}

func (m *MockStore) UpdateCartTimestamp(ctx context.Context, id int32) (db.Cart, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Cart), args.Error(1)
}

func (m *MockStore) SoftDeleteCart(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) SoftDeleteCartByUserID(ctx context.Context, userID int32) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// Cart Item methods
func (m *MockStore) CreateCartItem(ctx context.Context, arg db.CreateCartItemParams) (db.CartItem, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.CartItem), args.Error(1)
}

func (m *MockStore) GetCartItem(ctx context.Context, arg db.GetCartItemParams) (db.CartItem, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.CartItem), args.Error(1)
}

func (m *MockStore) GetCartItemByID(ctx context.Context, id int32) (db.CartItem, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.CartItem), args.Error(1)
}

func (m *MockStore) ListCartItems(ctx context.Context, cartID int32) ([]db.CartItem, error) {
	args := m.Called(ctx, cartID)
	return args.Get(0).([]db.CartItem), args.Error(1)
}

func (m *MockStore) UpdateCartItemQuantity(ctx context.Context, arg db.UpdateCartItemQuantityParams) (db.CartItem, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.CartItem), args.Error(1)
}

func (m *MockStore) UpsertCartItem(ctx context.Context, arg db.UpsertCartItemParams) (db.CartItem, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.CartItem), args.Error(1)
}

func (m *MockStore) RestoreCartItem(ctx context.Context, arg db.RestoreCartItemParams) (db.CartItem, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.CartItem), args.Error(1)
}

func (m *MockStore) SoftDeleteCartItem(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) SoftDeleteCartItemsByCartID(ctx context.Context, cartID int32) error {
	args := m.Called(ctx, cartID)
	return args.Error(0)
}

func (m *MockStore) CountCartItems(ctx context.Context, cartID int32) (int64, error) {
	args := m.Called(ctx, cartID)
	return args.Get(0).(int64), args.Error(1)
}

// Category methods
func (m *MockStore) CreateCategory(ctx context.Context, arg db.CreateCategoryParams) (db.Category, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Category), args.Error(1)
}

func (m *MockStore) GetCategoryByID(ctx context.Context, id int32) (db.Category, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Category), args.Error(1)
}

func (m *MockStore) GetCategoriesByIDs(ctx context.Context, ids []int32) ([]db.Category, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]db.Category), args.Error(1)
}

func (m *MockStore) ListCategories(ctx context.Context, arg db.ListCategoriesParams) ([]db.Category, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.Category), args.Error(1)
}

func (m *MockStore) ListActiveCategories(ctx context.Context) ([]db.Category, error) {
	args := m.Called(ctx)
	return args.Get(0).([]db.Category), args.Error(1)
}

func (m *MockStore) UpdateCategory(ctx context.Context, arg db.UpdateCategoryParams) (db.Category, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Category), args.Error(1)
}

func (m *MockStore) UpdateCategoryStatus(ctx context.Context, arg db.UpdateCategoryStatusParams) (db.Category, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Category), args.Error(1)
}

func (m *MockStore) SoftDeleteCategory(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) CountCategories(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// Product methods
func (m *MockStore) CreateProduct(ctx context.Context, arg db.CreateProductParams) (db.Product, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Product), args.Error(1)
}

func (m *MockStore) GetProductByID(ctx context.Context, id int32) (db.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Product), args.Error(1)
}

func (m *MockStore) GetProductByIDForUpdate(ctx context.Context, id int32) (db.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Product), args.Error(1)
}

func (m *MockStore) GetProductBySKU(ctx context.Context, sku string) (db.Product, error) {
	args := m.Called(ctx, sku)
	return args.Get(0).(db.Product), args.Error(1)
}

func (m *MockStore) GetProductsByIDs(ctx context.Context, ids []int32) ([]db.Product, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]db.Product), args.Error(1)
}

func (m *MockStore) GetProductsByIDsForUpdate(ctx context.Context, ids []int32) ([]db.Product, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]db.Product), args.Error(1)
}

func (m *MockStore) ListProducts(ctx context.Context, arg db.ListProductsParams) ([]db.Product, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.Product), args.Error(1)
}

func (m *MockStore) ListActiveProducts(ctx context.Context, arg db.ListActiveProductsParams) ([]db.Product, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.Product), args.Error(1)
}

func (m *MockStore) ListProductsByCategory(ctx context.Context, arg db.ListProductsByCategoryParams) ([]db.Product, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.Product), args.Error(1)
}

func (m *MockStore) UpdateProduct(ctx context.Context, arg db.UpdateProductParams) (db.Product, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Product), args.Error(1)
}

func (m *MockStore) UpdateProductStatus(ctx context.Context, arg db.UpdateProductStatusParams) (db.Product, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Product), args.Error(1)
}

func (m *MockStore) UpdateProductStock(ctx context.Context, arg db.UpdateProductStockParams) (db.Product, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Product), args.Error(1)
}

func (m *MockStore) SoftDeleteProduct(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) CountProducts(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStore) CountActiveProducts(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStore) CountProductsByCategory(ctx context.Context, categoryID int32) (int64, error) {
	args := m.Called(ctx, categoryID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStore) SearchProducts(ctx context.Context, arg db.SearchProductsParams) ([]db.SearchProductsRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.SearchProductsRow), args.Error(1)
}

func (m *MockStore) CountSearchProducts(ctx context.Context, arg db.CountSearchProductsParams) (int64, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(int64), args.Error(1)
}

// Product Image methods
func (m *MockStore) CreateProductImage(ctx context.Context, arg db.CreateProductImageParams) (db.ProductImage, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.ProductImage), args.Error(1)
}

func (m *MockStore) GetProductImageByID(ctx context.Context, id int32) (db.ProductImage, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.ProductImage), args.Error(1)
}

func (m *MockStore) GetPrimaryProductImage(ctx context.Context, productID int32) (db.ProductImage, error) {
	args := m.Called(ctx, productID)
	return args.Get(0).(db.ProductImage), args.Error(1)
}

func (m *MockStore) ListProductImages(ctx context.Context, productID int32) ([]db.ProductImage, error) {
	args := m.Called(ctx, productID)
	return args.Get(0).([]db.ProductImage), args.Error(1)
}

func (m *MockStore) ListProductImagesByProductIDs(ctx context.Context, ids []int32) ([]db.ProductImage, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]db.ProductImage), args.Error(1)
}

func (m *MockStore) UpdateProductImage(ctx context.Context, arg db.UpdateProductImageParams) (db.ProductImage, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.ProductImage), args.Error(1)
}

func (m *MockStore) SetPrimaryProductImage(ctx context.Context, arg db.SetPrimaryProductImageParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockStore) SoftDeleteProductImage(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) SoftDeleteProductImagesByProductID(ctx context.Context, productID int32) error {
	args := m.Called(ctx, productID)
	return args.Error(0)
}

// Order methods
func (m *MockStore) CreateOrder(ctx context.Context, arg db.CreateOrderParams) (db.Order, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Order), args.Error(1)
}

func (m *MockStore) GetOrderByID(ctx context.Context, id int32) (db.Order, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.Order), args.Error(1)
}

func (m *MockStore) ListOrders(ctx context.Context, arg db.ListOrdersParams) ([]db.Order, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.Order), args.Error(1)
}

func (m *MockStore) ListOrdersByUserID(ctx context.Context, arg db.ListOrdersByUserIDParams) ([]db.Order, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.Order), args.Error(1)
}

func (m *MockStore) ListOrdersByStatus(ctx context.Context, arg db.ListOrdersByStatusParams) ([]db.Order, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]db.Order), args.Error(1)
}

func (m *MockStore) UpdateOrderStatus(ctx context.Context, arg db.UpdateOrderStatusParams) (db.Order, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Order), args.Error(1)
}

func (m *MockStore) UpdateOrderTotal(ctx context.Context, arg db.UpdateOrderTotalParams) (db.Order, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.Order), args.Error(1)
}

func (m *MockStore) SoftDeleteOrder(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) CountOrders(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStore) CountOrdersByUserID(ctx context.Context, userID int32) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStore) CountOrdersByStatus(ctx context.Context, status db.NullOrderStatus) (int64, error) {
	args := m.Called(ctx, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStore) GetOrderTotal(ctx context.Context, orderID int32) (pgtype.Numeric, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(pgtype.Numeric), args.Error(1)
}

// Order Item methods
func (m *MockStore) CreateOrderItem(ctx context.Context, arg db.CreateOrderItemParams) (db.OrderItem, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.OrderItem), args.Error(1)
}

func (m *MockStore) GetOrderItemByID(ctx context.Context, id int32) (db.OrderItem, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(db.OrderItem), args.Error(1)
}

func (m *MockStore) ListOrderItems(ctx context.Context, orderID int32) ([]db.OrderItem, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).([]db.OrderItem), args.Error(1)
}

func (m *MockStore) SoftDeleteOrderItem(ctx context.Context, id int32) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockStore) SoftDeleteOrderItemsByOrderID(ctx context.Context, orderID int32) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

func (m *MockStore) CountOrderItems(ctx context.Context, orderID int32) (int64, error) {
	args := m.Called(ctx, orderID)
	return args.Get(0).(int64), args.Error(1)
}

// Idempotency Key methods
func (m *MockStore) CreateIdempotencyKey(ctx context.Context, arg db.CreateIdempotencyKeyParams) (db.OrderIdempotencyKey, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.OrderIdempotencyKey), args.Error(1)
}

func (m *MockStore) GetIdempotencyKey(ctx context.Context, arg db.GetIdempotencyKeyParams) (db.OrderIdempotencyKey, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(db.OrderIdempotencyKey), args.Error(1)
}

func (m *MockStore) UpdateIdempotencyKeyOrderID(ctx context.Context, arg db.UpdateIdempotencyKeyOrderIDParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}
