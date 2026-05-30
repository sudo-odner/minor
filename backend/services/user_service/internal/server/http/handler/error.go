package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/models"
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

func RenderModelError(w http.ResponseWriter, r *http.Request, err error, defaultMsg string) {
	if errors.Is(err, models.ErrPermissionDenied) {
		RenderError(w, r, http.StatusForbidden, CodeForbidden, "forbidden")
		return
	}
	if errors.Is(err, models.ErrNotFound) {
		RenderError(w, r, http.StatusNotFound, CodeNotFound, err.Error())
		return
	}
	if errors.Is(err, models.ErrImpossible) {
		RenderError(w, r, http.StatusBadRequest, CodeInvalidRequest, err.Error())
		return
	}
	if errors.Is(err, models.ErrAlreadyExists) {
		RenderError(w, r, http.StatusConflict, CodeInvalidRequest, err.Error())
		return
	}
	RenderError(w, r, http.StatusInternalServerError, CodeInternalServerError, defaultMsg)
}
