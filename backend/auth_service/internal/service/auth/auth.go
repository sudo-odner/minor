package auth

import "context"

type AuthRepository interface {
	Login(ctx context.Context, email string, password string) (Tokens, error)
	Register(ctx context.Context, input SignUpInput) error
	Logout(refresh_token)
	VerifyEmail(ctx context.Context, code string) error
}

type AuthService struct {
	
}



