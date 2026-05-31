package logger

import (
	"fmt"

	"go.uber.org/zap"
)

type Env string

const (
	EnvLocal Env = "local"
	EnvDev   Env = "dev"
	EnvProd  Env = "prod"
)

type Config struct {
	Env         Env
	ServiceName string
}

func New(cfg Config) (*zap.Logger, error) {
	var zapCfg zap.Config

	switch cfg.Env {
	case EnvLocal, EnvDev:
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case EnvProd:
		zapCfg = zap.NewProductionConfig()
		zapCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	default:
		return nil, fmt.Errorf("unknown env: %s", cfg.Env)
	}

	logger, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}

	return logger.With(zap.String("service", cfg.ServiceName)), err
}
