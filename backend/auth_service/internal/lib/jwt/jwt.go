package jwt

import (
	"errors"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var (
	accessDuration = os.Getenv("accessDuration")
	refreshDuration = os.Getenv("refreshDuration")
)

var (
	errUserUnauthorized = errors.New("user unauthorized")
)

type Claims struct {
	UserID int `json:"user_id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}