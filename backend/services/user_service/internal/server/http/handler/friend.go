package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/models"
	"go.uber.org/zap"
)

type FriendService interface {
	SendFriendRequest(ctx context.Context, userID, friendID uuid.UUID) error
	FriendList(ctx context.Context, userID uuid.UUID) ([]*models.RelationshipPreview, error)
	FriendRequestList(ctx context.Context, userID uuid.UUID) ([]*models.RelationshipPreview, error)
	AcceptFriendRequest(ctx context.Context, actorID, targetID uuid.UUID) error
	DenyFriendRequest(ctx context.Context, actorID, targetID uuid.UUID) error
	BlockUser(ctx context.Context, actorID, target uuid.UUID) error
	RemoveFriend(ctx context.Context, actorID, targetID uuid.UUID) error
}

type FriendHandler struct {
	log           *zap.Logger
	friendService FriendService
}

func NewFriendHandler(log *zap.Logger, friendService FriendService) *FriendHandler {
	return &FriendHandler{
		log:           log,
		friendService: friendService,
	}
}

type RelationshipPreviewResponse struct {
	UserID    string  `json:"user_id"`
	Username  string  `json:"username"`
	AvatarURL *string `json:"avatar_url"`
	Status    string  `json:"status"`
}

func toRelationshipPreviewResponse(rp *models.RelationshipPreview) RelationshipPreviewResponse {
	return RelationshipPreviewResponse{
		UserID:    rp.UserID.String(),
		Username:  rp.Username,
		AvatarURL: rp.AvatarURL,
		Status:    rp.Status.String(),
	}
}

func (h *FriendHandler) SendFriendRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.friend.SendFriendRequest"
		log := h.log.With(zap.String("op", op))

		actorID, err := ParseUserID(r)
		if err != nil {
			log.Debug("unauthorized request", zap.Error(err))
			RenderError(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
			return
		}

		friendID, err := ParseUUIDParam(r, "friend_id")
		if err != nil {
			log.Debug("invalid friend_id parameter", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid friend_id")
			return
		}

		if err := h.friendService.SendFriendRequest(r.Context(), actorID, friendID); err != nil {
			log.Error("failed to send friend request", zap.Error(err))
			RenderModelError(w, r, err, "failed to send friend request")
			return
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, map[string]string{"message": "friend request sent"})
	}
}

func (h *FriendHandler) FriendList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.friend.FriendList"
		log := h.log.With(zap.String("op", op))

		actorID, err := ParseUserID(r)
		if err != nil {
			log.Debug("unauthorized request", zap.Error(err))
			RenderError(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
			return
		}

		list, err := h.friendService.FriendList(r.Context(), actorID)
		if err != nil {
			log.Error("failed to get friend list", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		resp := make([]RelationshipPreviewResponse, len(list))
		for i, rp := range list {
			resp[i] = toRelationshipPreviewResponse(rp)
		}

		render.JSON(w, r, resp)
	}
}

func (h *FriendHandler) FriendRequestList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.friend.FriendRequestList"
		log := h.log.With(zap.String("op", op))

		actorID, err := ParseUserID(r)
		if err != nil {
			log.Debug("unauthorized request", zap.Error(err))
			RenderError(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
			return
		}

		list, err := h.friendService.FriendRequestList(r.Context(), actorID)
		if err != nil {
			log.Error("failed to get friend requests list", zap.Error(err))
			RenderModelError(w, r, err, "internal server error")
			return
		}

		resp := make([]RelationshipPreviewResponse, len(list))
		for i, rp := range list {
			resp[i] = toRelationshipPreviewResponse(rp)
		}

		render.JSON(w, r, resp)
	}
}

type AnswerFriendRequestInput struct {
	Status string `json:"status"`
}

func (h *FriendHandler) AnswerFriendRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.friend.AnswerFriendRequest"
		log := h.log.With(zap.String("op", op))

		actorID, err := ParseUserID(r)
		if err != nil {
			log.Debug("unauthorized request", zap.Error(err))
			RenderError(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
			return
		}

		friendID, err := ParseUUIDParam(r, "friend_id")
		if err != nil {
			log.Debug("invalid friend_id parameter", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid friend_id")
			return
		}

		var input AnswerFriendRequestInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			log.Debug("failed to decode body", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid request body")
			return
		}

		if input.Status == "accepted" {
			err = h.friendService.AcceptFriendRequest(r.Context(), actorID, friendID)
		} else if input.Status == "deny" {
			err = h.friendService.DenyFriendRequest(r.Context(), actorID, friendID)
		} else {
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "status must be accepted or deny")
			return
		}

		if err != nil {
			log.Error("failed to answer friend request", zap.Error(err))
			RenderModelError(w, r, err, "failed to answer friend request")
			return
		}

		render.JSON(w, r, map[string]string{"message": "friend request answered successfully"})
	}
}

func (h *FriendHandler) DeleteFriendship() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.friend.DeleteFriendship"
		log := h.log.With(zap.String("op", op))

		actorID, err := ParseUserID(r)
		if err != nil {
			log.Debug("unauthorized request", zap.Error(err))
			RenderError(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
			return
		}

		friendID, err := ParseUUIDParam(r, "friend_id")
		if err != nil {
			log.Debug("invalid friend_id parameter", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid friend_id")
			return
		}

		if err := h.friendService.RemoveFriend(r.Context(), actorID, friendID); err != nil {
			log.Error("failed to remove friend", zap.Error(err))
			RenderModelError(w, r, err, "failed to remove friend")
			return
		}

		render.Status(r, http.StatusNoContent)
		render.JSON(w, r, nil)
	}
}

func (h *FriendHandler) BlockUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.friend.BlockUser"
		log := h.log.With(zap.String("op", op))

		actorID, err := ParseUserID(r)
		if err != nil {
			log.Debug("unauthorized request", zap.Error(err))
			RenderError(w, r, http.StatusUnauthorized, CodeUnauthorized, "unauthorized")
			return
		}

		friendID, err := ParseUUIDParam(r, "friend_id")
		if err != nil {
			log.Debug("invalid friend_id parameter", zap.Error(err))
			RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, "invalid friend_id")
			return
		}

		if err := h.friendService.BlockUser(r.Context(), actorID, friendID); err != nil {
			log.Error("failed to block user", zap.Error(err))
			RenderModelError(w, r, err, "failed to block user")
			return
		}

		render.JSON(w, r, map[string]string{"message": "user blocked successfully"})
	}
}
