package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	// "github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig   
	GRPC     GRPCConfig     
	Postgres PostgresConfig 
	Redis    RedisConfig
	NATS     NATSConfig
	Auth     AuthConfig
}

type AppConfig struct {
	Name    string `env:"NAME"`
	Version string `env:"VERSION"`
	Env     string `env:"ENV"`
}

type HTTPConfig struct {
	Port           string        `env:"HTTP_PORT"`
	Timeout        time.Duration `env:"HTTP_TIMEOUT"`
	IdleTimeout    time.Duration `env:"HTTP_IDLE_TIMEOUT"`
	// AllowedOrigins []string      `env:"allowed_origins"`
}

type GRPCConfig struct {
	Port int `env:"GRPC_SERVER_PORT"`
}

type PostgresConfig struct {
	Host        string `env:"POSTGRES_HOST"`
	Port        int 	`env:"POSTGRES_PORT"`
	User        string `env:"POSTGRES_USER"`
	DBName      string `env:"POSTGRES_DB_NAME"`
	SSLMode     string `env:"POSTGRES_SSLMODE"`
	MaxPoolSize string `env:"POSTGRES_MAX_POOL_SIZE"`
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST"`
	Port     int `env:"REDIS_PORT"`
	Password string `env:"REDIS_PASSWORD"`
	DB       string `env:"REDIS_DB"`
}

type NATSConfig struct {
	URL        string `env:"NATS_URL"`
	StreamName string `env:"NATS_STREAM_NAME"`
}

type AuthConfig struct {
	JWTSecret       string        `env:"JWT_SECRET"`
	AccessTokenTTL  time.Duration `env:"ACCESS_TOKEN_TTL"`
	RefreshTokenTTL time.Duration `env:"REFRESH_TOKEN_TTL"`
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