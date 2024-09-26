package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("your_secret_key") // Replace with a secure secret key

type Claims struct {
	UserID           uint `json:"user_id"`
	RegisteredClaims jwt.RegisteredClaims
}

func (c *Claims) Valid() error {
	// Check if the token has expired
	if c.RegisteredClaims.ExpiresAt != nil && time.Now().After(c.RegisteredClaims.ExpiresAt.Time) {
		return errors.New("token has expired")
	}

	// Check if the token is being used before its "not before" time
	if c.RegisteredClaims.NotBefore != nil && time.Now().Before(c.RegisteredClaims.NotBefore.Time) {
		return errors.New("token is not valid yet")
	}

	// Check if the issued at time is valid (not issued in the future)
	if c.RegisteredClaims.IssuedAt != nil && time.Now().Before(c.RegisteredClaims.IssuedAt.Time) {
		return errors.New("token issued in the future")
	}

	// Custom validation: Check if the username is not empty
	if c.UserID == 0 {
		return errors.New("invalid claim: username is empty")
	}

	return nil
}

func GenerateToken(userID uint) (string, error) {
	expirationTime := time.Now().Add(6 * time.Hour)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
