package handler

import (
	"net/http"

	"github.com/go-chi/render"
)

type ErrorCode string

const (
	CodeInvalidRequset      = "INVALID_REQUEST"
	CodeNotFound            = "NOT_FOUND"
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
)

type ErrorDitail struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

type ErrorResponse struct {
	ErrorDitail `json:"error"`
}

func RenderError(w http.ResponseWriter, r *http.Request, status int, errorCode ErrorCode, message string) {
	render.Status(r, status)
	render.JSON(w, r, ErrorResponse{
		ErrorDitail: ErrorDitail{
			Code:    errorCode,
			Message: message,
		},
	})
}
