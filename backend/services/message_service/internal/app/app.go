package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gocql/gocql"
	gonats "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	goredis "github.com/redis/go-redis/v9"
	communityv1 "github.com/sudo-odner/minor-shared/pkg/pb/community/v1"
	relationshipv1 "github.com/sudo-odner/minor-shared/pkg/pb/relationship/v1"
	httpServ "github.com/sudo-odner/minor/backend/services/message_service/internal/app/http"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/repository/cassandra"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/repository/redis"
	messagesHandler "github.com/sudo-odner/minor/backend/services/message_service/internal/server/http/handler/messages"
	messagesService "github.com/sudo-odner/minor/backend/services/message_service/internal/service/messages"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/transport/grpc/client/community"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/transport/grpc/client/relationship"
	"github.com/sudo-odner/minor/backend/services/message_service/internal/transport/nats"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	log      *zap.Logger
	httpServ *httpServ.HttpServ

	grpcRelationship *grpc.ClientConn
	grpcCommunity    *grpc.ClientConn

	nats      *gonats.Conn
	redis     *goredis.Client
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

	grpcCommunity, grpcRelationship, err := a.initGRPCClient(ctx, &cfg.GRPC)
	if err != nil {
		return nil, err
	}
	consumer, producer, err := a.initNats(ctx, &cfg.Nats)
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
	service := messagesService.New(log, repo, producer, cache, grpcCommunity, grpcRelationship)
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

func (a *App) initGRPCClient(ctx context.Context, cfg *config.GRPC) (*community.Client, *relationship.Client, error) {
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

	// Relationship gRPC client
	grpcRelationship, err := grpc.NewClient(cfg.Client.TargetRelationship, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("%s: falied connect (gRPC) to relationship service: %w", op, err)
	}
	a.grpcRelationship = grpcRelationship

	return community.New(communityv1.NewCommunityServiceClient(grpcCommunity)), relationship.New(relationshipv1.NewRelationshipServiceClient(grpcRelationship)), nil
}

func (a *App) initNats(ctx context.Context, cfg *config.Nats) (*nats.Consumer, *nats.Producer, error) {
	const op = "app.initNats"

	nc, err := gonats.Connect(
		cfg.URL,
		gonats.Name("message_service"),
		gonats.Timeout(cfg.Timeout),
		gonats.MaxReconnects(cfg.MaxReconnects),
		gonats.ReconnectWait(cfg.ReconnectWait),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: failed to connect to NATS Core:%w", op, err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, nil, fmt.Errorf("s: failed to initilize JetStream: %w", op, err)
	}

	a.nats = nc
	return nats.NewConsumer(nc, js), nats.NewProducer(nc, js), nil
}

func (a *App) initRedis(ctx context.Context, cfg *config.Redis) (*redis.Cache, error) {
	const op = "app.initRedis"

	opts, err := goredis.ParseURL(cfg.Url)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse redis url: %w", op, err)
	}
	opts.PoolSize = cfg.PoolSize         // Максимальное колличество соединений на сервис
	opts.MinIdleConns = cfg.MinIdleConns // Минимальное значения откртых соединений (горячий старт)
	opts.DialTimeout = cfg.DialTimeout
	opts.ReadTimeout = cfg.ReadTimeout
	opts.WriteTimeout = cfg.WriteTimeout

	a.redis = goredis.NewClient(opts)
	return redis.New(a.redis), nil
}

func (a *App) initCassandra(ctx context.Context, cfg *config.Cassandra) (*cassandra.Repository, error) {
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
	return cassandra.New(session), nil
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
	if a.grpcCommunity != nil {
		if err := a.grpcCommunity.Close(); err != nil {
			log.Warn("failed to close community gRPC client", zap.Error(err))
		}
	}
	if a.grpcRelationship != nil {
		if err := a.grpcRelationship.Close(); err != nil {
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
