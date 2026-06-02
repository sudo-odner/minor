package auth

import (
	"context"

	authv1 "github.com/sudo-odner/minor-shared/pkg/pb/auth/v1"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/models"
	"go.uber.org/zap"
)

type AuthGRPCService interface {
	VerifyToken(ctx context.Context, token string) (*models.Claims, error)
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

func (gh *AuthGRPCHandler) VerifyToken(ctx context.Context, req *authv1.VerifyTokenRequest) (*authv1.VerifyTokenResponse, error) {
	const path = "grpc.handler.VerifyAccessToken"

	log := gh.log.With(
		zap.String("path", path),
		zap.String("access-token", req.AccessToken),
	)

	log.Info("starting verify token")
	
	claims, err := gh.authService.VerifyToken(ctx, req.AccessToken)

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