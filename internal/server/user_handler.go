package server

import (
	"github.com/gin-gonic/gin"
	"github.com/trenchesdeveloper/go-ai-store/internal/dto"
	"github.com/trenchesdeveloper/go-ai-store/internal/services"
	"github.com/trenchesdeveloper/go-ai-store/internal/utils"
)

func (s *Server) GetProfile(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	userService := services.NewUserService(s.store)

	user, err := userService.GetProfile(ctx, userID)
	if err != nil {
		utils.NotFoundResponse(ctx, "User not found", err)
		return
	}

	utils.SuccessResponse(ctx, "User profile retrieved successfully", user)
}

func (s *Server) UpdateProfile(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	userService := services.NewUserService(s.store)

	var req dto.UpdateProfileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(ctx, "Invalid request payload", err)
		return
	}

	user, err := userService.UpdateProfile(ctx, userID, req)
	if err != nil {
		utils.InternalErrorResponse(ctx, "Failed to update user profile", err)
		return
	}

	utils.SuccessResponse(ctx, "User profile updated successfully", user)
}
