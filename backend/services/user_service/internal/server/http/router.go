package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/server/http/handler"
	"go.uber.org/zap"
)

type Handlers struct {
	User   *handler.UserHandler
	Friend *handler.FriendHandler
}

func NewRouter(log *zap.Logger, handlers Handlers) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// API Routes
	r.Route("/api/v1/users", func(r chi.Router) {
		r.Post("/", handlers.User.CreateUser())
		r.Get("/me", handlers.User.GetMe())
		r.Get("/{user_id}", handlers.User.GetUser())
		r.Patch("/", handlers.User.UpdateUser())
		r.Delete("/", handlers.User.DeleteUser())

		// Friends sub-routes
		r.Route("/friends", func(r chi.Router) {
			r.Get("/", handlers.Friend.FriendList())
			r.Get("/requests", handlers.Friend.FriendRequestList())
			r.Post("/requests/{friend_id}", handlers.Friend.SendFriendRequest())
			r.Patch("/requests/{friend_id}", handlers.Friend.AnswerFriendRequest())
			r.Delete("/{friend_id}", handlers.Friend.DeleteFriendship())
			r.Post("/block/{friend_id}", handlers.Friend.BlockUser())
		})
	})

	return r
}
