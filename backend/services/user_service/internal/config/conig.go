package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env           string        `yaml:"env" env-required:"true"`
	PostgreConfig PostgreConfig `yaml:"psql"`
	ServerConfig  ServerConfig  `yaml:"http_server"`
}

type PostgreConfig struct {
	Host     string `yaml:"host"     env:"-"           env-required:"true"`
	Port     string `yaml:"port"     env:"-"           env-required:"true"`
	Username string `yaml:"username" env:"-"           env-required:"true"`
	DBname   string `yaml:"dbname"   env:"-"           env-required:"true"`
	SSLmode  string `yaml:"sslmode"  env:"-"           env-default:"enable"`
	Password string `yaml:"-"        env:"PG_PASSWORD" env-required:"true"`
}

type ServerConfig struct {
	Host        string        `yaml:"host"         env-required:"true"`
	Port        string        `yaml:"port"         env-required:"true"`
	Timeout     time.Duration `yaml:"timeout"      env-default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func (c *PostgreConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.DBname,
		c.SSLmode,
	)
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("fatal load .env file: %s", err.Error())
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("config path is not set")
	}

	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("cannot read config path: %s", err.Error())
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err.Error())
	}

	return &cfg
}
