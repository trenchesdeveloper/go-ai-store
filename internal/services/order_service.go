package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/trenchesdeveloper/go-ai-store/db/sqlc"
	"github.com/trenchesdeveloper/go-ai-store/internal/dto"
	"github.com/trenchesdeveloper/go-ai-store/internal/utils"
)

var (
	ErrOrderNotFound       = errors.New("order not found")
	ErrEmptyCart           = errors.New("cannot create order from empty cart")
	ErrOrderNotCancellable = errors.New("order cannot be cancelled")
	ErrUnauthorizedOrder   = errors.New("unauthorized access to order")
	ErrDuplicateOrder      = errors.New("duplicate order submission")
)

type OrderService struct {
	store       db.Store
	cartService *CartService
}

func NewOrderService(store db.Store, cartService *CartService) *OrderService {
	return &OrderService{
		store:       store,
		cartService: cartService,
	}
}

// CreateOrderWithIdempotency creates a new order with idempotency key support
// If the same idempotency key is provided twice, returns the existing order
func (s *OrderService) CreateOrderWithIdempotency(ctx context.Context, userID int32, idempotencyKey string) (*dto.OrderResponse, error) {
	// Check if order already exists for this idempotency key
	if idempotencyKey != "" {
		existingKey, err := s.store.GetIdempotencyKey(ctx, db.GetIdempotencyKeyParams{
			UserID:         userID,
			IdempotencyKey: idempotencyKey,
		})
		if err == nil && existingKey.OrderID.Valid {
			// Return existing order
			return s.GetOrderByID(ctx, userID, existingKey.OrderID.Int32, false)
		}
		// If key exists but no order_id, another request is in progress
		if err == nil && !existingKey.OrderID.Valid {
			return nil, ErrDuplicateOrder
		}
		// If not found, proceed to create
	}

	// Create the order using the existing logic
	orderResponse, err := s.CreateOrderFromCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Store the idempotency key with the order ID
	if idempotencyKey != "" {
		_, _ = s.store.CreateIdempotencyKey(ctx, db.CreateIdempotencyKeyParams{
			UserID:         userID,
			IdempotencyKey: idempotencyKey,
			OrderID: pgtype.Int4{
				Int32: int32(orderResponse.ID), //#nosec G115 -- order ID from DB
				Valid: true,
			},
		})
	}

	return orderResponse, nil
}

// CreateOrderFromCart creates a new order from the user's cart
// Uses database transaction with row-level locking to prevent race conditions
func (s *OrderService) CreateOrderFromCart(ctx context.Context, userID int32) (*dto.OrderResponse, error) {
	// Get the user's cart (outside transaction - read-only)
	cart, err := s.store.GetCartByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEmptyCart
		}
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}

	// Get cart items (outside transaction - read-only)
	cartItems, err := s.store.ListCartItems(ctx, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list cart items: %w", err)
	}

	if len(cartItems) == 0 {
		return nil, ErrEmptyCart
	}

	// Collect product IDs
	productIDs := make([]int32, len(cartItems))
	for i, item := range cartItems {
		productIDs[i] = item.ProductID
	}

	var order db.Order

	// Execute order creation within a transaction with row locking
	err = s.store.ExecTx(ctx, func(q *db.Queries) error {
		// Lock product rows with FOR UPDATE to prevent concurrent modifications
		products, err := q.GetProductsByIDsForUpdate(ctx, productIDs)
		if err != nil {
			return fmt.Errorf("failed to lock products: %w", err)
		}

		if len(products) != len(productIDs) {
			return ErrProductNotFound
		}

		productMap := make(map[int32]db.Product, len(products))
		for _, p := range products {
			productMap[p.ID] = p
		}

		// Validate stock and calculate total (under lock)
		var totalAmount float64
		for _, item := range cartItems {
			product, ok := productMap[item.ProductID]
			if !ok {
				return ErrProductNotFound
			}
			if int(product.Stock.Int32) < int(item.Quantity) {
				return ErrInsufficientStock
			}
			priceFloat, _ := product.Price.Float64Value()
			totalAmount += priceFloat.Float64 * float64(item.Quantity)
		}

		// Create order
		var totalNumeric pgtype.Numeric
		if err := totalNumeric.Scan(fmt.Sprintf("%.2f", totalAmount)); err != nil {
			return fmt.Errorf("failed to parse total amount: %w", err)
		}

		order, err = q.CreateOrder(ctx, db.CreateOrderParams{
			UserID:      userID,
			TotalAmount: totalNumeric,
		})
		if err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		// Create order items and update stock
		for _, item := range cartItems {
			product := productMap[item.ProductID]
			priceFloat, _ := product.Price.Float64Value()

			var priceNumeric pgtype.Numeric
			if err := priceNumeric.Scan(fmt.Sprintf("%.2f", priceFloat.Float64)); err != nil {
				return fmt.Errorf("failed to parse price: %w", err)
			}

			_, err := q.CreateOrderItem(ctx, db.CreateOrderItemParams{
				OrderID:   order.ID,
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
				Price:     priceNumeric,
			})
			if err != nil {
				return fmt.Errorf("failed to create order item: %w", err)
			}

			// Update product stock (still under lock)
			newStock := product.Stock.Int32 - item.Quantity
			_, err = q.UpdateProductStock(ctx, db.UpdateProductStockParams{
				ID: product.ID,
				Stock: pgtype.Int4{
					Int32: newStock,
					Valid: true,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to update product stock: %w", err)
			}
		}

		// Clear cart items within transaction
		err = q.SoftDeleteCartItemsByCartID(ctx, cart.ID)
		if err != nil {
			return fmt.Errorf("failed to clear cart: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Return the order response
	return s.buildOrderResponse(ctx, order)
}

// GetOrderByID retrieves an order by ID (validates ownership for non-admin users)
func (s *OrderService) GetOrderByID(ctx context.Context, userID int32, orderID int32, isAdmin bool) (*dto.OrderResponse, error) {
	order, err := s.store.GetOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Check ownership for non-admin users
	if !isAdmin && order.UserID != userID {
		return nil, ErrUnauthorizedOrder
	}

	return s.buildOrderResponse(ctx, order)
}

// GetUserOrders retrieves orders for a user with pagination
func (s *OrderService) GetUserOrders(ctx context.Context, userID int32, page, limit int) ([]dto.OrderResponse, *utils.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	// Get total count for pagination
	totalCount, err := s.store.CountOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count orders: %w", err)
	}

	// Calculate total pages
	totalPages := int(totalCount) / limit
	if int(totalCount)%limit > 0 {
		totalPages++
	}

	paginationMeta := &utils.PaginationMeta{
		Page:       page,
		Limit:      limit,
		TotalCount: int(totalCount),
		TotalPages: totalPages,
	}

	orders, err := s.store.ListOrdersByUserID(ctx, db.ListOrdersByUserIDParams{
		UserID: userID,
		Limit:  int32(limit),              //#nosec G115 -- pagination values are bounded
		Offset: int32((page - 1) * limit), //#nosec G115 -- pagination values are bounded
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list orders: %w", err)
	}

	if len(orders) == 0 {
		return []dto.OrderResponse{}, paginationMeta, nil
	}

	// Build responses for each order
	orderResponses := make([]dto.OrderResponse, len(orders))
	for i, order := range orders {
		response, err := s.buildOrderResponse(ctx, order)
		if err != nil {
			return nil, nil, err
		}
		orderResponses[i] = *response
	}

	return orderResponses, paginationMeta, nil
}

// UpdateOrderStatus updates the status of an order (admin function)
func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID int32, status string) (*dto.OrderResponse, error) {
	// Validate status
	orderStatus := db.OrderStatus(status)
	if !isValidOrderStatus(orderStatus) {
		return nil, fmt.Errorf("invalid order status: %s", status)
	}

	order, err := s.store.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID: orderID,
		Status: db.NullOrderStatus{
			OrderStatus: orderStatus,
			Valid:       true,
		},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	return s.buildOrderResponse(ctx, order)
}

// CancelOrder cancels an order (only pending orders can be cancelled)
func (s *OrderService) CancelOrder(ctx context.Context, userID int32, orderID int32) (*dto.OrderResponse, error) {
	order, err := s.store.GetOrderByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Check ownership
	if order.UserID != userID {
		return nil, ErrUnauthorizedOrder
	}

	// Only pending orders can be cancelled
	if order.Status.OrderStatus != db.OrderStatusPending {
		return nil, ErrOrderNotCancellable
	}

	// Update order status to cancelled
	updatedOrder, err := s.store.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID: orderID,
		Status: db.NullOrderStatus{
			OrderStatus: db.OrderStatusCancelled,
			Valid:       true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to cancel order: %w", err)
	}

	// Restore product stock
	orderItems, err := s.store.ListOrderItems(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to list order items: %w", err)
	}

	for _, item := range orderItems {
		product, err := s.store.GetProductByID(ctx, item.ProductID)
		if err != nil {
			continue // Skip if product not found
		}

		newStock := product.Stock.Int32 + item.Quantity
		_, err = s.store.UpdateProductStock(ctx, db.UpdateProductStockParams{
			ID: product.ID,
			Stock: pgtype.Int4{
				Int32: newStock,
				Valid: true,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to restore product stock: %w", err)
		}
	}

	return s.buildOrderResponse(ctx, updatedOrder)
}

// buildOrderResponse builds an OrderResponse with order items and product details
func (s *OrderService) buildOrderResponse(ctx context.Context, order db.Order) (*dto.OrderResponse, error) {
	orderItems, err := s.store.ListOrderItems(ctx, order.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list order items: %w", err)
	}

	// Collect product IDs
	productIDs := make([]int32, len(orderItems))
	for i, item := range orderItems {
		productIDs[i] = item.ProductID
	}

	// Fetch products in batch
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

	// Collect unique category IDs
	categoryIDSet := make(map[int32]struct{})
	for _, product := range productMap {
		categoryIDSet[product.CategoryID] = struct{}{}
	}

	categoryIDs := make([]int32, 0, len(categoryIDSet))
	for id := range categoryIDSet {
		categoryIDs = append(categoryIDs, id)
	}

	// Fetch categories in batch
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

	// Build order item responses
	orderItemResponses := make([]dto.OrderItemResponse, len(orderItems))
	for i, item := range orderItems {
		product := productMap[item.ProductID]
		category := categoryMap[product.CategoryID]
		priceFloat, _ := item.Price.Float64Value()
		productPriceFloat, _ := product.Price.Float64Value()

		orderItemResponses[i] = dto.OrderItemResponse{
			ID: uint(item.ID), //#nosec G115 -- DB ID is always positive
			Product: dto.ProductResponse{
				ID:          uint(product.ID), //#nosec G115 -- DB ID is always positive
				Name:        product.Name,
				Description: product.Description.String,
				Price:       productPriceFloat.Float64,
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
				Images: []dto.ProductImageResponse{},
			},
			Quantity: int(item.Quantity),
			Price:    priceFloat.Float64,
		}
	}

	totalFloat, _ := order.TotalAmount.Float64Value()

	return &dto.OrderResponse{
		ID:          uint(order.ID),     //#nosec G115 -- DB ID is always positive
		UserID:      uint(order.UserID), //#nosec G115 -- DB ID is always positive
		Status:      string(order.Status.OrderStatus),
		TotalAmount: totalFloat.Float64,
		OrderItems:  orderItemResponses,
		CreatedAt:   order.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

func isValidOrderStatus(status db.OrderStatus) bool {
	switch status {
	case db.OrderStatusPending, db.OrderStatusConfirmed, db.OrderStatusShipped, db.OrderStatusDelivered, db.OrderStatusCancelled:
		return true
	default:
		return false
	}
}
