package handler

import (
	"context"
	"log/slog"

	dmv1 "github.com/sudo-odner/minor-shared/pkg/pb/dm/v1"
)

type DMChannelService interface {
	Permission()
	Members()
}

type Handler struct {
	dmv1.UnimplementedDMServiceServer
	log              *slog.Logger
	dmChannelService DMChannelService
}

func New(log *slog.Logger, dmChannelSerivce DMChannelService) *Handler {
	return &Handler{
		log:              log,
		dmChannelService: dmChannelSerivce,
	}
}

func (h *Handler) FetchPermission(
	ctx context.Context,
	req *dmv1.FetchPermissionRequest,
) (*dmv1.FetchPermissionResponse, error) {
	return nil, nil
}

func (h *Handler) FetchMembers(
	ctx context.Context,
	req *dmv1.FetchMembersRequest,
) (*dmv1.FetchMembersResponse, error) {
	return nil, nil
}
