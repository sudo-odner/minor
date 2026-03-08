package postgres

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
	mu sync.Mutex
}

func New(ctx context.Context, storagePath string) (*Storage, error) {
	const op = "repository.postgres.New"

	pool, err := pgxpool.New(ctx, storagePath) 
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{
		pool: pool,
		mu: sync.Mutex{},
	}, nil
}

func Close(ctx context.Context, storage *Storage) {
	storage.pool.Close()
}