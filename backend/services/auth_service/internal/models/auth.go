package models

import "github.com/google/uuid"

type User struct {
	Email string
	PasswordHash string
}

type LoginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterUser struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Password string    `json:"password"`
}

type NormalizedUser struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
}
