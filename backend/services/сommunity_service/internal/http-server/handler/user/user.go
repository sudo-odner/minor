package user

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/model"
	"go.uber.org/zap"
)

type UserService interface {
	CreateUser(ctx context.Context, userID uuid.UUID, username, bio string) (*model.User, error) // Создание пользователя
	GetUser(ctx context.Context, userID uuid.UUID) (*model.User, error)                          // Получение профиля пользователя по ID
	UpdateUser(ctx context.Context, username, bio *string) (*model.User, error)                  // Изменнеие пользователя
	DeleteUser(ctx context.Context, userID uuid.UUID) error                                      // Удаление пользователя
}

type UserHandler struct {
	log         *zap.Logger
	userService UserService
}

func New(log *zap.Logger, userService UserService) *UserHandler {
	return &UserHandler{
		log:         log,
		userService: userService,
	}
}

func (h *UserHandler) CreateUser(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *UserHandler) GetUser(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *UserHandler) UpdateUser(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *UserHandler) DeleteUser(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
