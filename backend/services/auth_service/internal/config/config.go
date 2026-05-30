package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	HTTP     HTTPConfig     `yaml:"htttp"`
	GRPC     GRPCConfig     `yaml:"grpc"`
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
	NATS     NATSConfig     `yaml:"nats"`
	Auth     AuthConfig     `yaml:"auth"`
}

type AppConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Env     string `yaml:"env"`
}

type HTTPConfig struct {
	Port           string        `yaml:"port"`
	Timeout        time.Duration `yaml:"timeout"`
	IdleTimeout    time.Duration `yaml:"idle_timeout"`
	AllowedOrigins []string      `yaml:"allowed_origins"`
}

type GRPCConfig struct {
	Port int `yaml:"port"`
}

type PostgresConfig struct {
	Host        string `yaml:"host"`
	Port        int `yaml:"port"`
	User        string `yaml:"user"`
	DBName      string `yaml:"db_name"`
	SSLMode     string `yaml:"sslmode"`
	MaxPoolSize string `yaml:"max_pool_size"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int `yaml:"port"`
	Password string `yaml:"password"`
	DB       string `yaml:"db"`
}

type NATSConfig struct {
	URL        string `yaml:"url"`
	StreamName string `yaml:"stream_name"`
}

type AuthConfig struct {
	JWTSecret       string        `yaml:"jwt_secret"`
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl"`
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading env variables: %s", err.Error())
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatalf("config path is not set")
	}

	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("config file does not exist: %s", err.Error())
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err.Error())
	}

	// cfg.Auth = AuthConfig{
	// 	JWTSecret:    []byte(getEnv("accessSecret", "default_access_secret")),
	// 	AccessTokenTTL:  parseDuration(getEnv("accessTokenDuration", "15m")),
	// 	RefreshTokenTTL: parseDuration(getEnv("refreshTokenDuration", "168h")),
	// }

	return &cfg
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Printf("invalid duration %s, using default 15m", s)
		return 15 * time.Minute
	}

	return d
}
