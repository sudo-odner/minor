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
	GetBatchProfiles(ctx context.Context, userIDs []string) (map[string]*userv1.UserProfile, error)
	GetUserName(ctx context.Context, userID string) (string, error)
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

func (h *ServerAPI) GetBatchProfiles(ctx context.Context, req *userv1.GetBatchProfilesRequest) (*userv1.GetBatchProfilesResponse, error) {
	h.log.Info("gRPC: GetBatchProfiles called", zap.Int("count", len(req.UserIds)))

	// 1. Вызываем сервис пользователей, чтобы достать данные из Postgres
	// profilesMap должен быть map[string]*userv1.UserProfile
	profilesMap, err := h.userService.GetBatchProfiles(ctx, req.UserIds)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get profiles: %v", err)
	}

	return &userv1.GetBatchProfilesResponse{
		Profiles: profilesMap,
	}, nil
}

func (s *ServerAPI) GetUserEmail(ctx context.Context, req *userv1.GetUserEmailRequest) (*userv1.GetUserEmailResponse, error) {
	const op = "server.grpc.GetUserEmail"

	userIDStr := req.GetUserId()
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	email, err := s.userService.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "falied get user email")
	}

	return &userv1.GetUserEmailResponse{
		Email: email.Email,
	}, nil
}

// GetUserName реализует метод из user.proto
func (s *ServerAPI) GetUserName(ctx context.Context, req *userv1.GetUserNameRequest) (*userv1.GetUserNameResponse, error) {
	const op = "server.grpc.GetUserName"

	log := s.log.With(
		zap.String("op", op),
	)

	if req.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	username, err := s.userService.GetUserName(ctx, req.GetUserId())
	if err != nil {
		// Логируем ошибку и возвращаем gRPC статус NotFound или Internal
		return nil, status.Errorf(codes.NotFound, "failed to get username: %v", err)
	}

	log.Info("got username", zap.String("username", username))

	return &userv1.GetUserNameResponse{
		Username: username,
	}, nil
}
