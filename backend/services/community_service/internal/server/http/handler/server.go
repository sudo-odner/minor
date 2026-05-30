package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
	"go.uber.org/zap"
	"context"
)

type ServerService interface {
	CreateServer(ctx context.Context, name string, ownerID uuid.UUID, avatarURL string) (*models.Server, error)
	GetServer(ctx context.Context, serverID uuid.UUID) (*models.Server, error)
	UpdateServer(ctx context.Context, serverID uuid.UUID, name *string, avatarURL *string) (*models.Server, error)
	DeleteServer(ctx context.Context, serverID uuid.UUID) error
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
			RenderError(w, r, http.StatusInternalServerError, CodeInternalServerError, "failed to create server")
			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, server)
	}
}

func (h *ServerHandler) GetServer() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.server.GetServer"
		log := h.log.With(zap.String("op", op))

		serverID, err := ParseUUIDParam(r, "serverID")
		if err != nil {
			log.Debug("invalid serverID parameter", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid serverID")
			return
		}

		server, err := h.serverService.GetServer(r.Context(), serverID)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				RenderError(w, r, http.StatusNotFound, CodeNotFound, "server not found")
				return
			}
			log.Error("failed to get server", zap.Error(err))
			RenderError(w, r, http.StatusInternalServerError, CodeInternalServerError, "internal server error")
			return
		}

		render.JSON(w, r, server)
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

		serverID, err := ParseUUIDParam(r, "serverID")
		if err != nil {
			log.Debug("invalid serverID parameter", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid serverID")
			return
		}

		// Check server existence and ownership
		server, err := h.serverService.GetServer(r.Context(), serverID)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				RenderError(w, r, http.StatusNotFound, CodeNotFound, "server not found")
				return
			}
			log.Error("failed to verify server ownership", zap.Error(err))
			RenderError(w, r, http.StatusInternalServerError, CodeInternalServerError, "internal server error")
			return
		}

		if server.OwnerID != actorID {
			RenderError(w, r, http.StatusForbidden, CodeForbidden, "only server owner can update the server")
			return
		}

		var req UpdateServerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Debug("failed to decode request body", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
			return
		}

		updatedServer, err := h.serverService.UpdateServer(r.Context(), serverID, req.Name, req.AvatarURL)
		if err != nil {
			log.Error("failed to update server", zap.Error(err))
			RenderError(w, r, http.StatusInternalServerError, CodeInternalServerError, "failed to update server")
			return
		}

		render.JSON(w, r, updatedServer)
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

		serverID, err := ParseUUIDParam(r, "serverID")
		if err != nil {
			log.Debug("invalid serverID parameter", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid serverID")
			return
		}

		// Check ownership
		server, err := h.serverService.GetServer(r.Context(), serverID)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				RenderError(w, r, http.StatusNotFound, CodeNotFound, "server not found")
				return
			}
			log.Error("failed to verify server ownership", zap.Error(err))
			RenderError(w, r, http.StatusInternalServerError, CodeInternalServerError, "internal server error")
			return
		}

		if server.OwnerID != actorID {
			RenderError(w, r, http.StatusForbidden, CodeForbidden, "only server owner can delete the server")
			return
		}

		if err := h.serverService.DeleteServer(r.Context(), serverID); err != nil {
			log.Error("failed to delete server", zap.Error(err))
			RenderError(w, r, http.StatusInternalServerError, CodeInternalServerError, "failed to delete server")
			return
		}

		render.Status(r, http.StatusNoContent)
		render.JSON(w, r, nil)
	}
}
