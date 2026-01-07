package server

import (
	"github.com/gin-gonic/gin"
	"github.com/trenchesdeveloper/go-ai-store/internal/dto"
	"github.com/trenchesdeveloper/go-ai-store/internal/services"
	"github.com/trenchesdeveloper/go-ai-store/internal/utils"
)

func (s *Server) registerHandler(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// create auth service
	authService := services.NewAuthService(s.store, s.cfg)
	// call service
	resp, err := authService.Register(c.Request.Context(), req)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to register user", err)
		return
	}

	// send response
	utils.CreatedResponse(c, "User registered successfully", resp)
}

func (s *Server) loginHandler(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// create auth service
	authService := services.NewAuthService(s.store, s.cfg)
	// call service
	resp, err := authService.Login(c.Request.Context(), req)
	if err != nil {
		utils.UnauthorizedResponse(c, "Invalid email or password", err)
		return
	}

	// send response
	utils.SuccessResponse(c, "User logged in successfully", resp)
}

func (s *Server) refreshTokenHandler(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// create auth service
	authService := services.NewAuthService(s.store, s.cfg)
	// call service
	resp, err := authService.RefreshToken(c.Request.Context(), req)
	if err != nil {
		utils.UnauthorizedResponse(c, "Token refresh failed", err)
		return
	}

	// send response
	utils.SuccessResponse(c, "Token refreshed successfully", resp)
}

func (s *Server) logoutHandler(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	// create auth service
	authService := services.NewAuthService(s.store, s.cfg)
	// call service
	err := authService.Logout(c.Request.Context(), req.RefreshToken)
	if err != nil {
		utils.InternalErrorResponse(c, "Logout failed", err)
		return
	}

	// send response
	utils.SuccessResponse(c, "Logged out successfully", nil)
}
