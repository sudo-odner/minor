package messages

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

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
	GetMessage(ctx context.Context, userID, channelID, messageID uuid.UUID) (*models.Message, error)
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
			log.Debug("falied parce 'X-User-ID' in uuid format", zap.Error(err))
			handler.RenderError(w, r, http.StatusUnauthorized, handler.CodeInvalidRequset, "unauthorized")
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
				errInfo := "invalid 'reply_to' uuid"
				log.Debug(errInfo, zap.String("reply_to", req.ReplyTo))
				handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, errInfo)
				return
			}
			replyToUUID = &parcedReply
		}

		channelIDStr := chi.URLParam(r, "channel_id")
		channelID, err := uuid.Parse(channelIDStr)
		if err != nil {
			errInfo := "falied parce 'channel_id' in uuid format"
			log.Debug(errInfo, zap.Error(err))
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, errInfo)
			return
		}

		msg, err := mh.messageService.SaveMessage(r.Context(), userID, channelID, req.Content, replyToUUID)
		if err != nil {
			// TODO: Обработка ошибок
			log.Debug("invalid save message", zap.Error(err))
			handler.RenderError(w, r, http.StatusInternalServerError, handler.CodeInternalServerError, "internal error")
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
		const op = "http.handler.message.GetMessage"
		log := mh.log.With(zap.String("op", op))

		userID, err := parceUUIDHeader(w, r, "X-User-ID")
		if err != nil {
			log.Debug("falied parce 'X-User-ID' in uuid format", zap.Error(err))
			handler.RenderError(w, r, http.StatusUnauthorized, handler.CodeInvalidRequset, "unauthorized")
			return
		}

		channelIDStr := chi.URLParam(r, "channel_id")
		channelID, err := uuid.Parse(channelIDStr)
		if err != nil {
			errInfo := "falied parce 'channel_id' in uuid format"
			log.Debug(errInfo, zap.Error(err))
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, errInfo)
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit := 20
		if limitStr != "" {
			var err error
			limit, err = strconv.Atoi(limitStr)
			if err != nil {
				errInfo := "falied parce 'limit' in number"
				log.Debug(errInfo, zap.Error(err))
				handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, "'limit' must be a valid number")
				return
			}
			if limit <= 0 || limit > 100 {
				handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, "'limit' must by around 0 < 'limit' < 100")
				return
			}
		}

		beforeIDStr := r.URL.Query().Get("before_id")
		var beforeID *uuid.UUID
		if beforeIDStr != "" {
			parsedUUID, err := uuid.Parse(beforeIDStr)
			if err != nil {
				errInfo := "falied parce 'before_id' in uuid format"
				log.Debug(errInfo, zap.Error(err))
				handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, errInfo)
				return
			}
			beforeID = &parsedUUID
		} else {
			log.Debug("'before_id' not found")
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, "'before_id' parameter is required")
			return
		}

		msgs, err := mh.messageService.GetMessages(r.Context(), userID, channelID, limit, beforeID)
		if err != nil {
			// TODO: Обратотка ошибок сервера
			if errors.Is(err, models.ErrMessageNotFound) {
				errInfo := fmt.Sprintf("message before '%s' not found", beforeID.String())
				log.Debug(errInfo, zap.Error(err))
				handler.RenderError(w, r, http.StatusNotFound, handler.CodeNotFound, errInfo)
				return
			}
			log.Warn("error find message 'before_id'", zap.Error(err))
			handler.RenderError(w, r, http.StatusInternalServerError, handler.CodeInternalServerError, "internal error")
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
			log.Debug("falied parce 'X-User-ID' in uuid format", zap.Error(err))
			handler.RenderError(w, r, http.StatusUnauthorized, handler.CodeInvalidRequset, "unauthorized")
			return
		}

		channelIDStr := chi.URLParam(r, "channel_id")
		channelID, err := uuid.Parse(channelIDStr)
		if err != nil {
			errInfo := "falied parce 'channel_id' in uuid format"
			log.Debug(errInfo, zap.Error(err))
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, errInfo)
			return
		}

		messageIDStr := chi.URLParam(r, "message_id")
		messageID, err := uuid.Parse(messageIDStr)
		if err != nil {
			errInfo := "falied parce 'message_id' in uuid format"
			log.Debug(errInfo, zap.Error(err))
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, errInfo)
			return
		}

		msg, err := mh.messageService.GetMessage(r.Context(), userID, channelID, messageID)
		if err != nil {
			// TODO: Обратотка ошибок сервера
			if errors.Is(err, models.ErrMessageNotFound) {
				errInfo := fmt.Sprintf("message with '%s' not found", messageID.String())
				log.Debug(errInfo, zap.Error(err))
				handler.RenderError(w, r, http.StatusNotFound, handler.CodeNotFound, err.Error())
				return
			}
			log.Warn("error find message", zap.Error(err))
			handler.RenderError(w, r, http.StatusInternalServerError, handler.CodeInternalServerError, "internal error")
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
		const op = "http.handler.message.DeleteMessage"
		log := mh.log.With(zap.String("op", op))

		userID, err := parceUUIDHeader(w, r, "X-User-ID")
		if err != nil {
			log.Debug("falied parce 'X-User-ID' in uuid format", zap.Error(err))
			handler.RenderError(w, r, http.StatusUnauthorized, handler.CodeInvalidRequset, "unauthorized")
			return
		}

		channelIDStr := chi.URLParam(r, "channel_id")
		channelID, err := uuid.Parse(channelIDStr)
		if err != nil {
			errInfo := "falied parce 'channel_id' in uuid format"
			log.Debug(errInfo, zap.Error(err))
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, errInfo)
			return
		}

		messageIDStr := chi.URLParam(r, "message_id")
		messageID, err := uuid.Parse(messageIDStr)
		if err != nil {
			errInfo := "falied parce 'message_id' in uuid format"
			log.Debug(errInfo, zap.Error(err))
			handler.RenderError(w, r, http.StatusBadRequest, handler.CodeInvalidRequset, errInfo)
			return
		}

		// TODO: change to 'ok, err :='
		if err := mh.messageService.DeleteMessage(r.Context(), userID, channelID, messageID); err != nil {
			// TODO: Обратотка ошибок сервера
			if errors.Is(err, models.ErrMessageNotFound) {
				errInfo := fmt.Sprintf("message with '%s' not found", messageID.String())
				log.Debug(errInfo, zap.Error(err))
				handler.RenderError(w, r, http.StatusNotFound, handler.CodeNotFound, errInfo)
				return
			}
			log.Warn("error delete message", zap.Error(err))
			handler.RenderError(w, r, http.StatusInternalServerError, handler.CodeInternalServerError, "internal error")
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
