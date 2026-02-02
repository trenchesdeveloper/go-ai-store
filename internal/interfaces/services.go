package interfaces

import (
	"context"

	"github.com/trenchesdeveloper/go-ai-store/internal/dto"
	"github.com/trenchesdeveloper/go-ai-store/internal/utils"
)

// AuthServicer defines authentication service methods
type AuthServicer interface {
	Register(ctx context.Context, req dto.RegisterRequest) (dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (dto.AuthResponse, error)
	RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (dto.AuthResponse, error)
	Logout(ctx context.Context, refreshToken string) error
}

// UserServicer defines user management methods
type UserServicer interface {
	GetProfile(ctx context.Context, userID uint) (*dto.UserResponse, error)
	UpdateProfile(ctx context.Context, userID uint, req dto.UpdateProfileRequest) (*dto.UserResponse, error)
}

// ProductServicer defines product/category management methods
type ProductServicer interface {
	CreateCategory(ctx context.Context, req dto.CreateCategoryRequest) (*dto.CategoryResponse, error)
	GetCategories(ctx context.Context) ([]dto.CategoryResponse, error)
	GetCategoryByID(ctx context.Context, id uint) (*dto.CategoryResponse, error)
	UpdateCategory(ctx context.Context, req dto.UpdateCategoryRequest) (*dto.CategoryResponse, error)
	DeleteCategory(ctx context.Context, id uint) error
	CreateProduct(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetProducts(ctx context.Context, page, limit int) ([]dto.ProductResponse, *utils.PaginationMeta, error)
	SearchProducts(ctx context.Context, req dto.SearchProductsRequest) ([]dto.ProductSearchResult, *utils.PaginationMeta, error)
	GetProductByID(ctx context.Context, id uint) (*dto.ProductResponse, error)
	UpdateProductByID(ctx context.Context, id uint, req *dto.UpdateProductRequest) (*dto.ProductResponse, error)
	DeleteProductByID(ctx context.Context, id uint) error
	UpdateProductImage(ctx context.Context, productID uint, imageURL string, altText string) error
}

// CartServicer defines cart management methods
type CartServicer interface {
	GetCart(ctx context.Context, userID int32) (*dto.CartResponse, error)
	AddToCart(ctx context.Context, userID int32, req dto.AddToCartRequest) (*dto.CartResponse, error)
	UpdateCartItem(ctx context.Context, userID int32, itemID int32, req dto.UpdateCartItemRequest) (*dto.CartResponse, error)
	RemoveCartItem(ctx context.Context, userID int32, itemID int32) (*dto.CartResponse, error)
	ClearCart(ctx context.Context, userID int32) error
}

// OrderServicer defines order management methods
type OrderServicer interface {
	CreateOrderFromCart(ctx context.Context, userID int32) (*dto.OrderResponse, error)
	CreateOrderWithIdempotency(ctx context.Context, userID int32, idempotencyKey string) (*dto.OrderResponse, error)
	GetOrderByID(ctx context.Context, userID int32, orderID int32, isAdmin bool) (*dto.OrderResponse, error)
	GetUserOrders(ctx context.Context, userID int32, page, limit int) ([]dto.OrderResponse, *utils.PaginationMeta, error)
	UpdateOrderStatus(ctx context.Context, orderID int32, status string) (*dto.OrderResponse, error)
	CancelOrder(ctx context.Context, userID int32, orderID int32) (*dto.OrderResponse, error)
}
