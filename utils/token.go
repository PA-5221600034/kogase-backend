package utils

import (
	"os"
	"time"

	"github.com/atqamz/kogase-backend/middleware"
	"github.com/atqamz/kogase-backend/models"
	"github.com/golang-jwt/jwt/v5"
)

// CreateToken creates a JWT token for a user
func CreateToken(user models.User) (string, time.Time, error) {
	// Get JWT expiry from env
	expiryStr := os.Getenv("JWT_EXPIRES_IN")
	if expiryStr == "" {
		expiryStr = "24h" // Default to 24 hours
	}

	// Parse the duration
	expiryDuration, err := time.ParseDuration(expiryStr)
	if err != nil {
		expiryDuration = 24 * time.Hour // Default to 24 hours
	}

	expiresAt := time.Now().Add(expiryDuration)

	// Create claims
	claims := middleware.JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "kogase-api",
			Subject:   user.ID.String(),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}
