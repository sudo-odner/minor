package auth

import "context"

type Authorization interface {
	Login(ctx context.Context, email string, password string) (err error)
	Register(ctx context.Context, email string, username string, password string) (err error)
	Logout(refresh_token)
	RefreshToken(refresh_token)
	VerifyEmail(code)
}