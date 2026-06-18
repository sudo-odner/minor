package models

type ForgotPasswordPayload struct {
	Email string `json:"email"`
}

type ResetPasswordPayload struct {
	Email string `json:"email"`
	Token string `json:"token"`
	Password string `json:"password"`
}