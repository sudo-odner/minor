package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sudo-odner/minor/backend/services/user_service/internal/models"
)

func (repo *Repository) CreateUser(ctx context.Context, u *models.User) (*models.User, error) {
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
			return nil, models.ErrAlreadyExists
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return u, nil
}

func (repo *Repository) GetUser(ctx context.Context, id uuid.UUID) (*models.User, error) {
	const op = "repository.postgres.GetUser"

	query := `
		SELECT id, email, username, avatar_url, bio, create_at, update_at
		FROM users
		WHERE id = $1;
	`
	var u models.User
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
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &u, nil
}

func (repo *Repository) UpdateUser(ctx context.Context, id uuid.UUID, username *string, bio *string) (*models.User, error) {
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
	var u models.User
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
			return nil, models.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, models.ErrAlreadyExists
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
		return models.ErrNotFound
	}
	return nil
}

func (repo *Repository) GetUsernameByID(ctx context.Context, userID string) (string, error) {
	query := `SELECT username FROM users WHERE id = $1`
	
	var username string
	err := repo.pool.QueryRow(ctx, query, userID).Scan(&username)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("user not found")
		}
		return "", err
	}
	
	return username, nil
}

func (repo *Repository) GetEmailByID(ctx context.Context, userID string) (string, error) {
	query := `SELECT email FROM users WHERE id = $1`
	
	var email string
	err := repo.pool.QueryRow(ctx, query, userID).Scan(&email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("user not found")
		}
		return "", err
	}
	
	return email, nil
}