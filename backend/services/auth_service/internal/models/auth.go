package models

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID
	Email        string
	Username     string
	IsActive     bool
	PasswordHash string
}

type LoginUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterUser struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type NormalizedUser struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Username string    `json:"username"`
}

type AuthResponse struct {
	User         *NormalizedUser `json:"user"`
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
}
