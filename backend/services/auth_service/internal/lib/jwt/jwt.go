package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/models"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

func GenerateAccessToken(cfg config.AuthConfig, user *models.User) (accessToken string, err error) {
	accessClaims := models.Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to generate access token: %w", err)
	}

	return accessToken, nil
}

func ValidateAccessToken(cfg config.AuthConfig, tokenString string) (*models.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to check token: %w", err)
	}

	if claims, ok := token.Claims.(*models.Claims); ok && token.Valid {
		if !ok || !token.Valid {
			return nil, ErrInvalidToken
		}
		
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// func Refresh(cfg config.AuthConfig, oldRefreshToken string) (newAccess string, err error) {
// 	token, err := jwt.Parse(oldRefreshToken, func(t *jwt.Token) (any, error) {
// 		return cfg.JWTSecret, nil
// 	})

// 	if err != nil || !token.Valid {
// 		return "", fmt.Errorf("invalid refresh token")
// 	}

// 	claims, ok := token.Claims.(*models.Claims)
// 	if !ok || !token.Valid {
// 		return "", fmt.Errorf("invalid refresh token")
// 	}

// 	return GenerateAccessToken(cfg, &models.User{
// 		ID: claims.UserID,
// 		Email: claims.Email,
// 		Username: claims.Username,
// 	})
// }
