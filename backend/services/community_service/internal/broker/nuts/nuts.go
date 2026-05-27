package nuts

import "context"

type Broker struct{}

func New() (*Broker, error) {
	return nil, nil
}

func (b *Broker) Ping(ctx context.Context) error {
	return nil
}

func (b *Broker) Close(ctx context.Context) error {
	return nil
}
