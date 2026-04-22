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

type Nuts struct {
	Url string `env:"NATS_URL" env-required:"true"`
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
