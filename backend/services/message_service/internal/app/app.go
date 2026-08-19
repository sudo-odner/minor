package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	httpServ "github.com/sudo-odner/minor/backend/services/message_service/internal/app/http"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/cache/redis"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/repository/cassandra"
	messagesHandler "github.com/sudo-odner/minor/backend/services/message_service/internal/server/http/handler/messages"
	messagesService "github.com/sudo-odner/minor/backend/services/message_service/internal/service/messages"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/transport/grpc/client/community"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/transport/grpc/client/relationship"
	appNats "github.com/sudo-odner/minor/backend/services/message_service/internal/transport/nats"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	log      *zap.Logger
	httpServ *httpServ.HttpServ

	connRelationship *grpc.ClientConn
	connCommunity    *grpc.ClientConn
	nats             *nats.Conn

	repo  *cassandra.Repository
	cache *redis.Cache
}

func New(cfg *config.Config, log *zap.Logger) (*App, error) {
	const op = "app.New"
	ctx := context.Background()
	a := &App{log: log}

	var success bool
	defer func() {
		if !success {
			log.Warn("startup falild, rolling back initialized resourse")
			_ = a.Stop(ctx)
		}
	}()

	// Init clients
	// Community gRPC client
	grpcConnCommunity, err := grpc.NewClient(cfg.GRPC.Client.TargetCommunity, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("%s: falied connect (gRPC) to community service: %w", op, err)
	}
	a.connCommunity = grpcConnCommunity
	clientCommunity := community.New(communityv1.NewCommunityServiceClient(grpcConnCommunity))

	// Relationship gRPC client
	grpcConnRelationship, err := grpc.NewClient(cfg.GRPC.Client.TargetRelationship, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("%s: falied connect (gRPC) to relationship service: %w", op, err)
	}
	a.connRelationship = grpcConnRelationship
	clientRelationship := relationship.New(relationshipv1.NewRelationshipServiceClient(grpcConnRelationship))

	// Init Broker (Nuts)
	nc, err := nats.Connect(
		cfg.Nats.URL,
		nats.Name("message_service"),
		nats.Timeout(cfg.Nats.Timeout),
		nats.MaxReconnects(cfg.Nats.MaxReconnects),
		nats.ReconnectWait(cfg.Nats.ReconnectWait),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to NATS Core:%w", op, err)
	}
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("s: failed to initilize JetStream: %w", op, err)
	}
	a.nats = nc
	brokerProducer := appNats.NewProducer(nc, js)
	brokerConsumer := appNats.NewConsumer(nc, js)

	// Init cache Redis
	cache, err := redis.New(ctx, cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("%s: cache (Redis) init failed: %w", op, err)
	}
	a.cache = cache

	// Init repository (Cassandra)
	repo, err := cassandra.New(&cfg.Cassandra)
	if err != nil {
		return nil, fmt.Errorf("%s: repository (Cassandra) init failed: %w", op, err)
	}
	a.repo = repo

	// Init service, handler & router
	service := messagesService.New(log, a.repo, a.broker, a.cache, a.clientCommunity, a.clientUser)
	messageHandler := messagesHandler.New(log, service)

	// TODO: add logger middlaware
	// TODO: move from App file

	router := chi.NewRouter()
	router.Route("/api/v1/messages", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		})

		r.Route("/{channel_id}", func(r chi.Router) {
			r.Post("/", messageHandler.SendMessage())
			r.Get("/", messageHandler.GetMessages())
			r.Get("/{message_id}", messageHandler.GetMessage())
			r.Delete("/{message_id}", messageHandler.DeleteMessage())
		})
	})

	a.httpServ = httpServ.New(&cfg.HttpServer, router)
	success = true
	return a, nil
}

func (a *App) Run() error {
	const op = "app.Run"
	a.log.Info("starting http server", zap.String("address", a.httpServ.Address()))

	if err := a.httpServ.Run(); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	const op = "app.Stop"
	log := a.log.With(zap.String("op", op))
	log.Info("stopping application")

	// Close server (http)
	if a.httpServ != nil {
		if err := a.httpServ.Stop(ctx); err != nil {
			log.Warn("failed to stop http server", zap.Error(err))
		}
	}

	// Close clients (gRPC)
	if a.clientCommunity != nil {
		if err := a.clientCommunity.Close(); err != nil {
			log.Warn("failed to close community gRPC client", zap.Error(err))
		}
	}
	if a.clientUser != nil {
		if err := a.clientUser.Close(); err != nil {
			log.Warn("failed to close user gRPC client", zap.Error(err))
		}
	}

	// Close broker
	if a.broker != nil {
		if err := a.broker.Stop(); err != nil {
			log.Warn("failed to stop broker", zap.Error(err))
		}
	}

	// Close cache
	if a.cache != nil {
		if err := a.cache.Stop(); err != nil {
			log.Warn("failed to stop cache", zap.Error(err))
		}
	}

	// Close repository
	if a.repo != nil {
		if err := a.repo.Close(); err != nil {
			log.Warn("failed to close repository", zap.Error(err))
		}
	}

	log.Info("application stopped successfully")
	return nil
}
