package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func ParseUserID(r *http.Request) (uuid.UUID, error) {
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		return uuid.Nil, fmt.Errorf("missing X-User-ID header")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid X-User-ID format: %w", err)
	}
	return userID, nil
}

func ParseUUIDParam(r *http.Request, paramName string) (uuid.UUID, error) {
	val := chi.URLParam(r, paramName)
	if val == "" {
		return uuid.Nil, fmt.Errorf("missing parameter: %s", paramName)
	}
	id, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid uuid format for parameter %s: %w", paramName, err)
	}
	return id, nil
}
