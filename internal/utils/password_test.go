package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "securePassword123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "thisIsAVeryLongPasswordThatShouldStillWork12345!@#$%",
			wantErr:  false,
		},
		{
			name:     "password with special characters",
			password: "p@$$w0rd!#%^&*()",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			hash, err := HashPassword(tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tt.password, hash, "hash should not equal plain password")
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	t.Parallel()

	// Create a known hash for testing
	originalPassword := "testPassword123"
	hashedPassword, err := HashPassword(originalPassword)
	require.NoError(t, err)

	tests := []struct {
		name           string
		hashedPassword string
		password       string
		wantErr        bool
	}{
		{
			name:           "correct password",
			hashedPassword: hashedPassword,
			password:       originalPassword,
			wantErr:        false,
		},
		{
			name:           "incorrect password",
			hashedPassword: hashedPassword,
			password:       "wrongPassword",
			wantErr:        true,
		},
		{
			name:           "empty password against valid hash",
			hashedPassword: hashedPassword,
			password:       "",
			wantErr:        true,
		},
		{
			name:           "invalid hash format",
			hashedPassword: "invalidhash",
			password:       originalPassword,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := VerifyPassword(tt.hashedPassword, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestHashPasswordAndVerify_Integration(t *testing.T) {
	t.Parallel()

	passwords := []string{
		"simple",
		"P@ssw0rd!",
		"verylongpasswordwithmanycharacters123456789",
		"unicode密码测试",
		"  spaces  ",
	}

	for _, password := range passwords {
		t.Run(password, func(t *testing.T) {
			t.Parallel()

			// Hash the password
			hash, err := HashPassword(password)
			require.NoError(t, err)

			// Verify correct password succeeds
			err = VerifyPassword(hash, password)
			assert.NoError(t, err, "correct password should verify successfully")

			// Verify wrong password fails
			err = VerifyPassword(hash, "wrongpassword")
			assert.Error(t, err, "wrong password should fail verification")
		})
	}
}

func TestHashPassword_UniqueSalts(t *testing.T) {
	t.Parallel()

	password := "samePassword123"

	// Hash the same password twice
	hash1, err := HashPassword(password)
	require.NoError(t, err)

	hash2, err := HashPassword(password)
	require.NoError(t, err)

	// Hashes should be different due to unique salts
	assert.NotEqual(t, hash1, hash2, "same password should produce different hashes")

	// But both should verify correctly
	assert.NoError(t, VerifyPassword(hash1, password))
	assert.NoError(t, VerifyPassword(hash2, password))
}
