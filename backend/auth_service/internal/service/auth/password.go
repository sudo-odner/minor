package auth

type Password interface {
	ForgotPassword(ctx, email)
	ResetPassword(ctx, code, newPass)
}