package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/models"
	"go.uber.org/zap"
)

type UserService interface {
	CreateUser(ctx context.Context, userID uuid.UUID, username, bio string) (*models.User, error)
	GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, username, bio *string) (*models.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}

type UserHandler struct {
	log         *zap.Logger
	userService UserService
}

func NewUserHandler(log *zap.Logger, userService UserService) *UserHandler {
	return &UserHandler{
		log:         log,
		userService: userService,
	}
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Bio      string `json:"bio"`
}

type UserResponse struct {
	ID        string  `json:"id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	AvatarURL *string `json:"avatar_url"`
	Bio       string  `json:"bio"`
	CreateAt  string  `json:"created_at"`
	UpdateAt  string  `json:"updated_at"`
}

func toUserResponse(u *models.User) UserResponse {
	return UserResponse{
		ID:        u.ID.String(),
		Username:  u.Username,
		Email:     u.Email,
		AvatarURL: u.AvatarURL,
		Bio:       u.Bio,
		CreateAt:  u.CreateAt.Format(time.RFC3339),
		UpdateAt:  u.UpdateAt.Format(time.RFC3339),
	}
}

func (h *UserHandler) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.user.CreateUser"
		log := h.log.With(zap.String("op", op))

		actorID, err := ParseUserID(r)
		if err != nil {
			log.Debug("unauthorized request", zap.Error(err))
			RenderError(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
			return
		}

		var req CreateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Debug("failed to decode request body", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
			return
		}

		if req.Username == "" {
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "username is required")
			return
		}

		u, err := h.userService.CreateUser(r.Context(), actorID, req.Username, req.Bio)
		if err != nil {
			log.Error("failed to create user", zap.Error(err))
			RenderModelError(w, r, err, "failed to create user")
			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, toUserResponse(u))
	}
}

func (h *UserHandler) GetMe() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.user.GetMe"
		log := h.log.With(zap.String("op", op))

		actorID, err := ParseUserID(r)
		if err != nil {
			log.Debug("unauthorized request", zap.Error(err))
			RenderError(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
			return
		}

		u, err := h.userService.GetUser(r.Context(), actorID)
		if err != nil {
			log.Error("failed to get user", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		render.JSON(w, r, toUserResponse(u))
	}
}

func (h *UserHandler) GetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.user.GetUser"
		log := h.log.With(zap.String("op", op))

		userID, err := ParseUUIDParam(r, "user_id")
		if err != nil {
			log.Debug("invalid user_id parameter", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid user_id")
			return
		}

		u, err := h.userService.GetUser(r.Context(), userID)
		if err != nil {
			log.Error("failed to get user", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		render.JSON(w, r, toUserResponse(u))
	}
}

type UpdateUserRequest struct {
	Username *string `json:"username"`
	Bio      *string `json:"bio"`
}

func (h *UserHandler) UpdateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.user.UpdateUser"
		log := h.log.With(zap.String("op", op))

		actorID, err := ParseUserID(r)
		if err != nil {
			log.Debug("unauthorized request", zap.Error(err))
			RenderError(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
			return
		}

		var req UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Debug("failed to decode request body", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
			return
		}

		updated, err := h.userService.UpdateUser(r.Context(), actorID, req.Username, req.Bio)
		if err != nil {
			log.Error("failed to update user", zap.Error(err))
			RenderModelError(w, r, err, "failed to update user")
			return
		}

		render.JSON(w, r, toUserResponse(updated))
	}
}

func (h *UserHandler) DeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.user.DeleteUser"
		log := h.log.With(zap.String("op", op))

		actorID, err := ParseUserID(r)
		if err != nil {
			log.Debug("unauthorized request", zap.Error(err))
			RenderError(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
			return
		}

		if err := h.userService.DeleteUser(r.Context(), actorID); err != nil {
			log.Error("failed to delete user", zap.Error(err))
			RenderModelError(w, r, err, "failed to delete user")
			return
		}

		render.Status(r, http.StatusNoContent)
		render.JSON(w, r, nil)
	}
}
