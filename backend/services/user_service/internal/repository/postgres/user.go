package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/model"
)

func (repo *Repository) CreateUser(ctx context.Context, u *model.User) (*model.User, error) {
	const op = "repository.postgres.CreateUser"

	if u.Email == "" {
		u.Email = fmt.Sprintf("%s@example.com", u.Username)
	}

	query := `
		INSERT INTO users (id, username, email, avatar_url, bio, create_at, update_at)
		VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING create_at, update_at;
	`
	err := repo.pool.QueryRow(
		ctx,
		query,
		u.ID,
		u.Username,
		u.Email,
		u.AvatarURL,
		u.Bio,
	).Scan(&u.CreateAt, &u.UpdateAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, model.ErrAlreadyExists
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return u, nil
}

func (repo *Repository) GetUser(ctx context.Context, id uuid.UUID) (*model.User, error) {
	const op = "repository.postgres.GetUser"

	query := `
		SELECT id, email, username, avatar_url, bio, create_at, update_at
		FROM users
		WHERE id = $1;
	`
	var u model.User
	err := repo.pool.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.AvatarURL,
		&u.Bio,
		&u.CreateAt,
		&u.UpdateAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &u, nil
}

func (repo *Repository) UpdateUser(ctx context.Context, id uuid.UUID, username *string, bio *string) (*model.User, error) {
	const op = "repository.postgres.UpdateUser"

	query := `
		UPDATE users
		SET
			username = COALESCE($1, username),
			bio = COALESCE($2, bio),
			update_at = CURRENT_TIMESTAMP
		WHERE id = $3
		RETURNING id, email, username, avatar_url, bio, create_at, update_at;
	`
	var u model.User
	err := repo.pool.QueryRow(ctx, query, username, bio, id).Scan(
		&u.ID,
		&u.Email,
		&u.Username,
		&u.AvatarURL,
		&u.Bio,
		&u.CreateAt,
		&u.UpdateAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, model.ErrAlreadyExists
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &u, nil
}

func (repo *Repository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	const op = "repository.postgres.DeleteUser"

	query := `DELETE FROM users WHERE id = $1;`
	res, err := repo.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}
