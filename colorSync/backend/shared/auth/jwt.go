package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Secret keys - In production, use environment variables
const (
	// Used for user authentication tokens
	UserSecretKey = "user-jwt-secret-key-change-in-production"

	// Used for service-to-service authentication (Zero Trust)
	ServiceSecretKey = "service-to-service-secret-key-change-in-production"
)

// Claims structure for user JWT tokens
type UserClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Claims structure for service JWT tokens (Zero Trust)
type ServiceClaims struct {
	ServiceName string `json:"service_name"`
	jwt.RegisteredClaims
}

// GenerateUserToken creates a JWT token for authenticated users
// Token expires in 24 hours
func GenerateUserToken(userID, username string) (string, error) {
	// Create claims with user info and expiration
	claims := UserClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	tokenString, err := token.SignedString([]byte(UserSecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// VerifyUserToken validates a JWT token and returns the claims
func VerifyUserToken(tokenString string) (*UserClaims, error) {
	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(UserSecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Extract claims
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GenerateServiceToken creates a JWT token for service-to-service auth (Zero Trust)
// Token expires in 1 hour
func GenerateServiceToken(serviceName string) (string, error) {
	// Create claims with service name and expiration
	claims := ServiceClaims{
		ServiceName: serviceName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with service secret key
	tokenString, err := token.SignedString([]byte(ServiceSecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign service token: %w", err)
	}

	return tokenString, nil
}

// VerifyServiceToken validates a service JWT token
func VerifyServiceToken(tokenString string) (*ServiceClaims, error) {
	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &ServiceClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(ServiceSecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse service token: %w", err)
	}

	// Extract claims
	if claims, ok := token.Claims.(*ServiceClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid service token")
}
