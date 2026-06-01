package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	// "github.com/joho/godotenv"
)

type Config struct {
	Env        string `env:"ENV" env-required:"true"`
	ServerGRPC ServerGRPC
	Redis      Redis
	Nuts       Nuts
}

type ServerGRPC struct {
	// Required
	Address string `env:"GRPC_SERVER_ADDRESS" env-required:"true"`
	// Optional
	Timeout time.Duration `env:"GRPC_SERVER_TIMEOUT" env-default:"5s"`
}

type Redis struct {
	// Required
	Url string `env:"REDIS_URL" env-required:"true"`
	// Optional
	PoolSize     int           `env:"REDIS_POOL_SIZE" env-default:"10"`
	MinIdleConns int           `env:"REDIS_MIN_IDLE_CONNS" env-default:"3"`
	DialTimeout  time.Duration `env:"REDIS_DIAL_TIMEOUT" env-default:"5s"`
	ReadTimeout  time.Duration `env:"REDIS_READ_TIMEOUT" env-default:"3s"`
	WriteTimeout time.Duration `env:"REDIS_WRITE_TIMEOUT" env-default:"3s"`
}

type Nuts struct {
	// Required
	URL string `env:"NUTS_URL" env-required:"true"`
	// Optional
	Timeout          time.Duration `env:"NUTS_TIMEOUT" env-default:"10s"`
	MaxReconnects    int           `env:"NUTS_MAX_RECONNECTS" env-default:"5"`
	TimeoutReconnect time.Duration `env:"NUTS_TIMEOUT_RECONNECT" env-default:"2s"`
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
