package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	// "github.com/joho/godotenv"
)

type Config struct {
	App     AppConfig
	HTTP 	HTTPConfig
	Nats    NatsConfig    
	GRPC GRPCConfig 
	SMTP    SMTPConfig    
}

type AppConfig struct {
	// Name string `yaml:"name"`
	Env  string `env:"ENV"`
}

type HTTPConfig struct {
	Port string `env:"HTTP_PORT"`
	Timeout time.Duration `env:"HTTP_TIMEOUT"`
	IdleTimeout time.Duration `env:"HTTP_IDLE_TIMEOUT"`
}

type NatsConfig struct {
	URL         string `env:"NATS_URL"`
	Stream      string `env:"NATS_STREAM"`
	// Subject     string `env:"subject"`
	// DurableName string `env:"durable_name"`
}

type GRPCConfig struct {
	PresenceService struct {
		Address string        `env:"PRESENCE_GRPC_ADDR"`
	} 
	UserService struct {
		Address string        `env:"USER_GRPC_ADDR"`
	} 
}

type SMTPConfig struct {
	Host     string `env:"SMTP_HOST"`
	Port     int    `env:"SMTP_PORT"`
	User     string `env:"SMTP_USER"`
	Password string `env:"SMTP_PASS"`
	From     string `env:"SMTP_FROM"`
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
