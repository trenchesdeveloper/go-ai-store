package graph

import (
	"github.com/trenchesdeveloper/go-ai-store/internal/services"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	AuthService    *services.AuthService
	UserService    *services.UserService
	ProductService *services.ProductService
	CartService    *services.CartService
	OrderService   *services.OrderService
}

// NewResolver creates a new resolver with all service dependencies
func NewResolver(
	authService *services.AuthService,
	userService *services.UserService,
	productService *services.ProductService,
	cartService *services.CartService,
	orderService *services.OrderService,
) *Resolver {
	return &Resolver{
		AuthService:    authService,
		UserService:    userService,
		ProductService: productService,
		CartService:    cartService,
		OrderService:   orderService,
	}
}
