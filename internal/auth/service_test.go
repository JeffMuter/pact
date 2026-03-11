package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword_NotPlaintext(t *testing.T) {
	password := "TestPassword123!"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	// Hash should not equal plaintext
	assert.NotEqual(t, password, string(hashedPassword))
}

func TestCheckPasswordHash_CorrectPassword(t *testing.T) {
	password := "TestPassword123!"

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	// Verify the correct password matches the hash
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	assert.NoError(t, err)
}

func TestCheckPasswordHash_WrongPassword(t *testing.T) {
	password := "TestPassword123!"
	wrongPassword := "WrongPassword456!"

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	// Verify the wrong password fails
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(wrongPassword))
	assert.Error(t, err)
}

func TestCheckPasswordHash_EmptyPassword(t *testing.T) {
	password := "TestPassword123!"
	emptyPassword := ""

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	// Verify empty password fails
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(emptyPassword))
	assert.Error(t, err)
}
