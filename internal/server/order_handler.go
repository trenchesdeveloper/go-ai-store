package server

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trenchesdeveloper/go-ai-store/internal/services"
	"github.com/trenchesdeveloper/go-ai-store/internal/utils"
)

// CreateOrder creates a new order from the user's cart
// Accepts X-Idempotency-Key header to prevent duplicate submissions
func (s *Server) CreateOrder(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	// Get idempotency key from header (optional)
	idempotencyKey := ctx.GetHeader("X-Idempotency-Key")

	order, err := s.orderService.CreateOrderWithIdempotency(ctx, int32(userID), idempotencyKey) //#nosec G115 -- user ID from auth middleware
	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmptyCart):
			utils.BadRequestResponse(ctx, "Cart is empty", err)
		case errors.Is(err, services.ErrProductNotFound):
			utils.NotFoundResponse(ctx, "Product not found", err)
		case errors.Is(err, services.ErrInsufficientStock):
			utils.BadRequestResponse(ctx, "Insufficient stock", err)
		case errors.Is(err, services.ErrDuplicateOrder):
			utils.BadRequestResponse(ctx, "Duplicate order submission - please wait", err)
		default:
			utils.InternalErrorResponse(ctx, "Failed to create order", err)
		}
		return
	}

	utils.CreatedResponse(ctx, "Order created successfully", order)
}

// GetOrder retrieves a single order by ID
func (s *Server) GetOrder(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	orderIDStr := ctx.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(ctx, "Invalid order ID", err)
		return
	}

	// Check if user is admin
	isAdmin := false
	if role, exists := ctx.Get("role"); exists {
		isAdmin = role.(string) == "admin"
	}

	order, err := s.orderService.GetOrderByID(ctx, int32(userID), int32(orderID), isAdmin) //#nosec G115 -- IDs from validated request
	if err != nil {
		switch {
		case errors.Is(err, services.ErrOrderNotFound):
			utils.NotFoundResponse(ctx, "Order not found", err)
		case errors.Is(err, services.ErrUnauthorizedOrder):
			utils.ForbiddenResponse(ctx, "You don't have access to this order", err)
		default:
			utils.InternalErrorResponse(ctx, "Failed to get order", err)
		}
		return
	}

	utils.SuccessResponse(ctx, "Order retrieved successfully", order)
}

// GetOrders retrieves the user's orders with pagination
func (s *Server) GetOrders(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	// Parse pagination parameters
	page := 1
	limit := 10

	if pageStr := ctx.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	orders, pagination, err := s.orderService.GetUserOrders(ctx, int32(userID), page, limit) //#nosec G115 -- user ID from auth middleware
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to get orders", err)
		return
	}

	utils.PaginatedSuccessResponse(ctx, "Orders retrieved successfully", orders, *pagination)
}

// CancelOrder cancels a pending order
func (s *Server) CancelOrder(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	orderIDStr := ctx.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(ctx, "Invalid order ID", err)
		return
	}

	order, err := s.orderService.CancelOrder(ctx, int32(userID), int32(orderID)) //#nosec G115 -- IDs from validated request
	if err != nil {
		switch {
		case errors.Is(err, services.ErrOrderNotFound):
			utils.NotFoundResponse(ctx, "Order not found", err)
		case errors.Is(err, services.ErrUnauthorizedOrder):
			utils.ForbiddenResponse(ctx, "You don't have access to this order", err)
		case errors.Is(err, services.ErrOrderNotCancellable):
			utils.BadRequestResponse(ctx, "Order cannot be cancelled", err)
		default:
			utils.InternalErrorResponse(ctx, "Failed to cancel order", err)
		}
		return
	}

	utils.SuccessResponse(ctx, "Order cancelled successfully", order)
}

// UpdateOrderStatus updates the status of an order (admin only)
func (s *Server) UpdateOrderStatus(ctx *gin.Context) {
	orderIDStr := ctx.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(ctx, "Invalid order ID", err)
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(ctx, "Invalid request payload", err)
		return
	}

	order, err := s.orderService.UpdateOrderStatus(ctx, int32(orderID), req.Status) //#nosec G115 -- order ID from validated request
	if err != nil {
		switch {
		case errors.Is(err, services.ErrOrderNotFound):
			utils.NotFoundResponse(ctx, "Order not found", err)
		default:
			utils.BadRequestResponse(ctx, err.Error(), err)
		}
		return
	}

	utils.SuccessResponse(ctx, "Order status updated successfully", order)
}
