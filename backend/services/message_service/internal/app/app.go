package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gocql/gocql"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/redis/go-redis/v9"
	communityv1 "github.com/sudo-odner/minor-shared/pkg/pb/community/v1"
	dmv1 "github.com/sudo-odner/minor-shared/pkg/pb/dm/v1"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/config"
	cassandrarepo "github.com/sudo-odner/minor/backend/services/message_service/internal/repository/cassandra"
	redisrepo "github.com/sudo-odner/minor/backend/services/message_service/internal/repository/redis"
	messagesrv "github.com/sudo-odner/minor/backend/services/message_service/internal/service/messages"
	communityclient "github.com/sudo-odner/minor/backend/services/message_service/internal/transport/grpc/client/community"
	dbclient "github.com/sudo-odner/minor/backend/services/message_service/internal/transport/grpc/client/dm"
	messagehandler "github.com/sudo-odner/minor/backend/services/message_service/internal/transport/http/handler/messages"
	natstransport "github.com/sudo-odner/minor/backend/services/message_service/internal/transport/nats"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	log *zap.Logger

	httpServer *http.Server

	grpcDM        *grpc.ClientConn
	grpcCommunity *grpc.ClientConn

	nats      *nats.Conn
	redis     *redis.Client
	cassandra *gocql.Session
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

	// Init GRPC Client
	// Community gRPC client
	grpcCommunityConn, err := grpc.NewClient(
		cfg.GRPC.Client.TargetCommunity,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: falied connect (gRPC) to community service: %w", op, err)
	}
	a.grpcCommunity = grpcCommunityConn

	// DM gRPC client
	grpcDMConn, err := grpc.NewClient(cfg.GRPC.Client.TargetDM, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("%s: falied connect (gRPC) to dm service: %w", op, err)
	}
	a.grpcDM = grpcDMConn
	grpcCommunity, grpcDM := communityclient.New(communityv1.NewCommunityServiceClient(grpcCommunityConn)), dbclient.New(dmv1.NewDMServiceClient(grpcDMConn))

	// Init NATS
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
		return nil, fmt.Errorf("%s: failed to initilize JetStream: %w", op, err)
	}

	a.nats = nc
	_, producer := natstransport.NewConsumer(nc, js), natstransport.NewProducer(nc, js)

	// Init Redis
	opts, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse redis url: %w", op, err)
	}
	opts.PoolSize = cfg.Redis.PoolSize         // Максимальное колличество соединений на сервис
	opts.MinIdleConns = cfg.Redis.MinIdleConns // Минимальное значения откртых соединений (горячий старт)
	opts.DialTimeout = cfg.Redis.DialTimeout
	opts.ReadTimeout = cfg.Redis.ReadTimeout
	opts.WriteTimeout = cfg.Redis.WriteTimeout

	a.redis = redis.NewClient(opts)
	cache := redisrepo.New(a.redis)

	// Init Cassandra
	cluster := gocql.NewCluster(cfg.Cassandra.Host)
	cluster.Keyspace = cfg.Cassandra.Keyspace
	if cfg.Cassandra.Username != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.Cassandra.Username,
			Password: cfg.Cassandra.Password,
		}
	}

	cluster.Timeout = cfg.Cassandra.Timeout
	consistency, err := gocql.ParseConsistencyWrapper(cfg.Cassandra.Consistency)
	if err != nil {
		return nil, fmt.Errorf("%s: unncorect consistency type %w", op, err)
	}
	cluster.Consistency = consistency
	cluster.NumConns = cfg.Cassandra.NumConns

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create session: %w", op, err)
	}

	a.cassandra = session
	repo := cassandrarepo.New(session)

	// Init service, handler & router
	service := messagesrv.New(log, repo, producer, cache, grpcCommunity, grpcDM)
	messageHandler := messagehandler.New(log, service)

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

	// Setup HTTP Server
	a.httpServer = &http.Server{
		Addr:        cfg.HTTPServer.Address,
		Handler:     router,
		ReadTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout: cfg.HTTPServer.IdleTimeout,
	}

	success = true
	return a, nil
}

func (a *App) Run() error {
	const op = "app.Run"
	a.log.Info("starting http server", zap.String("address", a.httpServer.Addr))

	if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	const op = "app.Stop"
	log := a.log.With(zap.String("op", op))
	log.Info("stopping application")

	// Close server (http)
	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			log.Warn("failed to stop http server", zap.Error(err))
		}
	}

	// Close clients (gRPC)
	if a.grpcCommunity != nil {
		if err := a.grpcCommunity.Close(); err != nil {
			log.Warn("failed to close community gRPC client", zap.Error(err))
		}
	}
	if a.grpcDM != nil {
		if err := a.grpcDM.Close(); err != nil {
			log.Warn("failed to close user gRPC client", zap.Error(err))
		}
	}

	// Close nats
	if a.nats != nil {
		a.nats.Close()
	}

	// Close redis
	if a.redis != nil {
		if err := a.redis.Close(); err != nil {
			log.Warn("failed to stop cache", zap.Error(err))
		}
	}

	// Close cassandra
	if a.cassandra != nil {
		a.cassandra.Close()
	}

	log.Info("application stopped successfully")
	return nil
}
