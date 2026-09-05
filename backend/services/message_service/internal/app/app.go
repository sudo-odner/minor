package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

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
	httptransport "github.com/sudo-odner/minor/backend/services/message_service/internal/transport/http"
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

	grpcCommunity, grpcDM, err := a.initGRPCClient(ctx, &cfg.GRPC)
	if err != nil {
		return nil, err
	}
	_, producer, err := a.initNats(ctx, &cfg.Nats)
	if err != nil {
		return nil, err
	}
	cache, err := a.initRedis(ctx, &cfg.Redis)
	if err != nil {
		return nil, err
	}
	repo, err := a.initCassandra(ctx, &cfg.Cassandra)
	if err != nil {
		return nil, err
	}

	// Init service, handler & router
	service := messagesrv.New(log, repo, producer, cache, grpcCommunity, grpcDM)
	messageHandler := messagehandler.New(log, service)

	// Setup HTTP Server
	a.httpServer = &http.Server{
		Addr: cfg.HTTPServer.Address,
		Handler: httptransport.NewRouter(httptransport.Handlers{
			Message: messageHandler,
		}),
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

func (a *App) initGRPCClient(ctx context.Context, cfg *config.GRPC) (*communityclient.Client, *dbclient.Client, error) {
	const op = "app.initGRPCClient"

	// Community gRPC client
	grpcCommunity, err := grpc.NewClient(
		cfg.Client.TargetCommunity,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: falied connect (gRPC) to community service: %w", op, err)
	}
	a.grpcCommunity = grpcCommunity

	// DM gRPC client
	grpcDM, err := grpc.NewClient(cfg.Client.TargetDM, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("%s: falied connect (gRPC) to dm service: %w", op, err)
	}
	a.grpcDM = grpcDM

	return communityclient.New(communityv1.NewCommunityServiceClient(grpcCommunity)), dbclient.New(dmv1.NewDMServiceClient(grpcDM)), nil
}

func (a *App) initNats(ctx context.Context, cfg *config.Nats) (*natstransport.Consumer, *natstransport.Producer, error) {
	const op = "app.initNats"

	nc, err := nats.Connect(
		cfg.URL,
		nats.Name("message_service"),
		nats.Timeout(cfg.Timeout),
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.ReconnectWait(cfg.ReconnectWait),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: failed to connect to NATS Core:%w", op, err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: failed to initilize JetStream: %w", op, err)
	}

	a.nats = nc
	return natstransport.NewConsumer(nc, js), natstransport.NewProducer(nc, js), nil
}

func (a *App) initRedis(ctx context.Context, cfg *config.Redis) (*redisrepo.Cache, error) {
	const op = "app.initRedis"

	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse redis url: %w", op, err)
	}
	opts.PoolSize = cfg.PoolSize         // Максимальное колличество соединений на сервис
	opts.MinIdleConns = cfg.MinIdleConns // Минимальное значения откртых соединений (горячий старт)
	opts.DialTimeout = cfg.DialTimeout
	opts.ReadTimeout = cfg.ReadTimeout
	opts.WriteTimeout = cfg.WriteTimeout

	a.redis = redis.NewClient(opts)
	return redisrepo.New(a.redis), nil
}

func (a *App) initCassandra(ctx context.Context, cfg *config.Cassandra) (*cassandrarepo.Repository, error) {
	const op = "app.initCassandra"

	cluster := gocql.NewCluster(cfg.Host)
	cluster.Keyspace = cfg.Keyspace
	if cfg.Username != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}

	cluster.Timeout = cfg.Timeout
	consistency, err := gocql.ParseConsistencyWrapper(cfg.Consistency)
	if err != nil {
		return nil, fmt.Errorf("%s: unncorect consistency type %w", op, err)
	}
	cluster.Consistency = consistency
	cluster.NumConns = cfg.NumConns

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create session: %w", op, err)
	}

	a.cassandra = session
	return cassandrarepo.New(session), nil
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
