package auth

import (
	"context"

	authv1 "github.com/sudo-odner/minor-shared/pkg/pb/auth/v1"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/models"
	"go.uber.org/zap"
)

type AuthGRPCService interface {
	VerifyAccessToken(ctx context.Context, token string) (*models.Claims, error)
}

type AuthGRPCHandler struct {
	authv1.UnimplementedAuthServiceServer
	authService AuthGRPCService
	log         *zap.Logger
}

func NewGRPCHandler(authService AuthGRPCService, log *zap.Logger) *AuthGRPCHandler {
	return &AuthGRPCHandler{
		authService: authService,
		log:         log,
	}
}

func (gh *AuthGRPCHandler) VerifyAccessToken(ctx context.Context, req *authv1.VerifyTokenRequest) (*authv1.VerifyTokenResponse, error) {
	claims, err := gh.authService.VerifyAccessToken(ctx, req.AccessToken)

	if err != nil {
		return &authv1.VerifyTokenResponse{
			IsValid:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &authv1.VerifyTokenResponse{
		UserId:    claims.UserID,
		IsValid:   true,
		ExpiresAt: claims.ExpiresAt.Unix(),
	}, nil
}