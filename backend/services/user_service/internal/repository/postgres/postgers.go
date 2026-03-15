package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Storage, error) {
	const op = "repository.postgres.New"
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("%s: ping failed: %w", op, err)
	}

	return &Storage{
		pool: pool,
	}, nil
}

func Close(ctx context.Context, storage *Storage) {
	storage.pool.Close()
}
