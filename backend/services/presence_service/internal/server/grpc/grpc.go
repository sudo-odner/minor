package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	presencev1 "github.com/sudo-odner/minor-shared/pkg/pb/presence/v1"
	"go.uber.org/zap"
)

type PresenceService interface {
	FetchPermissions(ctx context.Context, userID, channelID uuid.UUID) (authz.Permission, error)
}

type ServerAPI struct {
	presencev1.UnimplementedPresenceServiceServer
	log             *zap.Logger
	presenceService PresenceService
}

func New(log *zap.Logger, presenceService PresenceService) *ServerAPI {
	return &ServerAPI{
		log:             log,
		presenceService: presenceService,
	}
}
