package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	communityv1 "github.com/sudo-odner/minor-shared/pkg/pb/community/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PermissionService interface {
	FetchPermissions(ctx context.Context, userID, channelID uuid.UUID) (authz.Permission, error)
}

type ServerAPI struct {
	communityv1.UnimplementedCommunityServiceServer
	log         *zap.Logger
	permService PermissionService
}

func New(log *zap.Logger, permService PermissionService) *ServerAPI {
	return &ServerAPI{
		log:         log,
		permService: permService,
	}
}

func (s *ServerAPI) FetchPermission(
	ctx context.Context,
	req *communityv1.FetchPermissionRequest,
) (*communityv1.FetchPermissionResponse, error) {
	const op = "server.grpc.FetchPermission"

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	channelID, err := uuid.Parse(req.GetChannelId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid channel_id format")
	}

	permissions, err := s.permService.FetchPermissions(ctx, userID, channelID)
	if err != nil {
		s.log.Error("failed to fetch permissions", zap.String("op", op), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to fetch permissions")
	}

	return &communityv1.FetchPermissionResponse{
		PermissionMask: uint64(permissions),
	}, nil
}
