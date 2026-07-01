package auth

import (
	"context"
	"fmt"
	"time"

	authv1 "github.com/sudo-odner/minor-shared/pkg/pb/auth/v1"
	"github.com/sudo-odner/minor/backend/services/gateway_service/internal/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthGRPCClient struct {
	api  authv1.AuthServiceClient
	conn *grpc.ClientConn
}

func New(log *zap.Logger, cfg *config.Config) (*AuthGRPCClient, error) {
	const op = "client.grpc.auth.New"

	log = log.With(
		zap.String("op", op),
	)

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	
	authConn, err := grpc.NewClient(cfg.GRPC.AuthService.Address, opts...)
	if err != nil {
		log.Fatal("failed to connect auth service", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &AuthGRPCClient{
		api:  authv1.NewAuthServiceClient(authConn),
		conn: authConn,
	}, nil
}

func (ac *AuthGRPCClient) VerifyToken(ctx context.Context, token string) (*authv1.VerifyTokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 2 * time.Second)
	defer cancel()

	return ac.api.VerifyToken(ctx, &authv1.VerifyTokenRequest{
		AccessToken: token,
	})
}
