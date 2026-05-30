package auth

import (
	"context"
	"errors"

	"github.com/sudo-odner/minor/backend/services/auth_service/internal/lib/jwt"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/models"
)

func (s *AuthService) VerifyAccessToken(ctx context.Context, tokenStr string) (*models.Claims, error) {
	claims, err := jwt.ValidateAccessToken(s.authConfig, tokenStr)
	if err != nil {
		if errors.Is(err, jwt.ErrExpiredToken) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
    
	return claims, nil
}