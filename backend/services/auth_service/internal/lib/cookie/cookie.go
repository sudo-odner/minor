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
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
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
