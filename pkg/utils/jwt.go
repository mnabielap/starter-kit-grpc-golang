package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenPayload defines the contents of the JWT
type TokenPayload struct {
	UserID string `json:"sub"`  // Subject (User ID)
	Role   string `json:"role"` // RBAC Role
	Type   string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT token
func GenerateToken(userID string, role string, tokenType string, expires time.Duration, secret string) (string, time.Time, error) {
	expirationTime := time.Now().Add(expires)

	claims := &TokenPayload{
		UserID: userID,
		Role:   role,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return signedToken, expirationTime, nil
}

// ValidateToken parses and verifies a JWT token
func ValidateToken(tokenString string, secret string) (*TokenPayload, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenPayload{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenPayload); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}