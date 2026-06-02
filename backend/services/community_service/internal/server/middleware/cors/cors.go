package cors

import (
	"net/http"

	"github.com/rs/cors"
)


func NewCORS(h http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		AllowCredentials: true,
		MaxAge: 120,
	})

	h = c.Handler(h)

	return h
}