package postgres

import "context"

type Repository struct{}

func New() (*Repository, error) {
	return nil, nil
}

func (repo *Repository) Ping(ctx context.Context) error {
	return nil
}

func (repo *Repository) Close(ctx context.Context) error {
	return nil
}
