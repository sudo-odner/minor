package http

import (
	"net/http"

	"github.com/go-chi/chi"
)

type MessageHandler interface {
	SendMessage() http.HandlerFunc
	GetMessage() http.HandlerFunc
	DeleteMessage() http.HandlerFunc
}

type Handlers struct {
	Message MessageHandler
}

func NewRouter(handlers Handlers) http.Handler {
	// TODO: add logger middlaware
	// TODO: move from App file

	router := chi.NewRouter()
	router.Route("/api/v1/messages", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})

		r.Route("/{channel_id}", func(r chi.Router) {
			r.Post("/", handlers.Message.SendMessage())
			r.Get("/", handlers.Message.GetMessage())
			r.Get("/{message_id}", handlers.Message.GetMessage())
			r.Delete("/{message_id}", handlers.Message.DeleteMessage())
		})
	})
	return router
}
