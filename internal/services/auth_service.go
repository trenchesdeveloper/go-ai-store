package services

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/trenchesdeveloper/go-ai-store/db/sqlc"
	"github.com/trenchesdeveloper/go-ai-store/internal/config"
	"github.com/trenchesdeveloper/go-ai-store/internal/dto"
	"github.com/trenchesdeveloper/go-ai-store/internal/utils"
)

type AuthService struct {
	db  db.Store
	cfg *config.Config
}

func NewAuthService(db db.Store, cfg *config.Config) *AuthService {
	return &AuthService{
		db:  db,
		cfg: cfg,
	}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (dto.AuthResponse, error) {
	// check if user exist
	_, err := s.db.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return dto.AuthResponse{}, errors.New("user already exists")
	}

	// check for error other than user not found
	if !errors.Is(err, pgx.ErrNoRows) {
		return dto.AuthResponse{}, err
	}

	// hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return dto.AuthResponse{}, errors.New("something went wrong")
	}

	// create user
	user, err := s.db.CreateUser(ctx, db.CreateUserParams{
		Email:     req.Email,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     pgtype.Text{String: req.Phone, Valid: true},
	})
	if err != nil {
		return dto.AuthResponse{}, errors.New("something went wrong")
	}

	// create cart
	_, err = s.db.CreateCart(ctx, user.ID)
	if err != nil {
		return dto.AuthResponse{}, errors.New("something went wrong")
	}

	// call generateAuthResponse function
	return s.generateAuthResponse(ctx, &user)
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (dto.AuthResponse, error) {
	// check if user exist
	user, err := s.db.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return dto.AuthResponse{}, errors.New("Invalid email or password")
	}

	// check if the user is active
	if !user.IsActive.Bool || !user.IsActive.Valid {
		return dto.AuthResponse{}, errors.New("user is not active")
	}

	// check password
	if err := utils.VerifyPassword(user.Password, req.Password); err != nil {
		return dto.AuthResponse{}, errors.New("Invalid email or password")
	}

	// call generateAuthResponse function
	return s.generateAuthResponse(ctx, &user)
}

func (s *AuthService) RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (dto.AuthResponse, error) {
	// validate refresh token
	claims, err := utils.ValidateToken(req.RefreshToken, s.cfg.JWT.Secret)
	if err != nil {
		return dto.AuthResponse{}, errors.New("Invalid refresh token")
	}
	// check if refresh token exist
	refreshToken, err := s.db.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return dto.AuthResponse{}, errors.New("Invalid refresh token")
	}

	// check if the refresh token is expired
	if refreshToken.ExpiresAt.Time.Before(time.Now()) {
		return dto.AuthResponse{}, errors.New("Invalid refresh token")
	}

	// check if the refresh token is valid
	if refreshToken.Token != req.RefreshToken {
		return dto.AuthResponse{}, errors.New("Invalid refresh token")
	}

	// find user
	user, err := s.db.GetUserByID(ctx, int32(claims.UserID))
	if err != nil {
		return dto.AuthResponse{}, errors.New("User not found")
	}

	// check if the user is active
	if !user.IsActive.Bool || !user.IsActive.Valid {
		return dto.AuthResponse{}, errors.New("user is not active")
	}

	// delete old refresh token
	err = s.db.DeleteRefreshToken(ctx, refreshToken.Token)
	if err != nil {
		return dto.AuthResponse{}, errors.New("something went wrong")
	}

	// call generateAuthResponse function
	return s.generateAuthResponse(ctx, &user)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	// delete refresh token
	err := s.db.DeleteRefreshToken(ctx, refreshToken)
	if err != nil {
		return err
	}
	return nil
}

func (s *AuthService) generateAuthResponse(ctx context.Context, user *db.User) (dto.AuthResponse, error) {
	// generate tokens
	accessToken, refreshToken, err := utils.GenerateTokenPair(s.cfg, uint(user.ID), user.Email, string(user.Role.UserRole))
	if err != nil {
		return dto.AuthResponse{}, err
	}

	// save refresh token
	_, err = s.db.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(time.Duration(s.cfg.JWT.RefreshTokenExpiresIn) * time.Hour), Valid: true},
	})
	if err != nil {
		return dto.AuthResponse{}, err
	}

	return dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
