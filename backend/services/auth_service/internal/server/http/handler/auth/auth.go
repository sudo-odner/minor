package auth

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/lib/cookie"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/models"
	"go.uber.org/zap"
)

type AuthHTTPService interface {
	Login(ctx context.Context, loginUser *models.LoginUser, ip, userAgent string) (user *models.AuthResponse, err error)
	Register(ctx context.Context, registerUser *models.RegisterUser) (user *models.AuthResponse, err error)
	Logout(ctx context.Context, refreshToken string) (err error)
	VerifyAccessToken(ctx context.Context, token string) (*models.Claims, error)
}

type AuthHTTPHandler struct {
	authService AuthHTTPService
	log         *zap.Logger
}

func NewHTTPHandler(authService AuthHTTPService, log *zap.Logger) *AuthHTTPHandler {
	return &AuthHTTPHandler{
		authService: authService,
		log:         log,
	}
}

func (ah *AuthHTTPHandler) Login(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const path = "server.http.handler.auth.Login"

		ctx = r.Context()

		log := ah.log.With(
			zap.String("path", path),
			zap.String("req-id", middleware.GetReqID(ctx)),
		)

		var req models.LoginUser
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Warn("failed to decode request body", zap.Error(err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": "invalid request body"})
			return
		}

		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
			ip = realIP
		}
		userAgent := r.Header.Get("User-Agent")

		res, err := ah.authService.Login(ctx, &req, ip, userAgent)
		if err != nil {
			// if errors.Is(err, service.ErrInvalidCredentials) {
			// 	render.Status(r, http.StatusUnauthorized)
			// 	render.JSON(w, r, map[string]string{"error": "invalid email or password"})
			// 	return
			// }

			log.Error("login failed", zap.Error(err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, map[string]string{"error": "internal server error"})
			return
		}

		cookie.SetCookie(w, res.RefreshToken)

		log.Info("user logged in", zap.String("user_id", res.User.ID.String()), zap.String("ip", ip))

		render.Status(r, http.StatusOK)
		render.JSON(w, r, map[string]interface{}{
			"access_token": res.AccessToken,
			"user": map[string]interface{}{
				"id":       res.User.ID.String(),
				"username": res.User.Username,
				"email":    res.User.Email,
			},
		})
	}
}

func (ah *AuthHTTPHandler) Register(ctx context.Context) http.HandlerFunc {
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

		normalizedUser, err := ah.authService.Register(ctx, &regUser)
		if err != nil {
			log.Error("failed to register user", zap.Error(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, "failed to register user")

			return
		}

		log.Info("user registered successfully")

		cookie.SetCookie(w, normalizedUser.RefreshToken)

		render.JSON(w, r, models.AuthResponse{
			User:        normalizedUser.User,
			AccessToken: normalizedUser.AccessToken,
		})
	}
}

func (ah *AuthHTTPHandler) Logout(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const path = "server.http.handler.auth.Logout"

		log := ah.log.With(
			zap.String("path", path),
			zap.String("request-id", middleware.GetReqID(r.Context())),
		)

		refreshToken, err := r.Cookie("refresh_token")
		if err != nil {
			log.Error("failed to get refresh cookie", zap.Error(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, "failed to get refresh cookie")

			return
		}

		ctx := r.Context()
		err = ah.authService.Logout(ctx, refreshToken.Value)
		if err != nil {
			log.Error("failed to logout", zap.Error(err))

			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, "failed to logout")

			return
		}

		cookie.DeleteCookie(w)

		render.JSON(w, r, "logged out successfully")
	}
}

func (ah *AuthHTTPHandler) RefreshToken(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: RefreshToken
	}
}

func (ah *AuthHTTPHandler) VerifyInternal(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := ah.authService.VerifyAccessToken(r.Context(), token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("X-User-ID", claims.UserID)
		w.WriteHeader(http.StatusOK)
	}
}
