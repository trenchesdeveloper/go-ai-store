package server

import (
	"github.com/gin-gonic/gin"
	"github.com/trenchesdeveloper/go-ai-store/internal/dto"
	"github.com/trenchesdeveloper/go-ai-store/internal/utils"
)

// registerHandler godoc
// @Summary      Register a new user
// @Description  Create a new user account
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "Registration details"
// @Success      201  {object}  utils.Response{data=dto.AuthResponse}
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /auth/register [post]
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

// loginHandler godoc
// @Summary      Login user
// @Description  Authenticate user and return tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "Login credentials"
// @Success      200  {object}  utils.Response{data=dto.AuthResponse}
// @Failure      400  {object}  utils.Response
// @Failure      401  {object}  utils.Response
// @Router       /auth/login [post]
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

// refreshTokenHandler godoc
// @Summary      Refresh access token
// @Description  Get a new access token using refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RefreshTokenRequest true "Refresh token"
// @Success      200  {object}  utils.Response{data=dto.AuthResponse}
// @Failure      400  {object}  utils.Response
// @Failure      401  {object}  utils.Response
// @Router       /auth/refresh-token [post]
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

// logoutHandler godoc
// @Summary      Logout user
// @Description  Invalidate the refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body dto.RefreshTokenRequest true "Refresh token to invalidate"
// @Success      200  {object}  utils.Response
// @Failure      400  {object}  utils.Response
// @Failure      500  {object}  utils.Response
// @Router       /auth/logout [post]
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
