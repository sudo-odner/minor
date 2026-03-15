package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/config"
)

type Claims struct {
	UserID uuid.UUID
	Email  string
	jwt.RegisteredClaims
}

func GenerateTokens(cfg config.TokenConfig, userID uuid.UUID, email string) (accessToken string, refreshToken string, err error) {
	accessClaims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(cfg.AccessSecret)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshClaims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.RefreshTokenTTL * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(cfg.RefreshSecret)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func ValidateAccessToken(cfg config.TokenConfig, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return cfg.AccessSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to check token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func Refresh(cfg config.TokenConfig, oldRefreshToken string) (newAccess string, newRefresh string, err error) {
	token, err := jwt.Parse(oldRefreshToken, func(t *jwt.Token) (any, error) {
		return cfg.RefreshSecret, nil
	})

	if err != nil || !token.Valid {
		return "", "", fmt.Errorf("invalid refresh token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return "", "", fmt.Errorf("invalid refresh token")
	}

	return GenerateTokens(cfg, claims.UserID, claims.Email)
}
