package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTManager(t *testing.T) {
	secret := "test-secret"
	manager := NewJWTManager(secret)

	assert.NotNil(t, manager)
	assert.Equal(t, secret, manager.secret)
}

func TestJWTManager_Generate_Success(t *testing.T) {
	manager := NewJWTManager("test-secret")

	token, err := manager.Generate("user123", "tenant456", "test@example.com", "admin")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTManager_Generate_ValidTokenStructure(t *testing.T) {
	manager := NewJWTManager("test-secret")

	token, err := manager.Generate("user123", "tenant456", "test@example.com", "editor")
	require.NoError(t, err)

	// Parse token to verify structure
	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})

	require.NoError(t, err)
	require.True(t, parsedToken.Valid)

	claims := parsedToken.Claims.(*Claims)
	assert.Equal(t, "user123", claims.UserID)
	assert.Equal(t, "tenant456", claims.TenantID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "editor", claims.Role)
}

func TestJWTManager_Generate_TokenExpiration(t *testing.T) {
	manager := NewJWTManager("test-secret")

	token, err := manager.Generate("user123", "tenant456", "test@example.com", "admin")
	require.NoError(t, err)

	parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})

	require.NoError(t, err)
	claims := parsedToken.Claims.(*Claims)

	// Check expiration is approximately 24 hours from now
	expectedExpiry := time.Now().Add(24 * time.Hour)
	timeDiff := expectedExpiry.Sub(claims.ExpiresAt.Time)

	assert.Less(t, timeDiff.Abs(), time.Minute, "Expiration should be approximately 24 hours")
}

func TestJWTManager_Validate_Success(t *testing.T) {
	manager := NewJWTManager("test-secret")

	token, err := manager.Generate("user123", "tenant456", "test@example.com", "admin")
	require.NoError(t, err)

	claims, err := manager.Validate(token)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "user123", claims.UserID)
	assert.Equal(t, "tenant456", claims.TenantID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "admin", claims.Role)
}

func TestJWTManager_Validate_InvalidToken(t *testing.T) {
	manager := NewJWTManager("test-secret")

	invalidTokens := []string{
		"invalid.token.string",
		"",
		"Bearer invalid",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid",
	}

	for _, token := range invalidTokens {
		claims, err := manager.Validate(token)
		assert.Error(t, err, "Should return error for invalid token: %s", token)
		assert.Nil(t, claims)
	}
}

func TestJWTManager_Validate_WrongSecret(t *testing.T) {
	manager1 := NewJWTManager("secret1")
	manager2 := NewJWTManager("secret2")

	token, err := manager1.Generate("user123", "tenant456", "test@example.com", "admin")
	require.NoError(t, err)

	claims, err := manager2.Validate(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTManager_Validate_MalformedToken(t *testing.T) {
	manager := NewJWTManager("test-secret")

	malformedTokens := []string{
		"not.a.token",
		"onlytwoparts",
		"header.payload", // Missing signature
	}

	for _, token := range malformedTokens {
		claims, err := manager.Validate(token)
		assert.Error(t, err, "Should return error for malformed token: %s", token)
		assert.Nil(t, claims)
	}
}

func TestJWTManager_GenerateAndValidate_RoundTrip(t *testing.T) {
	manager := NewJWTManager("test-secret")

	testCases := []struct {
		userID   string
		tenantID string
		email    string
		role     string
	}{
		{"user1", "tenant1", "user1@example.com", "admin"},
		{"user2", "tenant2", "user2@example.com", "editor"},
		{"user3", "tenant3", "user3@example.com", "viewer"},
	}

	for _, tc := range testCases {
		t.Run(tc.email, func(t *testing.T) {
			token, err := manager.Generate(tc.userID, tc.tenantID, tc.email, tc.role)
			require.NoError(t, err)

			claims, err := manager.Validate(token)
			require.NoError(t, err)

			assert.Equal(t, tc.userID, claims.UserID)
			assert.Equal(t, tc.tenantID, claims.TenantID)
			assert.Equal(t, tc.email, claims.Email)
			assert.Equal(t, tc.role, claims.Role)
		})
	}
}

func TestJWTManager_MultipleManagers(t *testing.T) {
	// Test that different managers can have different secrets
	manager1 := NewJWTManager("secret1")
	manager2 := NewJWTManager("secret2")

	token1, err := manager1.Generate("user1", "tenant1", "user1@example.com", "admin")
	require.NoError(t, err)

	token2, err := manager2.Generate("user2", "tenant2", "user2@example.com", "admin")
	require.NoError(t, err)

	// Each manager should validate its own tokens
	claims1, err := manager1.Validate(token1)
	assert.NoError(t, err)
	assert.Equal(t, "user1", claims1.UserID)

	claims2, err := manager2.Validate(token2)
	assert.NoError(t, err)
	assert.Equal(t, "user2", claims2.UserID)

	// But not the other's tokens
	_, err = manager1.Validate(token2)
	assert.Error(t, err)

	_, err = manager2.Validate(token1)
	assert.Error(t, err)
}

func TestJWTManager_Generate_AllRoles(t *testing.T) {
	manager := NewJWTManager("test-secret")

	roles := []string{"admin", "editor", "viewer"}

	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			token, err := manager.Generate("user123", "tenant456", "test@example.com", role)
			require.NoError(t, err)

			claims, err := manager.Validate(token)
			require.NoError(t, err)
			assert.Equal(t, role, claims.Role)
		})
	}
}

func TestJWTManager_Validate_EmptyToken(t *testing.T) {
	manager := NewJWTManager("test-secret")

	claims, err := manager.Validate("")

	assert.Error(t, err)
	assert.Nil(t, claims)
}
