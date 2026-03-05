package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env           string        `yaml:"env"`
	ClientDomain  string        `yaml:"client_domain"`
	PostgreConfig `yaml:"psql"`
	ServerConfig  `yaml:"http_server"`
	TokenConfig
}

type PostgreConfig struct {
	Username string `yaml:"username"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type TokenConfig struct {
	AccessSecret    []byte
	RefreshSecret   []byte
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type ServerConfig struct {
	Port            string        `yaml:"port"`
	Timeout         time.Duration `yaml:"timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout"`
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

	cfg.TokenConfig = TokenConfig{
		AccessSecret: []byte(getEnv("accessSecret", "default_access_secret")),
		RefreshSecret: []byte(getEnv("refreshSecret", "default_refresh_secret")),
		AccessTokenTTL: parseDuration(getEnv("accessTokenDuration", "15m")),
		RefreshTokenTTL: parseDuration(getEnv("refreshTokenDuration", "168h")),
	}

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