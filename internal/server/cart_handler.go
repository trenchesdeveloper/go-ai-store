package server

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trenchesdeveloper/go-ai-store/internal/dto"
	"github.com/trenchesdeveloper/go-ai-store/internal/services"
	"github.com/trenchesdeveloper/go-ai-store/internal/utils"
)

// GetCart godoc
// @Summary      Get user cart
// @Description  Get the authenticated user's shopping cart
// @Tags         cart
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.Response{data=dto.CartResponse}
// @Failure      500  {object}  utils.Response
// @Router       /cart [get]
func (s *Server) GetCart(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	cart, err := s.cartService.GetCart(ctx, int32(userID)) //#nosec G115 -- user ID from auth middleware
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to get cart", err)
		return
	}

	utils.SuccessResponse(ctx, "Cart retrieved successfully", cart)
}

// AddToCart godoc
// @Summary      Add item to cart
// @Description  Add a product to the user's cart
// @Tags         cart
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.AddToCartRequest true "Product and quantity"
// @Success      200  {object}  utils.Response{data=dto.CartResponse}
// @Failure      400  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /cart/items [post]
func (s *Server) AddToCart(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req dto.AddToCartRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(ctx, "Invalid request payload", err)
		return
	}

	cart, err := s.cartService.AddToCart(ctx, int32(userID), req) //#nosec G115 -- user ID from auth middleware
	if err != nil {
		switch {
		case errors.Is(err, services.ErrProductNotFound):
			utils.NotFoundResponse(ctx, "Product not found", err)
		case errors.Is(err, services.ErrInsufficientStock):
			utils.BadRequestResponse(ctx, "Insufficient stock", err)
		default:
			utils.InternalErrorResponse(ctx, "Failed to add to cart", err)
		}
		return
	}

	utils.SuccessResponse(ctx, "Item added to cart", cart)
}

// UpdateCartItem godoc
// @Summary      Update cart item quantity
// @Description  Update the quantity of an item in the cart
// @Tags         cart
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        itemId path int true "Cart Item ID"
// @Param        request body dto.UpdateCartItemRequest true "New quantity"
// @Success      200  {object}  utils.Response{data=dto.CartResponse}
// @Failure      400  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /cart/items/{itemId} [put]
func (s *Server) UpdateCartItem(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	itemIDStr := ctx.Param("itemId")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(ctx, "Invalid item ID", err)
		return
	}

	var req dto.UpdateCartItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(ctx, "Invalid request payload", err)
		return
	}

	cart, err := s.cartService.UpdateCartItem(ctx, int32(userID), int32(itemID), req) //#nosec G115 -- IDs from validated request
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCartNotFound):
			utils.NotFoundResponse(ctx, "Cart not found", err)
		case errors.Is(err, services.ErrCartItemNotFound):
			utils.NotFoundResponse(ctx, "Cart item not found", err)
		case errors.Is(err, services.ErrInsufficientStock):
			utils.BadRequestResponse(ctx, "Insufficient stock", err)
		default:
			utils.InternalErrorResponse(ctx, "Failed to update cart item", err)
		}
		return
	}

	utils.SuccessResponse(ctx, "Cart item updated", cart)
}

// RemoveCartItem godoc
// @Summary      Remove item from cart
// @Description  Remove a specific item from the cart
// @Tags         cart
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        itemId path int true "Cart Item ID"
// @Success      200  {object}  utils.Response{data=dto.CartResponse}
// @Failure      400  {object}  utils.Response
// @Failure      404  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /cart/items/{itemId} [delete]
func (s *Server) RemoveCartItem(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	itemIDStr := ctx.Param("itemId")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		utils.BadRequestResponse(ctx, "Invalid item ID", err)
		return
	}

	cart, err := s.cartService.RemoveCartItem(ctx, int32(userID), int32(itemID)) //#nosec G115 -- IDs from validated request
	if err != nil {
		switch {
		case errors.Is(err, services.ErrCartNotFound):
			utils.NotFoundResponse(ctx, "Cart not found", err)
		case errors.Is(err, services.ErrCartItemNotFound):
			utils.NotFoundResponse(ctx, "Cart item not found", err)
		default:
			utils.InternalErrorResponse(ctx, "Failed to remove cart item", err)
		}
		return
	}

	utils.SuccessResponse(ctx, "Cart item removed", cart)
}

// ClearCart godoc
// @Summary      Clear cart
// @Description  Remove all items from the cart
// @Tags         cart
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /cart [delete]
func (s *Server) ClearCart(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	err := s.cartService.ClearCart(ctx, int32(userID)) //#nosec G115 -- user ID from auth middleware
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to clear cart", err)
		return
	}

	utils.SuccessResponse(ctx, "Cart cleared", nil)
}
