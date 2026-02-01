package graph

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/trenchesdeveloper/go-ai-store/internal/utils"
)

// Context keys for user data
type contextKey string

const (
	userIDKey    contextKey = "user_id"
	userEmailKey contextKey = "user_email"
	userRoleKey  contextKey = "user_role"
)

// User represents the authenticated user from context
type User struct {
	ID    uint
	Email string
	Role  string
}

// Errors
var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden: admin access required")
)

// GetUserFromContext extracts the authenticated user from context
func GetUserFromContext(ctx context.Context) (*User, error) {
	userID, ok := ctx.Value(userIDKey).(uint)
	if !ok {
		return nil, ErrUnauthorized
	}

	email, _ := ctx.Value(userEmailKey).(string)
	role, _ := ctx.Value(userRoleKey).(string)

	return &User{
		ID:    userID,
		Email: email,
		Role:  role,
	}, nil
}

// MustGetUser extracts user from context or panics (use with error handling)
func MustGetUser(ctx context.Context) *User {
	user, err := GetUserFromContext(ctx)
	if err != nil {
		panic(err)
	}
	return user
}

// IsAdmin checks if the current user has admin role
func IsAdmin(ctx context.Context) bool {
	user, err := GetUserFromContext(ctx)
	if err != nil {
		return false
	}
	return user.Role == "admin"
}

// RequireAuth is a helper that returns ErrUnauthorized if user is not authenticated
func RequireAuth(ctx context.Context) (*User, error) {
	return GetUserFromContext(ctx)
}

// RequireAdmin is a helper that returns ErrForbidden if user is not admin
func RequireAdmin(ctx context.Context) (*User, error) {
	user, err := GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if user.Role != "admin" {
		return nil, ErrForbidden
	}
	return user, nil
}

// AuthMiddleware is an HTTP middleware that validates JWT and adds user to context
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			// If no auth header, continue without user (public queries still work)
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				next.ServeHTTP(w, r)
				return
			}

			tokenString := tokenParts[1]

			// Validate the token
			claims, err := utils.ValidateToken(tokenString, jwtSecret)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), userIDKey, uint(claims.UserID))
			ctx = context.WithValue(ctx, userEmailKey, claims.Email)
			ctx = context.WithValue(ctx, userRoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
