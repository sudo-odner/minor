package friend

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/model"
	"go.uber.org/zap"
)

type FriendService interface {
	SendFriendRequest(ctx context.Context, userID, friendID uuid.UUID) error                 // Запрос на дружбу
	FriendList(ctx context.Context, userID uuid.UUID) ([]*model.FriendPreview, error)        // Получить список друзей
	FriendRequestList(ctx context.Context, userID uuid.UUID) ([]*model.FriendPreview, error) // Получить список статусов запросов на дружбу
	AcceptFriendRequest(ctx context.Context, actorID, targetID uuid.UUID) error              // Принять запрос
	DenyFriendRequest(ctx context.Context, actorID, targetID uuid.UUID) error                // Отклонить запрос
	BlockUser(ctx context.Context, actorID, target uuid.UUID) error                          // Заблокировать
	RemoveFriend(ctx context.Context, actorID, targerID uuid.UUID) error                     // Удалить друга
}

type FriendHandler struct {
	log           *zap.Logger
	friendService FriendService
}

func New(log *zap.Logger, friendService FriendService) *FriendHandler {
	return &FriendHandler{
		log:           log,
		friendService: friendService,
	}
}

func (h *FriendHandler) SendFriendRequest(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *FriendHandler) FriendList(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *FriendHandler) FriendRequestList(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *FriendHandler) AnswerFriendRequest(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *FriendHandler) DeleteFriendship(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
