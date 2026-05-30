package handler

import (
	"net/http"

	"github.com/go-chi/render"
)

type ErrorCode string

const (
	CodeInvalidRequest      ErrorCode = "INVALID_REQUEST"
	CodeNotFound            ErrorCode = "NOT_FOUND"
	CodeForbidden           ErrorCode = "FORBIDDEN"
	CodeUnauthorized        ErrorCode = "UNAUTHORIZED"
	CodeInternalServerError ErrorCode = "INTERNAL_SERVER_ERROR"
)

type ErrorDetail struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

func RenderError(w http.ResponseWriter, r *http.Request, status int, errorCode ErrorCode, message string) {
	render.Status(r, status)
	render.JSON(w, r, ErrorResponse{
		Error: ErrorDetail{
			Code:    errorCode,
			Message: message,
		},
	})
}
