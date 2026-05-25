package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env        string `env:"ENV" env-required:"true"`
	HttpServer HttpServer
	Cassandra  Cassandra
	Nuts       Nuts
	Resid      Redis
	GRPC       GRPC
}

type HttpServer struct {
	Address     string        `env:"HTTP_SERVER_ADDRESS" env-required:"true"`
	Timeout     time.Duration `env:"HTTP_SERVER_TIMEOUT"`
	IdleTimeout time.Duration `env:"HTTP_SERVER_IDLE_TIMEOUT"`
}

type Cassandra struct {
	Host        string        `env:"CASSANDRA_HOST" env-required:"true"`
	Keyspace    string        `env:"CASSANDRA_KEYSPACE" env-required:"true"`
	Username    string        `env:"CASSANDRA_USERNAME" env-dafault:""`
	Password    string        `env:"CASSANSRA_PASSWORD" env-default:""`
	Consistency string        `env:"CASSANDRA_CONSISTENCY" env-default:"ONE"`
	NumConns    int           `env:"CASSANDRA_NUM_CONNS" env-default:"4"`
	Timeout     time.Duration `env:"CASSANDRA_TIMEOUT" env-default:"15s"`
}

type Redis struct {
	// REQUIRED
	Url string `env:"REDIS_URL" env-required:"true"`
	// OPTIONAL
	PoolSize     int           `env:"REDIS_POOL_SIZE" env-default:"10"`
	MinIdleConns int           `env:"REDIS_MIN_IDLE_CONNS" env-default:"3"`
	DialTimeout  time.Duration `env:"REDIS_DIAL_TIMEOUT" env-default:"5s"`
	ReadTimeout  time.Duration `env:"REDIS_READ_TIMEOUT" env-default:"3s"`
	WriteTimeout time.Duration `env:"REIDS_WRITE_TIMEOUT" env-default:"3s"`
}

type Nuts struct {
	Url           string        `env:"NATS_URL" env-required:"true"`
	Timeout       time.Duration `env:"NATS_TIMEOUT" env-default:"10s"`
	MaxReconnects int           `env:"NATS_MAX_RECONNECTS" env-default:"5"`
	ReconnectWait time.Duration `env:"NATS_RECONNECT_WAIT" env-default:"2s"`
}

type GRPC struct {
	Client GRPCClient
}

type GRPCClient struct {
	TargetUser      string `env:"GRPC_CLIENT_USER_TARGET" env-required:"true"`
	TargetCommunity string `env:"GRPC_CLIENT_COMMUNITY_TARGET" env-required:"true"`
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("DEBUG: not found .env file, read form env")
	}

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("ERROR: cannot read config: %s", err)
	}

	return &cfg
}
