package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken_ValidUserID(t *testing.T) {
	token, err := GenerateToken(123)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, len(token) > 0)
}

func TestGenerateToken_ContainsUserID(t *testing.T) {
	userId := uint(456)
	tokenString, err := GenerateToken(userId)
	require.NoError(t, err)

	// Parse token without validation to extract claims
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	require.NoError(t, err)

	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok, "claims should be MapClaims")

	extractedUserID, ok := claims["userId"].(float64)
	require.True(t, ok, "userId should be present in claims")
	assert.Equal(t, float64(userId), extractedUserID)
}

func TestValidateToken_ValidToken(t *testing.T) {
	// Generate a fresh token
	userId := uint(789)
	tokenString, err := GenerateToken(userId)
	require.NoError(t, err)

	// Validate it
	extractedUserID, err := ValidateToken(tokenString)
	require.NoError(t, err)
	assert.Equal(t, int(userId), extractedUserID)
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	// Create an expired token manually
	expiredClaims := jwt.MapClaims{
		"userId":     123,
		"expiration": time.Now().Add(-1 * time.Hour).Unix(), // 1 hour in the past
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, err := token.SignedString(jwtKey)
	require.NoError(t, err)

	// Validation should fail
	_, err = ValidateToken(expiredTokenString)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestValidateToken_MalformedToken(t *testing.T) {
	malformedToken := "not.a.valid.jwt.string"

	_, err := ValidateToken(malformedToken)
	require.Error(t, err)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	// Create a token signed with a different key
	wrongKey := []byte("this-is-not-the-real-secret-key")
	claims := jwt.MapClaims{
		"userId":     123,
		"expiration": time.Now().Add(time.Hour * 6).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	wrongSecretToken, err := token.SignedString(wrongKey)
	require.NoError(t, err)

	// Validation should fail because signature won't match
	_, err = ValidateToken(wrongSecretToken)
	require.Error(t, err)
}
