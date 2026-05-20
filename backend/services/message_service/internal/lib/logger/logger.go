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

func New(env Env) (*zap.Logger, error) {
	var zapCfg zap.Config

	switch env {
	case EnvLocal, EnvDev:
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case EnvProd:
		zapCfg = zap.NewProductionConfig()
		zapCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	default:
		return nil, fmt.Errorf("unknown env: %s", env)
	}

	logger, err := zapCfg.Build()
	if err != nil {
		return nil, err
	}

	return logger, nil
}
