package server

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trenchesdeveloper/go-ai-store/internal/services"
	"github.com/trenchesdeveloper/go-ai-store/internal/utils"
)

// CreateOrder godoc
// @Summary      Create order from cart
// @Description  Creates a new order from the user's cart. Supports idempotency via X-Idempotency-Key header.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        X-Idempotency-Key header string false "Idempotency key to prevent duplicate orders"
// @Success      201  {object}  utils.Response{data=dto.OrderResponse}
// @Failure      400  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /orders [post]
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

// GetOrder godoc
// @Summary      Get order by ID
// @Description  Get a single order by ID. Users can only access their own orders unless admin.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Order ID"
// @Success      200  {object}  utils.Response{data=dto.OrderResponse}
// @Failure      400  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /orders/{id} [get]
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

// GetOrders godoc
// @Summary      List user orders
// @Description  Get all orders for the authenticated user with pagination
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(10)
// @Success      200  {object}  utils.PaginatedResponse{data=[]dto.OrderResponse}
// @Failure      500  {object}  utils.Response
// @Router       /orders [get]
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

// CancelOrder godoc
// @Summary      Cancel order
// @Description  Cancel a pending order. Only pending orders can be cancelled.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Order ID"
// @Success      200  {object}  utils.Response{data=dto.OrderResponse}
// @Failure      400  {object}  utils.Response
// @Failure      403  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /orders/{id}/cancel [post]
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

// UpdateOrderStatus godoc
// @Summary      Update order status (Admin)
// @Description  Update the status of an order. Admin only.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "Order ID"
// @Param        request body object{status=string} true "New status (pending, confirmed, shipped, delivered, cancelled)"
// @Success      200  {object}  utils.Response{data=dto.OrderResponse}
// @Failure      400  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Router       /orders/{id}/status [put]
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
