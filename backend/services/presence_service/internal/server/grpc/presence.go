package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	presencev1 "github.com/sudo-odner/minor-shared/pkg/pb/presence/v1"
	"github.com/sudo-odner/minor/backend/services/presence_service/internal/models"
	"go.uber.org/zap"
)

type PresenceService interface {
	SetStatus(ctx context.Context, userID string, status models.UserStatus, customStatus string) error
	GetStatus(ctx context.Context, userID uuid.UUID) (*models.Presence, error)
	GetUserStatuses(ctx context.Context, userIDs []string) (map[string]*models.Presence, error)
}

func (s *ServerAPI) GetStatus(
	ctx context.Context,
	req *presencev1.GetStatusRequest,
) (*presencev1.GetStatusResponse, error) {
	const op = "server.grpc.GetStatus"

	userIDStr := req.GetUserId()
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("userID in uncurrect format uuid")
	}
	userStatus, err := s.presenceService.GetStatus(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &presencev1.GetStatusResponse{
		Presence: &presencev1.Presence{
			UserId:       userID.String(),
			Status:       presencev1.UserStatus(userStatus.Status),
			CustomStatus: userStatus.CustomStatus,
			LastActiveAt: userStatus.LastActiveAt,
		},
	}, nil
}

func (s *ServerAPI) SetStatus(
	ctx context.Context,
	req *presencev1.SetStatusRequest,
) (*presencev1.SetStatusResponse, error) {
	const op = "server.grpc.SetStatus"

	userID := req.GetUserId()
	status := models.UserStatus(req.GetStatus())
	customStatus := req.GetCutsomStatus()

	err := s.presenceService.SetStatus(ctx, userID, status, customStatus)
	if err != nil {
		s.log.Error("failed to set status", zap.String("op", op), zap.Error(err))
		return &presencev1.SetStatusResponse{Success: false}, err
	}

	return &presencev1.SetStatusResponse{
		Success: true,
	}, nil
}

func (s *ServerAPI) GetUserStatuses(
	ctx context.Context,
	req *presencev1.GetUserStatusesRequest,
) (*presencev1.GetUserStatusesResponse, error) {
	const op = "server.grpc.GetUserStatuses"

	userIDs := req.GetUserIds()
	res, err := s.presenceService.GetUserStatuses(ctx, userIDs)
	if err != nil {
		s.log.Error("failed to get user statuses", zap.String("op", op), zap.Error(err))
		return nil, err
	}

	pbStatuses := make(map[string]*presencev1.Presence)
	for id, p := range res {
		pbStatuses[id] = &presencev1.Presence{
			UserId:       p.UserID,
			Status:       presencev1.UserStatus(p.Status),
			CustomStatus: p.CustomStatus,
			LastActiveAt: p.LastActiveAt,
		}
	}

	return &presencev1.GetUserStatusesResponse{
		Statuses: pbStatuses,
	}, nil
}
