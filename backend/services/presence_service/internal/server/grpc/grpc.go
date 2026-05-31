package grpc

import (
	"context"

	presencev1 "github.com/sudo-odner/minor-shared/pkg/pb/presence/v1"
	"github.com/sudo-odner/minor/backend/services/presence_service/internal/models"
	"go.uber.org/zap"
)

type PresenceService interface {
	SetStatus(ctx context.Context, userID string, status models.UserStatus, customStatus string) error
	GetUserStatuses(ctx context.Context, userIDs []string) (map[string]*models.Presence, error)
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
