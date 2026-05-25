package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	httpServ "github.com/sudo-odner/minor/backend/services/message_service/internal/app/http_serv"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/broker/nuts"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/cache/redis"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/client/grpc/community"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/client/grpc/user"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/repository/cassandra"
	messagesHandler "github.com/sudo-odner/minor/backend/services/message_service/internal/server/http/handler/messages"
	messagesService "github.com/sudo-odner/minor/backend/services/message_service/internal/service/messages"
	"go.uber.org/zap"
)

type App struct {
	log             *zap.Logger
	httpServ        *httpServ.HttpServ
	resourseToClose []func() error
}

func New(cfg *config.Config, log *zap.Logger) (*App, error) {
	const op = "app.New"

	// Массив для закрытия всех ресурсов (Resource Collector)
	var resourseToClose []func() error
	rollback := func() {
		for i := len(resourseToClose) - 1; i >= 0; i-- {
			if err := resourseToClose[i](); err != nil {
				log.Warn("falied close resource", zap.Error(err))
			}
		}
	}

	// Init repository Cassandra
	repo, err := cassandra.New(&cfg.Cassandra)
	if err != nil {
		return nil, fmt.Errorf("%s: repository(Cassandra) not init: %w", op, err)
	}
	resourseToClose = append(resourseToClose, repo.Close)

	// Init brocker Nuts
	brocker, err := nuts.New(cfg.Nuts)
	if err != nil {
		rollback()
		return nil, fmt.Errorf("%s: brocker(Nuts) not init: %w", op, err)
	}
	resourseToClose = append(resourseToClose, brocker.Stop)

	// Init cache Redis
	cache, err := redis.New(cfg.Resid)
	if err != nil {
		rollback()
		return nil, fmt.Errorf("%s: cache(Redis) not init: %w", op, err)
	}
	resourseToClose = append(resourseToClose, cache.Stop)

	// Init Community client gRPC
	communityClient, err := community.New(cfg.GRPC.Client.TargetCommunity)
	if err != nil {
		rollback()
		return nil, fmt.Errorf("%s: Community client(gRPC) not init: %w", op, err)
	}
	resourseToClose = append(resourseToClose, communityClient.Close)

	// Init User client gRPC
	userClient, err := user.New(cfg.GRPC.Client.TargetCommunity)
	if err != nil {
		rollback()
		return nil, fmt.Errorf("%s: User client(gRPC) not init: %w", op, err)
	}
	resourseToClose = append(resourseToClose, userClient.Close)

	// Init services
	service := messagesService.New(log, repo, brocker, cache, communityClient, userClient)

	// Init handler
	handler := messagesHandler.New(log, service)

	// TODO: add logger middlaware
	router := chi.NewRouter()
	router.Route("/", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})

		r.Route("/channels/{channel_id}/messages", func(r chi.Router) {
			r.Post("/", handler.SendMessage())
			r.Get("/", handler.GetMessages())
			r.Delete("/{message_id}", handler.DeleteMessage())
		})
	})

	return &App{
		log:             log,
		resourseToClose: resourseToClose,
		httpServ:        httpServ.New(&cfg.HttpServer, router),
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

	// Останавливаем HTTP-server
	if err := a.httpServ.Stop(ctx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("http server stopped")

	// Останавливаем все сервисы
	for i := len(a.resourseToClose) - 1; i >= 0; i-- {
		if err := a.resourseToClose[i](); err != nil {
			log.Warn("falied close resource", zap.Error(err))
		}
	}

	log.Info("application stopped successfully")
	return nil
}
