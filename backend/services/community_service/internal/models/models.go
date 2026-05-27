package models

import (
	"errors"
)

var (
	ErrNotFound         = errors.New("resource not found")
	ErrAlreadyExists    = errors.New("resource already exists")
	ErrPermissionDenied = errors.New("permission denied")
)
