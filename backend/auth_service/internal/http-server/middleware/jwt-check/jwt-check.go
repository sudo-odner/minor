package jwtcheck

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/sudo-odner/minor/backend/auth_service/internal/config"
	"github.com/sudo-odner/minor/backend/auth_service/internal/lib/api/response"
	"github.com/sudo-odner/minor/backend/auth_service/internal/lib/jwt"
)


func AuthMiddleware(cfg config.TokenConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			claims, err := jwt.ValidateAccessToken(cfg, token)
			if err != nil {
				render.Status(r, http.StatusUnauthorized)

				render.JSON(w, r, response.ErrorResponse{
					Status: http.StatusUnauthorized,
					Error: "user unauthorized",
				})

				return
			}

			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}