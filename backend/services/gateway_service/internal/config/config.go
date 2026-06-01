package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	// "github.com/joho/godotenv"
)

type Config struct {
	App   AppConfig
	HTTP  HTTPConfig
	GRPC  GRPCConfig
	NATS  NATSConfig
}

type AppConfig struct {
	// Name    string `env:"name"`
	// Version string `env:"version"`
	Env string `env:"ENV"`
}

type HTTPConfig struct {
	Port string `env:"HTTP_PORT"`
	Timeout time.Duration `env:"HTTP_TIMEOUT"`
	IdleTimeout time.Duration `env:"HTTP_IDLE_TIMEOUT"`
}

type GRPCConfig struct {
	AuthService struct {
		Address string        `env:"AUTH_GRPC_ADDR"`
	}
	PresenceService struct {
		Address string        `env:"PRESENCE_GRPC_ADDR"`
	}
}

type NATSConfig struct {
	URL        string `env:"NATS_URL"`
	StreamName string `env:"NATS_STREAM_NAME"`
}

func MustLoad() *Config {
	// if err := godotenv.Load(); err != nil {
	// 	log.Println("DEBUG: not found .env file, read form env")
	// }

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("ERROR: cannot read config: %s", err)
	}

	return &cfg
}
