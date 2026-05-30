package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Email     string
	Username  string
	AvatarURL *string
	Bio       string
	CreateAt  time.Time
	UpdateAt  time.Time
}
