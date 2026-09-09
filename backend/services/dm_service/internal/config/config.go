package config

import (
	"fmt"
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `env:"ENV" env-required:"true"`
	HTTPServer HTTPServer
	Nats       Nats
	Postgres   Postgres
}

type HTTPServer struct {
	Address     string        `env:"HTTP_SERVER_ADDRESS" env-required:"true"`
	Timeout     time.Duration `env:"HTTP_SERVER_TIMEOUT"`
	IdleTimeout time.Duration `env:"HTTP_SERVER_IDLE_TIMEOUT"`
}

type Nats struct {
	// REQUIRED
	URL string `env:"NATS_URL" env-required:"true"`
	// OPTIONAL
	Timeout       time.Duration `env:"NATS_TIMEOUT" env-default:"10s"`
	MaxReconnects int           `env:"NATS_MAX_RECONNECTS" env-default:"5"`
	ReconnectWait time.Duration `env:"NATS_RECONNECT_WAIT" env-default:"2s"`
}

type Postgres struct {
	// Required
	Host     string `env:"POSTGRES_HOST" env-required:"true"`
	Port     int    `env:"POSTGRES_PORT" env-required:"true"`
	User     string `env:"POSTGRES_USER" env-required:"true"`
	Password string `env:"POSTGRES_PASSWORD" env-required:"true"`
	Database string `env:"POSTGRES_DATABASE" env-required:"true"`
	// Optional
}

func (p *Postgres) ConnString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		p.User, p.Password,
		p.Host, p.Port,
		p.Database,
	)
}

func MustLoad() *Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("ERROR: cannot read config: %s", err)
	}

	return &cfg
}
