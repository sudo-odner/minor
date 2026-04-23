package messages

import (
	"context"
	"net/http"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type MessageService interface {
	GetMessages()
	DeleteMessage(ctx context.Context)
}

type MessageHandler struct {
	log            *zap.Logger
	messageService MessageService
	validate       *validator.Validate
}

func New(log *zap.Logger, messageService MessageService) *MessageHandler {
	return &MessageHandler{
		log:            log,
		messageService: messageService,
		validate:       validator.New(),
	}
}

func (mh *MessageHandler) SendMessage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (mh *MessageHandler) GetMessages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

func (mh *MessageHandler) DeleteMessage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}

// In future, because Discord use lazy delete(delete before time) for delete all messages in channel
// func (mh *MessageHandler) DeleteAllMessage() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 	}
// }
