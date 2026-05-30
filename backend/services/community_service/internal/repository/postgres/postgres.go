package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Repository, error) {
	const op = "repository.postgres.New"

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("%s: ping falied: %w", op, err)
	}

	return &Repository{
		pool: pool,
	}, nil
}

func (repo *Repository) Ping(ctx context.Context) error {
	if repo.pool != nil {
		if err := repo.pool.Ping(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (repo *Repository) Close(ctx context.Context) error {
	if repo.pool != nil {
		repo.pool.Close()
	}
	return nil
}
