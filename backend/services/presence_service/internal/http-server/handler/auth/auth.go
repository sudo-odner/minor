package auth

import (
	"context"
	"net/http"

	"github.com/sudo-odner/minor/backend/services/presence_service/internal/models"
	"go.uber.org/zap"
)

type AuthorizationService interface {
	Login(ctx context.Context, email string, password string) (user *models.User, accessToken string, refreshToken string, err error)
	Register(ctx context.Context, email string, username string, password string) (user *models.User, accessToken string, refreshToken string, err error)
	// Logout()
	// RefreshToken()
	// VerifyEmail()
}

type AuthorizationHandler struct {
	authService AuthorizationService
	log         *zap.Logger
}

func New(authService AuthorizationService, log *zap.Logger) *AuthorizationHandler {
	return &AuthorizationHandler{
		authService: authService,
		log:         log,
	}
}

func (ah *AuthorizationHandler) Login(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (ah *AuthorizationHandler) Register(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (ah *AuthorizationHandler) Logout(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
