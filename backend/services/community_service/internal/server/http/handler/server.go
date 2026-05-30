package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
	"go.uber.org/zap"
)

type ServerService interface {
	CreateServer(ctx context.Context, name string, ownerID uuid.UUID, avatarURL string) (*models.Server, error)
	GetServer(ctx context.Context, serverID uuid.UUID) (*models.Server, error)
	UpdateServer(ctx context.Context, actorID uuid.UUID, serverID uuid.UUID, name *string, avatarURL *string) (*models.Server, error)
	DeleteServer(ctx context.Context, actorID uuid.UUID, serverID uuid.UUID) error
	GetUserServers(ctx context.Context, userID uuid.UUID) ([]models.Server, error)
}

type ServerHandler struct {
	log           *zap.Logger
	serverService ServerService
}

func NewServerHandler(log *zap.Logger, serverService ServerService) *ServerHandler {
	return &ServerHandler{
		log:           log,
		serverService: serverService,
	}
}

type CreateServerRequest struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

type ServerResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	OwnerID   string `json:"owner_id"`
	AvatarURL string `json:"avatar_url"`
	CreatedAt string `json:"created_at"`
}

func (h *ServerHandler) CreateServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.server.CreateServer"
		log := h.log.With(zap.String("op", op))

		actorID, err := ParseUserID(r)
		if err != nil {
			log.Debug("unauthorized request", zap.Error(err))
			RenderError(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
			return
		}

		var req CreateServerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Debug("failed to decode request body", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
			return
		}

		if req.Name == "" {
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "server name is required")
			return
		}

		server, err := h.serverService.CreateServer(r.Context(), req.Name, actorID, req.AvatarURL)
		if err != nil {
			log.Error("failed to create server", zap.Error(err))
			RenderModelError(w, r, err, "failed to create server")
			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, ServerResponse{
			ID:        server.ID.String(),
			Name:      server.Name,
			OwnerID:   server.OwnerID.String(),
			AvatarURL: server.AvatarURL,
			CreatedAt: server.CreatedAt.Format(time.RFC3339),
		})
	}
}

func (h *ServerHandler) GetServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.server.GetServer"
		log := h.log.With(zap.String("op", op))

		serverID, err := ParseUUIDParam(r, "server_id")
		if err != nil {
			log.Debug("invalid server_id parameter", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid server_id")
			return
		}

		server, err := h.serverService.GetServer(r.Context(), serverID)
		if err != nil {
			log.Error("failed to get server", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		render.JSON(w, r, ServerResponse{
			ID:        server.ID.String(),
			Name:      server.Name,
			OwnerID:   server.OwnerID.String(),
			AvatarURL: server.AvatarURL,
			CreatedAt: server.CreatedAt.Format(time.RFC3339),
		})
	}
}

func (h *ServerHandler) GetUserServers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.server.GetUserServers"
		log := h.log.With(zap.String("op", op))

		userID, err := ParseUserID(r)
		if err != nil {
			log.Debug("unauthorized request", zap.Error(err))
			RenderError(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
			return
		}

		serversList, err := h.serverService.GetUserServers(r.Context(), userID)
		if err != nil {
			log.Error("failed to get user servers", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		res := make([]ServerResponse, len(serversList))
		for i := range serversList {
			res[i] = ServerResponse{
				ID:        serversList[i].ID.String(),
				Name:      serversList[i].Name,
				OwnerID:   serversList[i].OwnerID.String(),
				AvatarURL: serversList[i].AvatarURL,
				CreatedAt: serversList[i].CreatedAt.Format(time.RFC3339),
			}
		}
		render.JSON(w, r, res)
	}
}

type UpdateServerRequest struct {
	Name      *string `json:"name"`
	AvatarURL *string `json:"avatar_url"`
}

func (h *ServerHandler) UpdateServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.server.UpdateServer"
		log := h.log.With(zap.String("op", op))

		actorID, err := ParseUserID(r)
		if err != nil {
			log.Debug("unauthorized request", zap.Error(err))
			RenderError(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
			return
		}

		serverID, err := ParseUUIDParam(r, "server_id")
		if err != nil {
			log.Debug("invalid server_id parameter", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid server_id")
			return
		}

		var req UpdateServerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Debug("failed to decode request body", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
			return
		}

		updatedServer, err := h.serverService.UpdateServer(r.Context(), actorID, serverID, req.Name, req.AvatarURL)
		if err != nil {
			log.Error("failed to update server", zap.Error(err))
			RenderModelError(w, r, err, "failed to update server")
			return
		}

		render.JSON(w, r, ServerResponse{
			ID:        updatedServer.ID.String(),
			Name:      updatedServer.Name,
			OwnerID:   updatedServer.OwnerID.String(),
			AvatarURL: updatedServer.AvatarURL,
			CreatedAt: updatedServer.CreatedAt.Format(time.RFC3339),
		})
	}
}

func (h *ServerHandler) DeleteServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.server.DeleteServer"
		log := h.log.With(zap.String("op", op))

		actorID, err := ParseUserID(r)
		if err != nil {
			log.Debug("unauthorized request", zap.Error(err))
			RenderError(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
			return
		}

		serverID, err := ParseUUIDParam(r, "server_id")
		if err != nil {
			log.Debug("invalid server_id parameter", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid server_id")
			return
		}

		if err := h.serverService.DeleteServer(r.Context(), actorID, serverID); err != nil {
			log.Error("failed to delete server", zap.Error(err))
			RenderModelError(w, r, err, "failed to delete server")
			return
		}

		render.Status(r, http.StatusNoContent)
		render.JSON(w, r, nil)
	}
}
