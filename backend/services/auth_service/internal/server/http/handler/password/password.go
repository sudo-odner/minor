package password

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/models"
	"go.uber.org/zap"
)

type PasswordHTTPService interface {
	ForgotPassword(ctx context.Context, payload *models.ForgotPasswordPayload) error
	ResetPassword(ctx context.Context, payload *models.ResetPasswordPayload) error
}

type PasswordHTTPHandler struct {
	log             *zap.Logger
	passwordService PasswordHTTPService
}

func NewHTTPHandler(log *zap.Logger, passwordHTTPService PasswordHTTPService) *PasswordHTTPHandler {
	return &PasswordHTTPHandler{
		log:             log,
		passwordService: passwordHTTPService,
	}
}

func (ph *PasswordHTTPHandler) ForgotPassword(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const path = "server.http.handler.password.ForgotPassword"

		ctx = r.Context()

		log := ph.log.With(
			zap.String("path", path),
			zap.String("req-id", middleware.GetReqID(ctx)),
		)

		var req models.ForgotPasswordPayload
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Warn("failed to decode request body", zap.Error(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "invalid request body"})
			return
		}

		err := ph.passwordService.ForgotPassword(ctx, &req)
		if err != nil {
			log.Warn("login failed", zap.Error(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": "internal server error"})
			return
		}

		log.Info("password reset start")

		render.Status(r, http.StatusOK)
	}
}

func (ph *PasswordHTTPHandler) ResetPassword(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: ForgotPassword
	}
}
