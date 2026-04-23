package auth

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/lib/cookie"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/models"
	"go.uber.org/zap"
)

type AuthorizationService interface {
	Login(ctx context.Context, logUser *models.LoginUser) (user *models.NormalizedUser, accessToken string, refreshToken string, err error)
	Register(ctx context.Context, regUser *models.RegisterUser) (user *models.NormalizedUser, accessToken string, refreshToken string, err error)
	Logout(ctx context.Context)
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

type RegisterResponse struct {
	User        models.NormalizedUser `json:"user"`
	AccessToken string                `json:"access_token"`
}

func (ah *AuthorizationHandler) Register(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const path = "server.http.handler.auth.Register"

		log := ah.log.With(
			zap.String("path", path),
			zap.String("request-id", middleware.GetReqID(r.Context())),
		)

		var regUser models.RegisterUser

		err := render.DecodeJSON(r.Body, &regUser)
		if err != nil {
			log.Error("failed to decode request body", zap.Error(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, "failed to decode body")

			return
		}

		normalizedUser, accessToken, refreshToken, err := ah.authService.Register(ctx, &regUser)
		if err != nil {
			log.Error("failed to register user", zap.Error(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, "failed to register user")

			return
		}
		
		log.Info("user registered successfully")

		cookie.SetCookie(w, refreshToken)

		render.JSON(w, r, RegisterResponse{
			User:        *normalizedUser,
			AccessToken: accessToken,
		})
	}
}

func (ah *AuthorizationHandler) Logout(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (ah *AuthorizationHandler) RefreshToken(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
