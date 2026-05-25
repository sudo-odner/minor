package messages

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/models"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/server/http/handler"
	"go.uber.org/zap"
)

type MessageService interface {
	SaveMessage(ctx context.Context, userID, channelID uuid.UUID, content string, replyTo *uuid.UUID) (*models.Message, error)
	GetMessages(ctx context.Context, userID, channelID uuid.UUID, limit int, beforeID *uuid.UUID) ([]models.Message, error)
	GetMessage(ctx context.Context, userID, channelID, messageID uuid.UUID) (models.Message, error)
	DeleteMessage(ctx context.Context, userID, channelID, messageID uuid.UUID) error
}

type MessageHandler struct {
	log            *zap.Logger
	messageService MessageService
	validate       *validator.Validate
}

// TODO: Need to write logs

func New(log *zap.Logger, messageService MessageService) *MessageHandler {
	return &MessageHandler{
		log:            log,
		messageService: messageService,
		validate:       validator.New(),
	}
}

func parceUUIDHeader(w http.ResponseWriter, r *http.Request, headerName string) (uuid.UUID, error) {
	idStr := r.Header.Get(headerName)
	if idStr == "" {
		return uuid.Nil, fmt.Errorf("header is required: %s", headerName)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid format in %s", idStr)
	}

	return id, nil
}

// TODO: revrite channel id from url path
func (mh *MessageHandler) SendMessage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.message.SendMessage"
		log := mh.log.With(zap.String("op", op))

		userID, err := parceUUIDHeader(w, r, "X-User-ID")
		if err != nil {
			log.Debug("falied parce uuid", zap.Error(err))
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, err.Error())
			return
		}

		var req ReqSaveMessage
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Debug("failed to decode request body", zap.Error(err))
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, "invalid request body")
			return
		}

		if err := mh.validate.Struct(req); err != nil {
			log.Debug("validation falied", zap.Error(err))
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, "invalid request body")
			return
		}

		var replyToUUID *uuid.UUID
		if req.ReplyTo != "" {
			parcedReply, err := uuid.Parse(req.ReplyTo)
			if err != nil {
				log.Debug("invalid reply_to uuid", zap.String("reply_to", req.ReplyTo))
				handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, "invalid reply_to uuid")
				return
			}
			replyToUUID = &parcedReply
		}
		channelID, err := uuid.Parse(req.ChannelID)
		if err != nil {
			log.Debug("invalid channel_id uuid", zap.String("channel_id", req.ChannelID))
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, "invalid reply_to uuid")
			return
		}

		msg, err := mh.messageService.SaveMessage(r.Context(), userID, channelID, req.Content, replyToUUID)
		if err != nil {
			// TODO: Обработка ошибок
			log.Debug("invalid save message", zap.Error(err))
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, "invalid reply_to uuid")
			return
		}

		var replyToStr string
		if msg.ReplyTo != nil {
			replyToStr = msg.ReplyTo.String()
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, Message{
			MessageID: msg.MessageID.String(),
			ChannelID: msg.ChannelID.String(),
			AuthorID:  msg.UserID.String(),
			Content:   msg.Content,
			ReplyTo:   replyToStr,
			CreateAt:  msg.CreatedAt,
		})
	}
}

func (mh *MessageHandler) GetMessages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.message.SendMessage"
		log := mh.log.With(zap.String("op", op))

		userID, err := parceUUIDHeader(w, r, "X-User-ID")
		if err != nil {
			mh.log.Debug("falied parce uuid", zap.Error(err))
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, err.Error())
			return
		}

		channelIDStr := chi.URLParam(r, "channel_id")
		channerID, err := uuid.Parse(channelIDStr)
		if err != nil {
			fmt.Println(channelIDStr)
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, "invalid chennel_id uuid")
			return
		}

		// TODO: Write normal limits
		msgs, err := mh.messageService.GetMessages(r.Context(), userID, channerID, 1, nil)
		if err != nil {
			// TODO: Обработка ошибок сервера
			log.Debug("invalid save message", zap.Error(err))
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, "invalid reply_to uuid")
			return
		}

		msgModels := make([]Message, len(msgs))
		var replyToStr string
		for _, msg := range msgs {
			if msg.ReplyTo != nil {
				replyToStr = msg.ReplyTo.String()
			}
			msgModels = append(msgModels, Message{
				MessageID: msg.MessageID.String(),
				ChannelID: msg.ChannelID.String(),
				AuthorID:  msg.UserID.String(),
				Content:   msg.Content,
				ReplyTo:   replyToStr,
				CreateAt:  msg.CreatedAt,
			})
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, ResGetMessages{
			Messages: msgModels,
		})
	}
}

func (mh *MessageHandler) GetMessage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.message.GetMessage"
		log := mh.log.With(zap.String("op", op))

		userID, err := parceUUIDHeader(w, r, "X-User-ID")
		if err != nil {
			log.Debug("falied parce channel_id in uuid format", zap.Error(err))
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, err.Error())
			return
		}

		channelIDStr := chi.URLParam(r, "channel_id")
		channelID, err := uuid.Parse(channelIDStr)
		if err != nil {
			log.Debug("falied parce channel_id in uuid format", zap.Error(err))
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, err.Error())
			return
		}

		messageIDStr := chi.URLParam(r, "message_id")
		messageID, err := uuid.Parse(messageIDStr)
		if err != nil {
			log.Debug("falied parce message_id in uuid format", zap.Error(err))
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, err.Error())
			return
		}

		msg, err := mh.messageService.GetMessage(r.Context(), userID, channelID, messageID)
		if err != nil {
			// TODO: Обратотка ошибок сервера
			if errors.Is(err, models.ErrMessageNotFound) {
				log.Debug(fmt.Sprintf("message with %s not found", messageID.String()))
				handler.RenderError(w, r, http.StatusNotFound, handler.CodeNotFound, err.Error())
				return
			}
			log.Warn("error find message", zap.Error(err))
			handler.RenderError(w, r, http.StatusInternalServerError, handler.CodeInternalServerError, err.Error())
			return
		}
		var strReplyTo string
		if msg.ReplyTo != nil {
			strReplyTo = msg.ReplyTo.String()
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, Message{
			MessageID: msg.MessageID.String(),
			ChannelID: msg.ChannelID.String(),
			AuthorID:  msg.UserID.String(),
			Content:   msg.Content,
			ReplyTo:   strReplyTo,
			CreateAt:  msg.CreatedAt,
		})
	}
}

func (mh *MessageHandler) DeleteMessage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "http.handler.message.SendMessage"
		log := mh.log.With(zap.String("op", op))

		userID, err := parceUUIDHeader(w, r, "X-User-ID")
		if err != nil {
			mh.log.Debug("falied parce uuid", zap.Error(err))
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, err.Error())
		}

		channelIDStr := chi.URLParam(r, "channel_id")
		channerID, err := uuid.Parse(channelIDStr)
		if err != nil {
			log.Debug("invalid chennle_id uuid")
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, "invalid chennel_id uuid")
			return
		}

		messageIDStr := chi.URLParam(r, "message_id")
		messageID, err := uuid.Parse(messageIDStr)
		if err != nil {
			log.Debug("invalid message_id uuid")
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, "invalid message_id uuid")
			return
		}

		// TODO: change to 'ok, err :='
		if err := mh.messageService.DeleteMessage(r.Context(), userID, channerID, messageID); err != nil {
			// TODO: custom error
			log.Debug("failed delete messageID")
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, "invalid")
			return
		}

		render.Status(r, http.StatusOK)
	}
}

// TODO: Implement later. Bulk deletion is heavy(in Cassandra), in Discord usess asynchronus soft-deletion
// func (mh *MessageHandler) DeleteAllMessage() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 	}
// }
