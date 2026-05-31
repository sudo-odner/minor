package grpc

import (
	"context"

	presencev1 "github.com/sudo-odner/minor-shared/pkg/pb/presence/v1"
	"github.com/sudo-odner/minor/backend/services/presence_service/internal/models"
	"go.uber.org/zap"
)

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
