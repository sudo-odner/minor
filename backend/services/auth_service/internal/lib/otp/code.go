package otp

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateRandomCode(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	
	return base64.URLEncoding.EncodeToString(b)[:length], nil
}