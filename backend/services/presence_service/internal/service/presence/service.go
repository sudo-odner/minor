package presence

import (
	"go.uber.org/zap"
)

type Cache interface{}

type Broker interface{}

type Service struct {
	log    *zap.Logger
	cache  Cache
	broker Broker
}

func New(log *zap.Logger, cache Cache, broker Broker) *Service {
	return &Service{
		log:    log,
		cache:  cache,
		broker: broker,
	}
}
