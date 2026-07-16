package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
	"go.uber.org/zap"
)

type MemberService interface {
	AddMember(ctx context.Context, serverID, userID uuid.UUID, nickname string) (*models.Member, error)
	GetServerMember(ctx context.Context, serverID, userID uuid.UUID) (*models.Member, error)
	GetServerMembers(ctx context.Context, serverID uuid.UUID) ([]models.Member, error)
	RemoveMember(ctx context.Context, actorID uuid.UUID, serverID uuid.UUID, targetUserID uuid.UUID) error
	UpdateNickname(ctx context.Context, actorID, serverID, targetUserID uuid.UUID, nickname string) error
	AddRoleToMember(ctx context.Context, actorID, serverID, targetUserID, roleID uuid.UUID) error
	RemoveRoleFromMember(ctx context.Context, actorID, serverID, targetUserID, roleID uuid.UUID) error
}

type MemberHandler struct {
	log           *zap.Logger
	memberService MemberService
}

func NewMemberHandler(log *zap.Logger, memberService MemberService) *MemberHandler {
	return &MemberHandler{
		log:           log,
		memberService: memberService,
	}
}

type AddMemberRequest struct {
	UserID   uuid.UUID `json:"user_id"`
	Nickname string    `json:"nickname,omitempty"`
}

type MemberResponse struct {
	ServerID  string         `json:"server_id"`
	UserID    string         `json:"user_id"`
	Nickname  *string        `json:"nickname,omitempty"`
	Username  *string        `json:"username"`
	AvatarURL *string        `json:"avatar_url,omitempty"`
	Status    string         `json:"status"`
	JoinedAt  string         `json:"joined_at"`
	Roles     []RoleResponse `json:"roles"`
}

func mapRolesToResponse(roles []models.Role) []RoleResponse {
	if roles == nil {
		return []RoleResponse{}
	}
	res := make([]RoleResponse, len(roles))
	for i, r := range roles {
		res[i] = RoleResponse{
			ID:       r.ID.String(),
			Name:     r.Name,
			Position: r.Position,
		}
	}
	return res
}

func (h *MemberHandler) AddMember() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.member.AddMember"
		log := h.log.With(zap.String("op", op))

		serverIDStr := chi.URLParam(r, "server_id")
		serverID, err := uuid.Parse(serverIDStr)
		if err != nil {
			log.Warn("invalid server_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInternalServerError, "invalid_server_id")
			return
		}

		userIDStr := r.Header.Get("X-User-ID")
		if userIDStr == "" {
			log.Error("X-User-ID header is missing")
			RenderError(w, r, http.StatusUnauthorized, CodeInternalServerError, "unauthorized")
			return
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			log.Error("invalid user_id from header", zap.Error(err))
			RenderError(w, r, http.StatusInternalServerError, CodeInternalServerError, "internal_error")
			return
		}

		var req AddMemberRequest
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil && err.Error() != "EOF" {
			log.Warn("failed to decode body", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInternalServerError, "invalid_request_body")
			return
		}

		targetUserID := userID
		if req.UserID != uuid.Nil {
			targetUserID = req.UserID
		}

		_, err = h.memberService.AddMember(r.Context(), serverID, targetUserID, req.Nickname)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				h.log.Warn("user already a member", zap.String("user_id", targetUserID.String()))
				RenderError(w, r, http.StatusConflict, "already_member", "User is already a member of this server")
				return
			}
			log.Error("failed to add member", zap.Error(err))
			RenderError(w, r, http.StatusInternalServerError, CodeInternalServerError, "failed_to_join")
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func (h *MemberHandler) GetServerMembers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.member.GetServerMembers"
		log := h.log.With(zap.String("op", op))

		serverID, err := ParseUUIDParam(r, "server_id")
		if err != nil {
			log.Debug("invalid server_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid server_id")
			return
		}

		membersList, err := h.memberService.GetServerMembers(r.Context(), serverID)
		if err != nil {
			log.Error("failed to get members", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		res := make([]MemberResponse, len(membersList))
		for i := range membersList {
			res[i] = MemberResponse{
				ServerID:  membersList[i].ServerID.String(),
				UserID:    membersList[i].UserID.String(),
				Username:  &membersList[i].Username,
				AvatarURL: &membersList[i].AvatarURL,
				Status:    membersList[i].Status,
				Nickname:  membersList[i].Nickname,
				JoinedAt:  membersList[i].JoinedAt.Format(time.RFC3339),
				Roles:     mapRolesToResponse(membersList[i].Roles),
			}
		}
		render.JSON(w, r, res)
	}
}

func (h *MemberHandler) GetServerMember() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.member.GetServerMember"
		log := h.log.With(zap.String("op", op))

		serverID, err := ParseUUIDParam(r, "server_id")
		if err != nil {
			log.Debug("invalid server_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid server_id")
			return
		}

		userID, err := ParseUUIDParam(r, "user_id")
		if err != nil {
			log.Debug("invalid user_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid user_id")
			return
		}

		m, err := h.memberService.GetServerMember(r.Context(), serverID, userID)
		if err != nil {
			log.Error("failed to get member", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		render.JSON(w, r, MemberResponse{
			ServerID:  m.ServerID.String(),
			UserID:    m.UserID.String(),
			Username:  &m.Username,
			AvatarURL: &m.AvatarURL,
			Status:    m.Status,
			Nickname:  m.Nickname,
			JoinedAt:  m.JoinedAt.Format(time.RFC3339),
			Roles:     mapRolesToResponse(m.Roles),
		})
	}
}

func (h *MemberHandler) RemoveMember() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.member.RemoveMember"
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

		userID, err := ParseUUIDParam(r, "user_id")
		if err != nil {
			log.Debug("invalid user_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid user_id")
			return
		}

		err = h.memberService.RemoveMember(r.Context(), actorID, serverID, userID)
		if err != nil {
			log.Error("failed to remove member", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		render.Status(r, http.StatusNoContent)
		render.JSON(w, r, nil)
	}
}

type UpdateNicknameRequest struct {
	Nickname string `json:"nickname"`
}

func (h *MemberHandler) UpdateNickname() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.member.UpdateNickname"
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

		userID, err := ParseUUIDParam(r, "user_id")
		if err != nil {
			log.Debug("invalid user_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid user_id")
			return
		}

		var req UpdateNicknameRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Debug("failed to decode body", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
			return
		}

		err = h.memberService.UpdateNickname(r.Context(), actorID, serverID, userID, req.Nickname)
		if err != nil {
			log.Error("failed to update nickname", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		render.Status(r, http.StatusNoContent)
		render.JSON(w, r, nil)
	}
}

func (h *MemberHandler) AddRoleToMember() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.member.AddRoleToMember"
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

		userID, err := ParseUUIDParam(r, "user_id")
		if err != nil {
			log.Debug("invalid user_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid user_id")
			return
		}

		roleID, err := ParseUUIDParam(r, "role_id")
		if err != nil {
			log.Debug("invalid role_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid role_id")
			return
		}

		err = h.memberService.AddRoleToMember(r.Context(), actorID, serverID, userID, roleID)
		if err != nil {
			log.Error("failed to add role to member", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		render.Status(r, http.StatusNoContent)
		render.JSON(w, r, nil)
	}
}

func (h *MemberHandler) RemoveRoleFromMember() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.member.RemoveRoleFromMember"
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

		userID, err := ParseUUIDParam(r, "user_id")
		if err != nil {
			log.Debug("invalid user_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid user_id")
			return
		}

		roleID, err := ParseUUIDParam(r, "role_id")
		if err != nil {
			log.Debug("invalid role_id", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid role_id")
			return
		}

		err = h.memberService.RemoveRoleFromMember(r.Context(), actorID, serverID, userID, roleID)
		if err != nil {
			log.Error("failed to remove role from member", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		render.Status(r, http.StatusNoContent)
		render.JSON(w, r, nil)
	}
}
