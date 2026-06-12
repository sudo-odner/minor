package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/models"
)

func (s *Storage) Create(ctx context.Context, input *models.User) error {
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
		INSERT INTO credentials(id, email, username, password_hash, is_active, created_at)
		VALUES($1, $2, $3, $4, true, CURRENT_TIMESTAMP)
		RETURNING id;
	`, input.ID, input.Email, input.Username, input.PasswordHash)

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
		SELECT c.id, c.email, c.username, c.password_hash
		FROM credentials c
		WHERE c.email = $1;
	`, email)

	var user models.User
	err = row.Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) GetByID(ctx context.Context, id string) (*models.User, error) {
	const op = "repository.postgres.auth.GetByID"
	
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
			err = fmt.Errorf("%s: %w", op, commitErr)
		}
	}()
	
	row := s.pool.QueryRow(ctx, `
		SELECT c.id, c.email, c.username, c.password_hash
		FROM credentials c
		WHERE c.id = $1;
	`, id)
	
	var user models.User
	
	err = row.Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) UpdatePassword(ctx context.Context, id string, newPasswordHash string) error {
	const op = "repository.postgres.auth.UpdatePassword"
	
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
			err = fmt.Errorf("%s: %w", op, commitErr)
		}
	}()
	
	_, err = s.pool.Exec(ctx, `
		UPDATE credentials c
		SET c.password_hash = $1
		WHERE c.id = $2;
	`, newPasswordHash, id)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}