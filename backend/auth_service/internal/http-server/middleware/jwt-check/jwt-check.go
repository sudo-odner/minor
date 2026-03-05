package jwtcheck

import (
	"net/http"

	"github.com/gin-gonic/gin/render"
	"github.com/go-chi/render"
)


func JWTCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := cookie.CheckCookie(w, r)
		if err != nil || token == nil {
			render.Status(r, http.StatusUnauthorized)

			render.JSON(w, r, resp.ErrorResponse{
				Status: http.StatusUnauthorized,
				Error: "user unauthorized",
			})

			return
		}

		next.ServeHTTP(w, r)
	})
}