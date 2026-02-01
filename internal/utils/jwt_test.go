package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trenchesdeveloper/go-ai-store/internal/config"
)

func newTestConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:                "test-secret-key-for-testing-purposes-only",
			ExpiresIn:             time.Hour,
			RefreshTokenExpiresIn: 24 * time.Hour,
		},
	}
}

func TestGenerateTokenPair(t *testing.T) {
	t.Parallel()

	cfg := newTestConfig()

	tests := []struct {
		name    string
		userID  uint
		email   string
		role    string
		wantErr bool
	}{
		{
			name:    "valid user token",
			userID:  1,
			email:   "user@example.com",
			role:    "user",
			wantErr: false,
		},
		{
			name:    "admin user token",
			userID:  2,
			email:   "admin@example.com",
			role:    "admin",
			wantErr: false,
		},
		{
			name:    "zero user ID",
			userID:  0,
			email:   "test@example.com",
			role:    "user",
			wantErr: false,
		},
		{
			name:    "large user ID",
			userID:  999999999,
			email:   "largeuser@example.com",
			role:    "user",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			accessToken, refreshToken, err := GenerateTokenPair(cfg, tt.userID, tt.email, tt.role)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, accessToken)
			assert.NotEmpty(t, refreshToken)
			assert.NotEqual(t, accessToken, refreshToken, "access and refresh tokens should be different")
		})
	}
}

func TestValidateToken(t *testing.T) {
	t.Parallel()

	cfg := newTestConfig()

	// Generate a valid token for testing
	userID := uint(1)
	email := "test@example.com"
	role := "user"

	accessToken, _, err := GenerateTokenPair(cfg, userID, email, role)
	require.NoError(t, err)

	tests := []struct {
		name      string
		token     string
		secret    string
		wantErr   bool
		checkFunc func(t *testing.T, claims *Claims)
	}{
		{
			name:    "valid token",
			token:   accessToken,
			secret:  cfg.JWT.Secret,
			wantErr: false,
			checkFunc: func(t *testing.T, claims *Claims) {
				assert.Equal(t, userID, claims.UserID)
				assert.Equal(t, email, claims.Email)
				assert.Equal(t, role, claims.Role)
			},
		},
		{
			name:    "invalid secret",
			token:   accessToken,
			secret:  "wrong-secret",
			wantErr: true,
		},
		{
			name:    "malformed token",
			token:   "invalid.token.format",
			secret:  cfg.JWT.Secret,
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			secret:  cfg.JWT.Secret,
			wantErr: true,
		},
		{
			name:    "completely invalid token",
			token:   "notavalidtoken",
			secret:  cfg.JWT.Secret,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			claims, err := ValidateToken(tt.token, tt.secret)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, claims)

			if tt.checkFunc != nil {
				tt.checkFunc(t, claims)
			}
		})
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	t.Parallel()

	// Create a config with very short expiration
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:                "test-secret",
			ExpiresIn:             -time.Hour, // Already expired
			RefreshTokenExpiresIn: -time.Hour,
		},
	}

	// Generate token that's already expired
	accessToken, _, err := GenerateTokenPair(cfg, 1, "test@example.com", "user")
	require.NoError(t, err)

	// Validation should fail for expired token
	_, err = ValidateToken(accessToken, cfg.JWT.Secret)
	assert.Error(t, err, "expired token should fail validation")
}

func TestGenerateTokenPair_TokensContainClaims(t *testing.T) {
	t.Parallel()

	cfg := newTestConfig()
	userID := uint(42)
	email := "claims@example.com"
	role := "admin"

	accessToken, refreshToken, err := GenerateTokenPair(cfg, userID, email, role)
	require.NoError(t, err)

	// Validate access token and check claims
	accessClaims, err := ValidateToken(accessToken, cfg.JWT.Secret)
	require.NoError(t, err)
	assert.Equal(t, userID, accessClaims.UserID)
	assert.Equal(t, email, accessClaims.Email)
	assert.Equal(t, role, accessClaims.Role)

	// Validate refresh token and check claims
	refreshClaims, err := ValidateToken(refreshToken, cfg.JWT.Secret)
	require.NoError(t, err)
	assert.Equal(t, userID, refreshClaims.UserID)
	assert.Equal(t, email, refreshClaims.Email)
	assert.Equal(t, role, refreshClaims.Role)
}

func TestGenerateTokenPair_DifferentExpirations(t *testing.T) {
	t.Parallel()

	cfg := newTestConfig()
	cfg.JWT.ExpiresIn = time.Hour
	cfg.JWT.RefreshTokenExpiresIn = 24 * time.Hour

	accessToken, refreshToken, err := GenerateTokenPair(cfg, 1, "test@example.com", "user")
	require.NoError(t, err)

	accessClaims, err := ValidateToken(accessToken, cfg.JWT.Secret)
	require.NoError(t, err)

	refreshClaims, err := ValidateToken(refreshToken, cfg.JWT.Secret)
	require.NoError(t, err)

	// Refresh token should have later expiration than access token
	assert.True(t, refreshClaims.ExpiresAt.After(accessClaims.ExpiresAt.Time),
		"refresh token should expire after access token")
}
