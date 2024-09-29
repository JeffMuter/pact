package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

// this JWT package is solely responsible for operations involving the creation & validation of JWTs.
// keep in mind, we need a secret key which is stored elsewhere.

var jwtKey []byte

func init() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("godot couldn't open env")
	}
	// Load the JWT key from an environment variable
	jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))
	if len(jwtKey) == 0 {
		panic("JWT_SECRET_KEY environment variable not set")
	}
}

// GenerateToken() takes in a userId, and generates a tokenString.
func GenerateToken(userId uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":     userId,
		"expiration": time.Now().Add(time.Hour * 6).Unix(),
	})

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", fmt.Errorf("error signing string: %w", err)
	}
	return tokenString, nil
}

// ValidateToken() takes a tokenString,and validates that is hasn't expired. If expired, error.
func ValidateToken(tokenString string) (uint, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})

	if err != nil {
		return 0, fmt.Errorf("error parsing token: %w", err)
	}

	// Check if the token is valid
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check if the token has expired
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			return 0, errors.New("token has expired")
		}

		// Extract the user ID
		userID, ok := claims["user_id"].(float64)
		if !ok {
			return 0, errors.New("invalid user_id in token")
		}

		return uint(userID), nil
	}
	return 0, errors.New("invalid token")
}
