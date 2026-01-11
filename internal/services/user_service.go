package services

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/trenchesdeveloper/go-ai-store/db/sqlc"
	"github.com/trenchesdeveloper/go-ai-store/internal/dto"
)

type UserService struct {
	store db.Store
}

func NewUserService(store db.Store) *UserService {
	return &UserService{store: store}
}


func (s *UserService) GetProfile(ctx context.Context, userID uint) (*dto.UserResponse, error) {
	user, err := s.store.GetUserByID(ctx, int32(userID))
	if err != nil {
		return nil, err
	}
	return &dto.UserResponse{
		ID:        int64(user.ID),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone.String,
		Role:      string(user.Role.UserRole),
		IsActive:  user.IsActive.Bool,
	}, nil
}


func (s *UserService) UpdateProfile(ctx context.Context, userID uint, req dto.UpdateProfileRequest) (*dto.UserResponse, error) {
	user, err := s.store.GetUserByID(ctx, int32(userID))
	if err != nil {
		return nil, err
	}

	// update user
	user, err = s.store.UpdateUser(ctx, db.UpdateUserParams{
		ID:        user.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     pgtype.Text{String: req.Phone, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return &dto.UserResponse{
		ID:        int64(user.ID),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone.String,
		Role:      string(user.Role.UserRole),
		IsActive:  user.IsActive.Bool,
	}, nil
}