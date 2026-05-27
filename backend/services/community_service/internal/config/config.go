package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env        string `env:"ENV" env-required:"true"`
	ServerHTTP ServerHTTP
	ServerGRPC ServerGRPC
	Broker     Nuts
	Repository Postgres
}

type ServerHTTP struct {
	// Required
	Address string `env:"HTTP_SERVER_ADDRESS" env-required:"true"`
	// Optional
	Timeout     time.Duration `env:"HTTP_SERVER_TIMEOUT" env-defalut:"5s"`
	IdleTimeout time.Duration `env:"HTTP_SERVER_IDLE_TIMEOUT" env-defalut:"20s"`
}

type ServerGRPC struct {
	// Required
	Address string `env:"GRPC_SERVER_ADDRESS" env-required:"true"`
	// Optional
	Timeout time.Duration `env:"GRPC_SERVER_TIMEOUT" env-required:"5s"`
}

type Nuts struct {
	// Required
	URL string `env:"NUTS_URL" env-required:"true"`
	// Optional
	Timeout          time.Duration `env:"NUTS_TIMEOUT" env-defalut:"10s"`
	MaxReconnects    int           `env:"NUTS_MAX_RECONNECTS" env-defalut:"5"`
	TimeoutReconnect time.Duration `env:"NUTS_TIMEOUT_RECONNECT" env-defalut:"2s"`
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

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("DEBUG: '.env' file not found, read from env")
	}
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("FATAL: cannot load config: %s", err)
	}
	return &cfg
}
