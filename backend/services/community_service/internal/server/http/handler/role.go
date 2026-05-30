package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
	"go.uber.org/zap"
)

type RoleService interface {
	CreateRole(ctx context.Context, actorID, serverID uuid.UUID, name string, permissions authz.Permission) (*models.Role, error)
	GetRole(ctx context.Context, roleID uuid.UUID) (*models.Role, error)
	GetServerRoles(ctx context.Context, serverID uuid.UUID) ([]models.Role, error)
	UpdateRole(ctx context.Context, actorID, serverID, roleID uuid.UUID, name *string, permission *authz.Permission) (*models.Role, error)
	DeleteRole(ctx context.Context, actorID, serverID, roleID uuid.UUID) error
	ReplaceChannelPermissionOverrides(ctx context.Context, actorID, serverID, channelID uuid.UUID, overrides []models.ChannelPermissionOverride) error
}

type RoleHandler struct {
	log         *zap.Logger
	roleService RoleService
}

func NewRoleHandler(log *zap.Logger, roleService RoleService) *RoleHandler {
	return &RoleHandler{
		log:         log,
		roleService: roleService,
	}
}

type CreateRoleRequest struct {
	Name        string           `json:"name"`
	Permissions authz.Permission `json:"permissions"`
}

type RoleResponse struct {
	ID         string `json:"id"`
	ServerID   string `json:"server_id"`
	Name       string `json:"name"`
	Permission uint64 `json:"permissions"`
	Position   int    `json:"position"`
	CreatedAt  string `json:"created_at"`
}

func (h *RoleHandler) CreateRole() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.role.CreateRole"
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

		var req CreateRoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Debug("failed to decode body", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
			return
		}

		if req.Name == "" {
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "role name is required")
			return
		}

		role, err := h.roleService.CreateRole(r.Context(), actorID, serverID, req.Name, req.Permissions)
		if err != nil {
			log.Error("failed to create role", zap.Error(err))
			RenderModelError(w, r, err, "failed to create role")
			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, RoleResponse{
			ID:         role.ID.String(),
			ServerID:   role.ServerID.String(),
			Name:       role.Name,
			Permission: uint64(role.Permission),
			Position:   role.Position,
			CreatedAt:  role.CreatedAt.Format(time.RFC3339),
		})
	}
}

func (h *RoleHandler) GetServerRoles() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.role.GetServerRoles"
		log := h.log.With(zap.String("op", op))

		serverID, err := ParseUUIDParam(r, "server_id")
		if err != nil {
			log.Debug("invalid server_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid server_id")
			return
		}

		roles, err := h.roleService.GetServerRoles(r.Context(), serverID)
		if err != nil {
			log.Error("failed to get roles", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		res := make([]RoleResponse, len(roles))
		for i := range roles {
			res[i] = RoleResponse{
				ID:         roles[i].ID.String(),
				ServerID:   roles[i].ServerID.String(),
				Name:       roles[i].Name,
				Permission: uint64(roles[i].Permission),
				Position:   roles[i].Position,
				CreatedAt:  roles[i].CreatedAt.Format(time.RFC3339),
			}
		}
		render.JSON(w, r, res)
	}
}

func (h *RoleHandler) GetRole() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.role.GetRole"
		log := h.log.With(zap.String("op", op))

		roleID, err := ParseUUIDParam(r, "role_id")
		if err != nil {
			log.Debug("invalid role_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid role_id")
			return
		}

		role, err := h.roleService.GetRole(r.Context(), roleID)
		if err != nil {
			log.Error("failed to get role", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		render.JSON(w, r, RoleResponse{
			ID:         role.ID.String(),
			ServerID:   role.ServerID.String(),
			Name:       role.Name,
			Permission: uint64(role.Permission),
			Position:   role.Position,
			CreatedAt:  role.CreatedAt.Format(time.RFC3339),
		})
	}
}

type UpdateRoleRequest struct {
	Name        *string           `json:"name"`
	Permissions *authz.Permission `json:"permissions"`
}

func (h *RoleHandler) UpdateRole() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.role.UpdateRole"
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

		roleID, err := ParseUUIDParam(r, "role_id")
		if err != nil {
			log.Debug("invalid role_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid role_id")
			return
		}

		var req UpdateRoleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Debug("failed to decode body", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
			return
		}

		role, err := h.roleService.UpdateRole(r.Context(), actorID, serverID, roleID, req.Name, req.Permissions)
		if err != nil {
			log.Error("failed to update role", zap.Error(err))
			RenderModelError(w, r, err, "failed to update role")
			return
		}

		render.JSON(w, r, RoleResponse{
			ID:         role.ID.String(),
			ServerID:   role.ServerID.String(),
			Name:       role.Name,
			Permission: uint64(role.Permission),
			Position:   role.Position,
			CreatedAt:  role.CreatedAt.Format(time.RFC3339),
		})
	}
}

func (h *RoleHandler) DeleteRole() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.role.DeleteRole"
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

		roleID, err := ParseUUIDParam(r, "role_id")
		if err != nil {
			log.Debug("invalid role_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid role_id")
			return
		}

		err = h.roleService.DeleteRole(r.Context(), actorID, serverID, roleID)
		if err != nil {
			log.Error("failed to delete role", zap.Error(err))
			RenderModelError(w, r, err, "failed to delete role")
			return
		}

		render.Status(r, http.StatusNoContent)
		render.JSON(w, r, nil)
	}
}

type OverridePayload struct {
	TargetType models.OverrideType `json:"target_type"`
	TargetID   uuid.UUID           `json:"target_id"`
	Allow      authz.Permission    `json:"allow"`
	Deny       authz.Permission    `json:"deny"`
}

func (h *RoleHandler) ReplaceChannelPermissionOverrides() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.role.ReplaceChannelPermissionOverrides"
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

		var payloads []OverridePayload
		if err := json.NewDecoder(r.Body).Decode(&payloads); err != nil {
			log.Debug("failed to decode overrides payload", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
			return
		}

		overrides := make([]models.ChannelPermissionOverride, len(payloads))
		for i, p := range payloads {
			overrides[i] = models.ChannelPermissionOverride{
				ChannelID:  channelID,
				TargetType: p.TargetType,
				TargetID:   p.TargetID,
				Allow:      p.Allow,
				Deny:       p.Deny,
			}
		}

		err = h.roleService.ReplaceChannelPermissionOverrides(r.Context(), actorID, serverID, channelID, overrides)
		if err != nil {
			log.Error("failed to replace channel overrides", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		render.Status(r, http.StatusNoContent)
		render.JSON(w, r, nil)
	}
}
