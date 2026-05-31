package grpc

import (
	presencev1 "github.com/sudo-odner/minor-shared/pkg/pb/presence/v1"
	"go.uber.org/zap"
)

type PresenceService interface{}

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
