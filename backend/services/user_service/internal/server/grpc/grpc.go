package grpc

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	userv1 "github.com/sudo-odner/minor-shared/pkg/pb/user/v1"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/models"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService interface {
	GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error)
}

type ServerAPI struct {
	userv1.UnimplementedUserServiceServer
	log         *zap.Logger
	userService UserService
}

func New(log *zap.Logger, userService UserService) *ServerAPI {
	return &ServerAPI{
		log:         log,
		userService: userService,
	}
}

func (s *ServerAPI) GetUserProfile(
	ctx context.Context,
	req *userv1.GetUserProfileRequest,
) (*userv1.GetUserProfileResponse, error) {
	const op = "server.grpc.GetUserProfile"

	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	u, err := s.userService.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "user profile not found")
		}
		s.log.Error("failed to get user profile", zap.String("op", op), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get user profile")
	}

	avatarURL := ""
	if u.AvatarURL != nil {
		avatarURL = *u.AvatarURL
	}

	return &userv1.GetUserProfileResponse{
		UserId:    u.ID.String(),
		Email:     u.Email,
		Username:  u.Username,
		AvatarUrl: avatarURL,
		Bio:       u.Bio,
		CreatedAt: u.CreateAt.Format(time.RFC3339),
		UpdatedAt: u.UpdateAt.Format(time.RFC3339),
	}, nil
}
