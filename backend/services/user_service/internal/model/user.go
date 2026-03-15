package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID
	Username string
	Bio      string
	CreateAt time.Time
	UpdateAt time.Time
}
