package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpServ "github.com/sudo-odner/minor/backend/services/message_service/internal/app/http_serv"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/broker/nuts"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/cache/redis"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/client/grpc/guild"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/client/grpc/user"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/repository/cassandra"
	messagesHandler "github.com/sudo-odner/minor/backend/services/message_service/internal/server/http/handler/messages"
	messagesService "github.com/sudo-odner/minor/backend/services/message_service/internal/service/messages"
	"go.uber.org/zap"
)

type App struct {
	log         *zap.Logger
	httpServ    *httpServ.HttpServ
	repo        *cassandra.Repository
	broker      *nuts.Broker
	cache       *redis.Cache
	guildClient *guild.Clinet
	userClient  *user.Clinet
}

func New(cfg *config.Config, log *zap.Logger) (*App, error) {
	// repo cassandra
	repo, err := cassandra.New(&cfg.Cassandra)
	if err != nil {
		return nil, fmt.Errorf("cassandra not init")
	}
	// brocker nuts
	brocker, err := nuts.New(cfg.Nuts)
	if err != nil {
		repo.Close()
		return nil, fmt.Errorf("cassandra not init")
	}
	// chache redis
	cache, err := redis.New(cfg.Resid)
	if err != nil {
		repo.Close()
		_ = brocker.Stop()
		return nil, fmt.Errorf("redis not init")
	}

	// guild client
	guildClient, err := guild.New(cfg.grpcGuild.target)
	if err != nil {
		repo.Close()
		_ = brocker.Stop()
		_ = cache.Stop()
		return nil, fmt.Errorf("guild client not init")
	}

	// user client
	userClient, err := user.New(cfg.grpcUser.target)
	if err != nil {
		repo.Close()
		_ = brocker.Stop()
		_ = cache.Stop()
		return nil, fmt.Errorf("user client not init")
	}

	// services
	service := messagesService.New(log, repo, brocker, cache, guildClient, userClient)

	// handler
	handler := messagesHandler.New(log, service)

	// TODO: add logger middlaware

	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})

		r.Route("/channels/{channelID}/messages", func(r chi.Router) {
			r.Post("/", handler.SendMessage())
			r.Get("/", handler.GetMessages())
			r.Delete("/{messageID}", handler.DeleteMessage())
		})
	})

	return &App{
		log:         log,
		httpServ:    httpServ.New(&cfg.HttpServer, router),
		repo:        repo,
		broker:      brocker,
		cache:       cache,
		guildClient: guildClient,
		userClient:  userClient,
	}, nil
}

func (a *App) Run() error {
	const op = "app.Run"
	log := a.log.With(zap.String("op", op))

	log.Info("starting application")

	log.Info("starting http server", zap.String("address", a.httpServ.Address()))
	if err := a.httpServ.Run(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// TODO: Safe stop(repo, grpc and etc)
func (a *App) Stop(ctx context.Context) error {
	const op = "app.Stop"
	log := a.log.With(zap.String("op", op))

	log.Info("stopping application")
	if err := a.httpServ.Stop(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("http server stopped")

	log.Info("application stopped successfully")
	return nil
}
