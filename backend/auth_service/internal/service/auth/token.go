package auth

type Token interface {
	GenerateTokens(user)
	RefreshToken(ctx, refreshToken)
	ValidateToken(accessToken)
}