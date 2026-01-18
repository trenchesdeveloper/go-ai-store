package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	db "github.com/trenchesdeveloper/go-ai-store/db/sqlc"
	"github.com/trenchesdeveloper/go-ai-store/internal/dto"
)

var (
	ErrCartNotFound      = errors.New("cart not found")
	ErrCartItemNotFound  = errors.New("cart item not found")
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock")
)

type CartService struct {
	store db.Store
}

func NewCartService(store db.Store) *CartService {
	return &CartService{store: store}
}

// GetOrCreateCart gets the user's cart or creates a new one if it doesn't exist
func (s *CartService) GetOrCreateCart(ctx context.Context, userID int32) (*db.Cart, error) {
	cart, err := s.store.GetCartByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Create a new cart for the user
			cart, err = s.store.CreateCart(ctx, userID)
			if err != nil {
				return nil, fmt.Errorf("failed to create cart: %w", err)
			}
			return &cart, nil
		}
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}
	return &cart, nil
}

// GetCart returns the user's cart with all items and product details
func (s *CartService) GetCart(ctx context.Context, userID int32) (*dto.CartResponse, error) {
	cart, err := s.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.buildCartResponse(ctx, cart)
}

// AddToCart adds a product to the user's cart
func (s *CartService) AddToCart(ctx context.Context, userID int32, req dto.AddToCartRequest) (*dto.CartResponse, error) {
	// Get or create cart
	cart, err := s.GetOrCreateCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Verify product exists and has sufficient stock
	product, err := s.store.GetProductByID(ctx, int32(req.ProductID)) //#nosec G115 -- product ID from validated request
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if int(product.Stock.Int32) < req.Quantity {
		return nil, ErrInsufficientStock
	}

	// Check if item already exists in cart (active)
	existingItem, err := s.store.GetCartItem(ctx, db.GetCartItemParams{
		CartID:    cart.ID,
		ProductID: int32(req.ProductID), //#nosec G115 -- product ID from validated request
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing cart item: %w", err)
	}

	if err == nil {
		// Update existing active item quantity
		newQuantity := existingItem.Quantity + int32(req.Quantity) //#nosec G115 -- quantity is validated
		if int(product.Stock.Int32) < int(newQuantity) {
			return nil, ErrInsufficientStock
		}
		_, err = s.store.UpdateCartItemQuantity(ctx, db.UpdateCartItemQuantityParams{
			ID:       existingItem.ID,
			Quantity: newQuantity,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to update cart item: %w", err)
		}
	} else {
		// Try to restore a soft-deleted item first
		restoredItem, restoreErr := s.store.RestoreCartItem(ctx, db.RestoreCartItemParams{
			CartID:    cart.ID,
			ProductID: int32(req.ProductID), //#nosec G115 -- product ID from validated request
			Quantity:  int32(req.Quantity),  //#nosec G115 -- quantity is validated
		})
		if restoreErr == nil {
			// Item was restored, update quantity if needed
			_ = restoredItem // Successfully restored
		} else if errors.Is(restoreErr, pgx.ErrNoRows) {
			// No soft-deleted item exists, create new
			_, err = s.store.CreateCartItem(ctx, db.CreateCartItemParams{
				CartID:    cart.ID,
				ProductID: int32(req.ProductID), //#nosec G115 -- product ID from validated request
				Quantity:  int32(req.Quantity),  //#nosec G115 -- quantity is validated
			})
			if err != nil {
				return nil, fmt.Errorf("failed to create cart item: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to restore cart item: %w", restoreErr)
		}
	}

	// Update cart timestamp
	_, err = s.store.UpdateCartTimestamp(ctx, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update cart timestamp: %w", err)
	}

	return s.buildCartResponse(ctx, cart)
}

// UpdateCartItem updates the quantity of a cart item
func (s *CartService) UpdateCartItem(ctx context.Context, userID int32, itemID int32, req dto.UpdateCartItemRequest) (*dto.CartResponse, error) {
	// Get cart
	cart, err := s.store.GetCartByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCartNotFound
		}
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	// Get cart item and verify it belongs to this cart
	item, err := s.store.GetCartItemByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCartItemNotFound
		}
		return nil, fmt.Errorf("failed to get cart item: %w", err)
	}

	if item.CartID != cart.ID {
		return nil, ErrCartItemNotFound
	}

	// Verify product has sufficient stock
	product, err := s.store.GetProductByID(ctx, item.ProductID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	if int(product.Stock.Int32) < req.Quantity {
		return nil, ErrInsufficientStock
	}

	// Update quantity
	_, err = s.store.UpdateCartItemQuantity(ctx, db.UpdateCartItemQuantityParams{
		ID:       itemID,
		Quantity: int32(req.Quantity), //#nosec G115 -- quantity is validated
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update cart item: %w", err)
	}

	// Update cart timestamp
	_, err = s.store.UpdateCartTimestamp(ctx, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update cart timestamp: %w", err)
	}

	return s.buildCartResponse(ctx, &cart)
}

// RemoveCartItem removes an item from the cart
func (s *CartService) RemoveCartItem(ctx context.Context, userID int32, itemID int32) (*dto.CartResponse, error) {
	// Get cart
	cart, err := s.store.GetCartByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCartNotFound
		}
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	// Get cart item and verify it belongs to this cart
	item, err := s.store.GetCartItemByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCartItemNotFound
		}
		return nil, fmt.Errorf("failed to get cart item: %w", err)
	}

	if item.CartID != cart.ID {
		return nil, ErrCartItemNotFound
	}

	// Soft delete the cart item
	err = s.store.SoftDeleteCartItem(ctx, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete cart item: %w", err)
	}

	// Update cart timestamp
	_, err = s.store.UpdateCartTimestamp(ctx, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update cart timestamp: %w", err)
	}

	return s.buildCartResponse(ctx, &cart)
}

// ClearCart removes all items from the cart
func (s *CartService) ClearCart(ctx context.Context, userID int32) error {
	cart, err := s.store.GetCartByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // No cart means nothing to clear
		}
		return fmt.Errorf("failed to get cart: %w", err)
	}

	err = s.store.SoftDeleteCartItemsByCartID(ctx, cart.ID)
	if err != nil {
		return fmt.Errorf("failed to clear cart items: %w", err)
	}

	return nil
}

// buildCartResponse builds a complete cart response with product details
func (s *CartService) buildCartResponse(ctx context.Context, cart *db.Cart) (*dto.CartResponse, error) {
	items, err := s.store.ListCartItems(ctx, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list cart items: %w", err)
	}

	// Collect product IDs
	productIDs := make([]int32, len(items))
	for i, item := range items {
		productIDs[i] = item.ProductID
	}

	// Fetch products in batch if there are items
	productMap := make(map[int32]db.Product)
	if len(productIDs) > 0 {
		products, err := s.store.GetProductsByIDs(ctx, productIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get products: %w", err)
		}
		for _, p := range products {
			productMap[p.ID] = p
		}
	}

	// Collect unique category IDs from products
	categoryIDSet := make(map[int32]struct{})
	for _, product := range productMap {
		categoryIDSet[product.CategoryID] = struct{}{}
	}

	// Convert category ID set to slice
	categoryIDs := make([]int32, 0, len(categoryIDSet))
	for id := range categoryIDSet {
		categoryIDs = append(categoryIDs, id)
	}

	// Batch fetch categories
	categoryMap := make(map[int32]db.Category)
	if len(categoryIDs) > 0 {
		categories, err := s.store.GetCategoriesByIDs(ctx, categoryIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to get categories: %w", err)
		}
		for _, cat := range categories {
			categoryMap[cat.ID] = cat
		}
	}

	// Build response
	var totalPrice float64
	cartItems := make([]dto.CartItemResponse, len(items))

	for i, item := range items {
		product := productMap[item.ProductID]
		category := categoryMap[product.CategoryID]
		priceFloat, _ := product.Price.Float64Value()
		subtotal := priceFloat.Float64 * float64(item.Quantity)
		totalPrice += subtotal

		cartItems[i] = dto.CartItemResponse{
			ID: uint(item.ID), //#nosec G115 -- DB ID is always positive
			Product: dto.ProductResponse{
				ID:          uint(product.ID), //#nosec G115 -- DB ID is always positive
				Name:        product.Name,
				Description: product.Description.String,
				Price:       priceFloat.Float64,
				Stock:       int(product.Stock.Int32),
				CategoryID:  uint(product.CategoryID), //#nosec G115 -- DB ID is always positive
				SKU:         product.Sku,
				IsActive:    product.IsActive.Bool,
				Category: dto.CategoryResponse{
					ID:          int64(category.ID),
					Name:        category.Name,
					Description: category.Description.String,
					IsActive:    category.IsActive.Bool,
				},
			},
			Quantity: int(item.Quantity),
			Subtotal: subtotal,
		}
	}

	return &dto.CartResponse{
		ID:        uint(cart.ID),     //#nosec G115 -- DB ID is always positive
		UserID:    uint(cart.UserID), //#nosec G115 -- DB ID is always positive
		CartItems: cartItems,
		Total:     totalPrice,
	}, nil
}
