package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/presence_service/internal/models"
)

func (s *Storage) Create(ctx context.Context, input models.User) error {
	const op = "repository.postgres.auth.Create"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}

		commitErr := tx.Commit(ctx)
		if commitErr != nil {
			err = fmt.Errorf("%s: %w", op, err)
		}
	}()

	row := tx.QueryRow(ctx, `
		INSERT INTO credentials(id, email, password_hash, is_active, created_at)
		VALUES($1, $2, $3, TRUE, CURRENT_TIMESTAMP())
		RETURNING id;
	`, input.ID, input.Email, input.PasswordHash)

	var returnID uuid.UUID

	err = row.Scan(&returnID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "repository.postgres.auth.GetByEmail"

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}

		commitErr := tx.Commit(ctx)
		if commitErr != nil {
			err = fmt.Errorf("%s: %w", op, err)
		}
	}()

	row := tx.QueryRow(ctx, `
		SELECT c.id, c.email, c.password_hash
		FROM credentials c
		WHERE c.email = $1;
	`, email)

	var user models.User
	err = row.Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	const op = "repository.postgres.auth.GetByID"

	var user models.User

	return &user, nil
}

// func (s *Storage) SetEmailVerified(id User) {

// }
