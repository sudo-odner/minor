package models

type ForgotPasswordPayload struct {
	Email string `json:"email"`
}

type ResetPasswordPayload struct {
	Email string `json:"email"`
	Code string `json:"code"`
	NewPassword string `json:"new_password"`
}