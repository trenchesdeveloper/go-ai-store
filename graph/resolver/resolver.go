package graph

import (
	"github.com/trenchesdeveloper/go-ai-store/internal/interfaces"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	AuthService    interfaces.AuthServicer
	UserService    interfaces.UserServicer
	ProductService interfaces.ProductServicer
	CartService    interfaces.CartServicer
	OrderService   interfaces.OrderServicer
}

// NewResolver creates a new resolver with all service dependencies
func NewResolver(
	authService interfaces.AuthServicer,
	userService interfaces.UserServicer,
	productService interfaces.ProductServicer,
	cartService interfaces.CartServicer,
	orderService interfaces.OrderServicer,
) *Resolver {
	return &Resolver{
		AuthService:    authService,
		UserService:    userService,
		ProductService: productService,
		CartService:    cartService,
		OrderService:   orderService,
	}
}
