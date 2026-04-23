package cookie

import (
	"net/http"
	"time"
)

func SetCookie(w http.ResponseWriter, refreshToken string) {
	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/secure/refresh",
		HttpOnly: true,
		Secure:   false,
	}
	http.SetCookie(w, refreshCookie)
}

func DeleteCookie(w http.ResponseWriter) {
	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/secure/refresh",
		HttpOnly: true,
		Secure:   false,
		Expires:  time.Now(),
	}
	http.SetCookie(w, refreshCookie)
}
