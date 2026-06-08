package password

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

type PasswordHTTPService interface {
	ForgotPassword(email string) (string, error)
}

type PasswordHTTPHandler struct {
	log *zap.Logger
	PasswordService PasswordHTTPService
}

func NewHTTPHandler(log *zap.Logger, passwordHTTPService PasswordHTTPService) *PasswordHTTPHandler {
	return &PasswordHTTPHandler{
		log: log,
		PasswordService: passwordHTTPService,
	}
}

func (ph *PasswordHTTPHandler) ForgotPassword(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: ForgotPassword
	}
}

func (ph *PasswordHTTPHandler) ResetPassword(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: ForgotPassword
	}
}