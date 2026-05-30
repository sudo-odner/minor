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

type ChannelService interface {
	CreateChannel(ctx context.Context, actorID uuid.UUID, serverID uuid.UUID, name string, typeChannel models.ChannelType, parentID *uuid.UUID) (*models.Channel, error)
	GetChannel(ctx context.Context, channelID uuid.UUID) (*models.Channel, error)
	GetServerChannel(ctx context.Context, serverID uuid.UUID) ([]models.Channel, error)
	UpdateChannel(ctx context.Context, actorID uuid.UUID, channelID, serverID uuid.UUID, name *string, parentID *uuid.UUID) (*models.Channel, error)
	DeleteChannel(ctx context.Context, actorID, serverID, channelID uuid.UUID) error
	MoveChannel(ctx context.Context, actorID uuid.UUID, serverID, channelID uuid.UUID, newParentID *uuid.UUID, newPos int) error
}

type ChannelHandler struct {
	log            *zap.Logger
	channelService ChannelService
}

func NewChannelHandler(log *zap.Logger, channelService ChannelService) *ChannelHandler {
	return &ChannelHandler{
		log:            log,
		channelService: channelService,
	}
}

type CreateChannelRequest struct {
	Name     string             `json:"name"`
	Type     models.ChannelType `json:"type"`
	ParentID *uuid.UUID         `json:"parent_id"`
}

type ChannelResponse struct {
	ID        string  `json:"id"`
	ServerID  string  `json:"server_id"`
	Name      string  `json:"name"`
	Type      int     `json:"type"`
	ParentID  *string `json:"parent_id,omitempty"`
	Position  int     `json:"position"`
	CreatedAt string  `json:"created_at"`
}

func (h *ChannelHandler) CreateChannel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.channel.CreateChannel"
		log := h.log.With(zap.String("op", op))

		actorID, err := ParseUserID(r)
		if err != nil {
			log.Debug("unauthorized request", zap.Error(err))
			RenderError(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
			return
		}

		serverID, err := ParseUUIDParam(r, "server_id")
		if err != nil {
			log.Debug("invalid server_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid server_id")
			return
		}

		var req CreateChannelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Debug("failed to decode body", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
			return
		}

		if req.Name == "" {
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "channel name is required")
			return
		}

		ch, err := h.channelService.CreateChannel(r.Context(), actorID, serverID, req.Name, req.Type, req.ParentID)
		if err != nil {
			log.Error("failed to create channel", zap.Error(err))
			RenderModelError(w, r, err, "failed to create channel")
			return
		}

		var parentID *string
		if ch.ParentID != nil && *ch.ParentID != uuid.Nil {
			pid := ch.ParentID.String()
			parentID = &pid
		}
		render.Status(r, http.StatusCreated)
		render.JSON(w, r, ChannelResponse{
			ID:        ch.ID.String(),
			ServerID:  ch.ServerID.String(),
			Name:      ch.Name,
			Type:      int(ch.Type),
			ParentID:  parentID,
			Position:  ch.Position,
			CreatedAt: ch.CreatedAt.Format(time.RFC3339),
		})
	}
}

func (h *ChannelHandler) GetServerChannels() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.channel.GetServerChannels"
		log := h.log.With(zap.String("op", op))

		serverID, err := ParseUUIDParam(r, "server_id")
		if err != nil {
			log.Debug("invalid server_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid server_id")
			return
		}

		channels, err := h.channelService.GetServerChannel(r.Context(), serverID)
		if err != nil {
			log.Error("failed to get channels", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		res := make([]ChannelResponse, len(channels))
		for i := range channels {
			var parentID *string
			if channels[i].ParentID != nil && *channels[i].ParentID != uuid.Nil {
				pid := channels[i].ParentID.String()
				parentID = &pid
			}
			res[i] = ChannelResponse{
				ID:        channels[i].ID.String(),
				ServerID:  channels[i].ServerID.String(),
				Name:      channels[i].Name,
				Type:      int(channels[i].Type),
				ParentID:  parentID,
				Position:  channels[i].Position,
				CreatedAt: channels[i].CreatedAt.Format(time.RFC3339),
			}
		}
		render.JSON(w, r, res)
	}
}

func (h *ChannelHandler) GetChannel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.channel.GetChannel"
		log := h.log.With(zap.String("op", op))

		channelID, err := ParseUUIDParam(r, "channel_id")
		if err != nil {
			log.Debug("invalid channel_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid channel_id")
			return
		}

		ch, err := h.channelService.GetChannel(r.Context(), channelID)
		if err != nil {
			log.Error("failed to get channel", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		var parentID *string
		if ch.ParentID != nil && *ch.ParentID != uuid.Nil {
			pid := ch.ParentID.String()
			parentID = &pid
		}
		render.JSON(w, r, ChannelResponse{
			ID:        ch.ID.String(),
			ServerID:  ch.ServerID.String(),
			Name:      ch.Name,
			Type:      int(ch.Type),
			ParentID:  parentID,
			Position:  ch.Position,
			CreatedAt: ch.CreatedAt.Format(time.RFC3339),
		})
	}
}

type UpdateChannelRequest struct {
	Name     *string    `json:"name"`
	ParentID *uuid.UUID `json:"parent_id"`
}

func (h *ChannelHandler) UpdateChannel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.channel.UpdateChannel"
		log := h.log.With(zap.String("op", op))

		actorID, err := ParseUserID(r)
		if err != nil {
			log.Debug("unauthorized request", zap.Error(err))
			RenderError(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
			return
		}

		serverID, err := ParseUUIDParam(r, "server_id")
		if err != nil {
			log.Debug("invalid server_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid server_id")
			return
		}

		channelID, err := ParseUUIDParam(r, "channel_id")
		if err != nil {
			log.Debug("invalid channel_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid channel_id")
			return
		}

		var req UpdateChannelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Debug("failed to decode body", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
			return
		}

		ch, err := h.channelService.UpdateChannel(r.Context(), actorID, channelID, serverID, req.Name, req.ParentID)
		if err != nil {
			log.Error("failed to update channel", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		var parentID *string
		if ch.ParentID != nil && *ch.ParentID != uuid.Nil {
			pid := ch.ParentID.String()
			parentID = &pid
		}
		render.JSON(w, r, ChannelResponse{
			ID:        ch.ID.String(),
			ServerID:  ch.ServerID.String(),
			Name:      ch.Name,
			Type:      int(ch.Type),
			ParentID:  parentID,
			Position:  ch.Position,
			CreatedAt: ch.CreatedAt.Format(time.RFC3339),
		})
	}
}

func (h *ChannelHandler) DeleteChannel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.channel.DeleteChannel"
		log := h.log.With(zap.String("op", op))

		actorID, err := ParseUserID(r)
		if err != nil {
			log.Debug("unauthorized request", zap.Error(err))
			RenderError(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
			return
		}

		serverID, err := ParseUUIDParam(r, "server_id")
		if err != nil {
			log.Debug("invalid server_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid server_id")
			return
		}

		channelID, err := ParseUUIDParam(r, "channel_id")
		if err != nil {
			log.Debug("invalid channel_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid channel_id")
			return
		}

		err = h.channelService.DeleteChannel(r.Context(), actorID, serverID, channelID)
		if err != nil {
			log.Error("failed to delete channel", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		render.Status(r, http.StatusNoContent)
		render.JSON(w, r, nil)
	}
}

type MoveChannelRequest struct {
	NewParentID *uuid.UUID `json:"new_parent_id"`
	NewPosition int        `json:"new_position"`
}

func (h *ChannelHandler) MoveChannel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.channel.MoveChannel"
		log := h.log.With(zap.String("op", op))

		actorID, err := ParseUserID(r)
		if err != nil {
			log.Debug("unauthorized request", zap.Error(err))
			RenderError(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
			return
		}

		serverID, err := ParseUUIDParam(r, "server_id")
		if err != nil {
			log.Debug("invalid server_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid server_id")
			return
		}

		channelID, err := ParseUUIDParam(r, "channel_id")
		if err != nil {
			log.Debug("invalid channel_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid channel_id")
			return
		}

		var req MoveChannelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Debug("failed to decode body", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
			return
		}

		err = h.channelService.MoveChannel(r.Context(), actorID, serverID, channelID, req.NewParentID, req.NewPosition)
		if err != nil {
			log.Error("failed to move channel", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		render.Status(r, http.StatusNoContent)
		render.JSON(w, r, nil)
	}
}
