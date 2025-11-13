package jwt

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lzimin05/course-todo/config"
)

func TestNewTokenator(t *testing.T) {
	jwtConfig := &config.JWTConfig{
		Signature:     "test-secret",
		TokenLifeSpan: time.Hour,
	}

	tokenator := NewTokenator(jwtConfig)

	assert.NotNil(t, tokenator)
	assert.Equal(t, "test-secret", tokenator.sign)
	assert.Equal(t, time.Hour, tokenator.tokenLifeSpan)
}

func TestTokenator_CreateJWT(t *testing.T) {
	jwtConfig := &config.JWTConfig{
		Signature:     "test-secret-key",
		TokenLifeSpan: time.Hour,
	}

	tokenator := NewTokenator(jwtConfig)
	userID := "user123"

	token, err := tokenator.CreateJWT(userID)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Contains(t, token, ".") // JWT should contain dots
}

func TestTokenator_ParseJWT_ValidToken(t *testing.T) {
	jwtConfig := &config.JWTConfig{
		Signature:     "test-secret-key",
		TokenLifeSpan: time.Hour,
	}

	tokenator := NewTokenator(jwtConfig)
	userID := "user123"

	// Create a valid token
	token, err := tokenator.CreateJWT(userID)
	assert.NoError(t, err)

	// Parse the token
	claims, err := tokenator.ParseJWT(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, userID, claims.UserID)
}

func TestTokenator_ParseJWT_InvalidToken(t *testing.T) {
	jwtConfig := &config.JWTConfig{
		Signature:     "test-secret-key",
		TokenLifeSpan: time.Hour,
	}

	tokenator := NewTokenator(jwtConfig)

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "completely invalid token",
			token:       "invalid-token",
			expectError: true,
		},
		{
			name:        "empty token",
			token:       "",
			expectError: true,
		},
		{
			name:        "malformed JWT",
			token:       "header.payload", // Missing signature
			expectError: true,
		},
		{
			name:        "wrong signature",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdCJ9.wrong_signature",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := tokenator.ParseJWT(tt.token)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
			}
		})
	}
}

func TestTokenator_ParseJWT_ExpiredToken(t *testing.T) {
	// Create tokenator with very short lifespan
	jwtConfig := &config.JWTConfig{
		Signature:     "test-secret-key",
		TokenLifeSpan: -time.Hour, // Already expired
	}

	tokenator := NewTokenator(jwtConfig)
	userID := "user123"

	// Create an expired token
	token, err := tokenator.CreateJWT(userID)
	assert.NoError(t, err)

	// Try to parse the expired token
	claims, err := tokenator.ParseJWT(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestTokenator_CreateAndParseRoundTrip(t *testing.T) {
	jwtConfig := &config.JWTConfig{
		Signature:     "test-secret-key",
		TokenLifeSpan: time.Hour,
	}

	tokenator := NewTokenator(jwtConfig)

	testCases := []string{
		"user123",
		"admin",
		"user-with-dashes",
		"user_with_underscores",
		"123456",
	}

	for _, userID := range testCases {
		t.Run("user_"+userID, func(t *testing.T) {
			// Create token
			token, err := tokenator.CreateJWT(userID)
			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			// Parse token
			claims, err := tokenator.ParseJWT(token)
			assert.NoError(t, err)
			assert.NotNil(t, claims)
			assert.Equal(t, userID, claims.UserID)

			// Verify claims structure
			assert.NotNil(t, claims.IssuedAt)
			assert.NotNil(t, claims.ExpiresAt)
		})
	}
}

func TestTokenator_DifferentSecrets(t *testing.T) {
	// Create two tokenators with different secrets
	tokenator1 := NewTokenator(&config.JWTConfig{
		Signature:     "secret1",
		TokenLifeSpan: time.Hour,
	})

	tokenator2 := NewTokenator(&config.JWTConfig{
		Signature:     "secret2",
		TokenLifeSpan: time.Hour,
	})

	userID := "user123"

	// Create token with first tokenator
	token, err := tokenator1.CreateJWT(userID)
	assert.NoError(t, err)

	// Try to parse with second tokenator (different secret)
	claims, err := tokenator2.ParseJWT(token)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTClaims_Structure(t *testing.T) {
	// Test that JWTClaims has the expected structure
	claims := JWTClaims{
		UserID: "test-user",
	}

	assert.Equal(t, "test-user", claims.UserID)
	assert.NotNil(t, claims.RegisteredClaims)
}
