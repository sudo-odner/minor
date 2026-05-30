package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/server/http/handler"
	"go.uber.org/zap"
)

type Handlers struct {
	Server  *handler.ServerHandler
	Channel *handler.ChannelHandler
	Member  *handler.MemberHandler
	Role    *handler.RoleHandler
}

func NewRouter(log *zap.Logger, handlers Handlers) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// API Routes
	r.Route("/api/v1", func(r chi.Router) {
		// Server endpoints
		r.Route("/servers", func(r chi.Router) {
			r.Post("/", handlers.Server.CreateServer())
			r.Get("/", handlers.Server.GetUserServers())
			r.Route("/{server_id}", func(r chi.Router) {
				r.Get("/", handlers.Server.GetServer())
				r.Patch("/", handlers.Server.UpdateServer())
				r.Delete("/", handlers.Server.DeleteServer())

				// Channels nested in Server
				r.Route("/channels", func(r chi.Router) {
					r.Post("/", handlers.Channel.CreateChannel())
					r.Get("/", handlers.Channel.GetServerChannels())
					r.Route("/{channel_id}", func(r chi.Router) {
						r.Patch("/", handlers.Channel.UpdateChannel())
						r.Delete("/", handlers.Channel.DeleteChannel())
						r.Post("/move", handlers.Channel.MoveChannel())
					})
				})

				// Members nested in Server
				r.Route("/members", func(r chi.Router) {
					r.Post("/", handlers.Member.AddMember())
					r.Get("/", handlers.Member.GetServerMembers())
					r.Route("/{user_id}", func(r chi.Router) {
						r.Get("/", handlers.Member.GetServerMember())
						r.Delete("/", handlers.Member.RemoveMember())
						r.Patch("/nickname", handlers.Member.UpdateNickname())
						r.Route("/roles/{role_id}", func(r chi.Router) {
							r.Patch("/", handlers.Member.AddRoleToMember())
							r.Delete("/", handlers.Member.RemoveRoleFromMember())
						})
					})
				})

				// Roles nested in Server
				r.Route("/roles", func(r chi.Router) {
					r.Post("/", handlers.Role.CreateRole())
					r.Get("/", handlers.Role.GetServerRoles())
					r.Route("/{role_id}", func(r chi.Router) {
						r.Patch("/", handlers.Role.UpdateRole())
						r.Delete("/", handlers.Role.DeleteRole())
					})
				})

				// Channel permission overrides nested in Server
				r.Put("/channels/{channel_id}/overrides", handlers.Role.ReplaceChannelPermissionOverrides())
			})
		})
	})

	return r
}
