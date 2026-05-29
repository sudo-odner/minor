package events

import (
	"time"
)

type UserLoginSuccessEvent struct {
	UserID    string    `json:"user_id"` // UUID v7
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

type UserRegisteredEvent struct {
	UserID    string    `json:"user_id"` // UUID v7
	Email     string    `json:"email"`
	Timestamp time.Time `json:"timestamp"`
}

type UserLoggedOutEvent struct {
	UserID    string    `json:"user_id"`  // UUID v7
	TokenID   string    `json:"token_id"` // UUID рефреш-токена, который был удален
	Timestamp time.Time `json:"timestamp"`
}
