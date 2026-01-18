package server

import (
	"github.com/gin-gonic/gin"
	"github.com/trenchesdeveloper/go-ai-store/internal/dto"
	"github.com/trenchesdeveloper/go-ai-store/internal/utils"
)

// GetProfile godoc
// @Summary      Get user profile
// @Description  Get the authenticated user's profile
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.Response{data=dto.UserResponse}
// @Failure      404  {object}  utils.Response
// @Router       /user/profile [get]
func (s *Server) GetProfile(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	user, err := s.userService.GetProfile(ctx, userID)
	if err != nil {
		utils.NotFoundResponse(ctx, "User not found", err)
		return
	}

	utils.SuccessResponse(ctx, "User profile retrieved successfully", user)
}

// UpdateProfile godoc
// @Summary      Update user profile
// @Description  Update the authenticated user's profile
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.UpdateProfileRequest true "Profile update data"
// @Success      200  {object}  utils.Response{data=dto.UserResponse}
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /user/profile [put]
func (s *Server) UpdateProfile(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")

	var req dto.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(ctx, "Invalid request payload", err)
		return
	}

	user, err := s.userService.UpdateProfile(ctx, userID, req)
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to update user profile", err)
		return
	}

	utils.SuccessResponse(ctx, "User profile updated successfully", user)
}
