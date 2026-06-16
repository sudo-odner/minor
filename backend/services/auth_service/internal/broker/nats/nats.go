// package nats

// import (
// 	"time"

// 	"github.com/nats-io/nats.go"
// 	"github.com/nats-io/nats.go/jetstream"
// 	"github.com/sudo-odner/minor/backend/services/auth_service/internal/config"
// 	"go.uber.org/zap"
// )

// type App struct {
// 	log        *zap.Logger
// 	natsConn *nats.Conn
// 	JS jetstream.JetStream
// }

// func New(log *zap.Logger, cfg *config.Config) *App {
// 	var nc *nats.Conn
// 	var err error
// 	natsURL := cfg.NATS.URL

// 	for i := 0; i < 10; i++ {
// 		nc, err = nats.Connect(natsURL)
// 		if err == nil {
// 			break
// 		}
// 		log.Info("NATS not ready yet (attempt %d): %v", zap.Int("attempt", i+1), zap.String("error", err.Error()))
// 		time.Sleep(time.Second * 2)
// 	}

// 	if err != nil || nc == nil {
// 		log.Error("Fatal: could not connect to NATS: %v", zap.Error(err))
// 	}

// 	js, err := jetstream.New(nc)
// 	if err != nil {
// 		log.Error("Fatal: could not initialize JetStream: %v", zap.Error(err))
// 	}

// 	return &App{
// 		log: log,
// 		natsConn: nc,
// 		JS: js,
// 	}
// }

package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/config"
	"go.uber.org/zap"
)

type App struct {
	Conn *nats.Conn
	JS   jetstream.JetStream
	log  *zap.Logger
}

func New(log *zap.Logger, cfg *config.Config) (*App, error) {
	// 1. Настройка опций подключения с логикой Retry
	opts := []nats.Option{
		nats.Name("auth_service"),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2 * time.Second),
	}

	// 2. Подключение
	nc, err := nats.Connect(cfg.NATS.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}

	// 3. Инициализация JetStream
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("jetstream init: %w", err)
	}

	return &App{
		Conn: nc,
		JS:   js,
		log:  log,
	}, nil
}

// InitAuthStream создает поток для всех событий авторизации.
// ВАЖНО: Subjects должны покрывать все темы, которые вы используете.
func (a *App) InitAuthStream(ctx context.Context) error {
	streamName := "AUTH_STREAM"
	
	cfg := jetstream.StreamConfig{
		Name:     streamName,
		// Указываем маски тем, которые этот поток будет сохранять.
		// auth.user.> покроет регистрацию и логин
		// auth.password.> покроет запросы на сброс пароля
		Subjects: []string{"auth.user.>", "auth.password.>"}, 
		Storage:  jetstream.FileStorage, // Храним на диске
		Retention: jetstream.LimitsPolicy,
		MaxAge:   7 * 24 * time.Hour,    // Храним события неделю
	}

	_, err := a.JS.CreateOrUpdateStream(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create/update stream %s: %w", streamName, err)
	}

	a.log.Info("NATS JetStream initialized", zap.String("stream", streamName))
	return nil
}

func (a *App) Stop() {
	a.log.Info("closing nats connection")
	a.Conn.Close()
}