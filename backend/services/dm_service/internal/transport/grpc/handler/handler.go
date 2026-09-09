package handler

import (
	"context"
	"log/slog"

	dmv1 "github.com/sudo-odner/minor-shared/pkg/pb/dm/v1"
)

type ChannelService interface {
	Permission()
	Members()
}

type GRPCHandler struct {
	dmv1.UnimplementedDMServiceServer
	log            *slog.Logger
	channelService ChannelService
}

func New(log *slog.Logger) *GRPCHandler {
	return &GRPCHandler{
		log: log,
	}
}

func (h *GRPCHandler) FetchPermission(
	ctx context.Context,
	req *dmv1.FetchPermissionRequest,
) (*dmv1.FetchPermissionResponse, error) {
	return nil, nil
}

func (h *GRPCHandler) FetchMembers(
	ctx context.Context,
	req *dmv1.FetchMembersRequest,
) (*dmv1.FetchMembersResponse, error) {
	return nil, nil
}
