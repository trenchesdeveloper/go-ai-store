package server

import (
	"github.com/gin-gonic/gin"
	"github.com/trenchesdeveloper/go-ai-store/internal/dto"
	"github.com/trenchesdeveloper/go-ai-store/internal/utils"
)

func (s *Server) registerHandler(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	resp, err := s.authService.Register(c.Request.Context(), req)
	if err != nil {
		utils.InternalErrorResponse(c, "Failed to register user", err)
		return
	}

	utils.CreatedResponse(c, "User registered successfully", resp)
}

func (s *Server) loginHandler(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	resp, err := s.authService.Login(c.Request.Context(), req)
	if err != nil {
		utils.UnauthorizedResponse(c, "Invalid email or password", err)
		return
	}

	utils.SuccessResponse(c, "User logged in successfully", resp)
}

func (s *Server) refreshTokenHandler(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	resp, err := s.authService.RefreshToken(c.Request.Context(), req)
	if err != nil {
		utils.UnauthorizedResponse(c, "Token refresh failed", err)
		return
	}

	utils.SuccessResponse(c, "Token refreshed successfully", resp)
}

func (s *Server) logoutHandler(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequestResponse(c, "Invalid request data", err)
		return
	}

	err := s.authService.Logout(c.Request.Context(), req.RefreshToken)
	if err != nil {
		utils.InternalErrorResponse(c, "Logout failed", err)
		return
	}

	utils.SuccessResponse(c, "Logged out successfully", nil)
}
