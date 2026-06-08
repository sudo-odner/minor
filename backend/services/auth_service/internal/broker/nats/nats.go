package nats

import (
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/config"
	"go.uber.org/zap"
)

type App struct {
	log        *zap.Logger
	natsConn *nats.Conn
	JS jetstream.JetStream
}

func New(log *zap.Logger, cfg *config.Config) *App {
	var nc *nats.Conn
	var err error
	natsURL := cfg.NATS.URL

	for i := 0; i < 10; i++ {
		nc, err = nats.Connect(natsURL)
		if err == nil {
			break
		}
		log.Info("NATS not ready yet (attempt %d): %v", zap.Int("attempt", i+1), zap.String("error", err.Error()))
		time.Sleep(time.Second * 2)
	}

	if err != nil || nc == nil {
		log.Error("Fatal: could not connect to NATS: %v", zap.Error(err))
	}

	js, err := jetstream.New(nc)
	if err != nil {
		log.Error("Fatal: could not initialize JetStream: %v", zap.Error(err))
	}

	return &App{
		log: log,
		natsConn: nc,
		JS: js,
	}
}
