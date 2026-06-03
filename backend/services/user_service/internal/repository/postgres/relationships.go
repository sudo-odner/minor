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

func (repo *Repository) SendFriendRequest(ctx context.Context, userID, friendID uuid.UUID) error {
	const op = "repository.postgres.SendFriendRequest"

	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: begin tx failed: %w", op, err)
	}
	defer tx.Rollback(ctx)

	// 1. Check actor's perspective
	var status models.RelationshipStatus
	queryActor := `
		SELECT status FROM relationships 
		WHERE user_id = $1 AND target_id = $2;
	`
	err = tx.QueryRow(ctx, queryActor, userID, friendID).Scan(&status)
	if err == nil {
		// Relation already exists
		if status == models.StatusFriends {
			return models.ErrAlreadyExists
		}
		if status == models.StatusRequestSent {
			return models.ErrAlreadyExists
		}
		if status == models.StatusBlocked {
			return models.ErrPermissionDenied
		}
		if status == models.StatusRequestReceived {
			// If target already sent a request, this action accepts it!
			queryAccept := `
				UPDATE relationships 
				SET status = 1, update_at = CURRENT_TIMESTAMP 
				WHERE (user_id = $1 AND target_id = $2) OR (user_id = $2 AND target_id = $1);
			`
			_, err = tx.Exec(ctx, queryAccept, userID, friendID)
			if err != nil {
				return fmt.Errorf("%s: auto accept failed: %w", op, err)
			}
			if err := tx.Commit(ctx); err != nil {
				return fmt.Errorf("%s: commit failed: %w", op, err)
			}
			return nil
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: check actor status failed: %w", op, err)
	}

	// 2. Check if target blocked actor
	var targetStatus models.RelationshipStatus
	queryTarget := `
		SELECT status FROM relationships 
		WHERE user_id = $1 AND target_id = $2;
	`
	err = tx.QueryRow(ctx, queryTarget, friendID, userID).Scan(&targetStatus)
	if err == nil {
		if targetStatus == models.StatusBlocked {
			return models.ErrPermissionDenied
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s: check target status failed: %w", op, err)
	}

	// 3. Create bidirectional relationship rows:
	//    Actor: StatusRequestSent (2)
	//    Target: StatusRequestReceived (3)
	queryInsert := `
		INSERT INTO relationships (user_id, target_id, status, create_at, update_at)
		VALUES 
			($1, $2, 2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
			($2, $1, 3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
	`
	_, err = tx.Exec(ctx, queryInsert, userID, friendID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return models.ErrNotFound
		}
		return fmt.Errorf("%s: insert relationship failed: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: commit failed: %w", op, err)
	}

	return nil
}

func (repo *Repository) FriendList(ctx context.Context, userID uuid.UUID) ([]*models.Relationship, error) {
	const op = "repository.postgres.FriendList"

	query := `
		SELECT user_id, target_id, status, create_at, update_at
		FROM relationships
		WHERE user_id = $1 AND status = 1;
	`
	rows, err := repo.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: query failed: %w", op, err)
	}
	defer rows.Close()

	var list []*models.Relationship
	for rows.Next() {
		var r models.Relationship
		if err := rows.Scan(&r.UserID, &r.TargetID, &r.Status, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("%s: scan failed: %w", op, err)
		}
		list = append(list, &r)
	}

	return list, nil
}

func (repo *Repository) RelationshipList(ctx context.Context, userID uuid.UUID) ([]*models.Relationship, error) {
	const op = "repository.postgres.RelationshipList"

	query := `
		SELECT user_id, target_id, status, create_at, update_at
		FROM relationships
		WHERE user_id = $1;
	`
	rows, err := repo.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: query failed: %w", op, err)
	}
	defer rows.Close()

	var list []*models.Relationship
	for rows.Next() {
		var r models.Relationship
		if err := rows.Scan(&r.UserID, &r.TargetID, &r.Status, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("%s: scan failed: %w", op, err)
		}
		list = append(list, &r)
	}

	return list, nil
}


func (repo *Repository) FriendRequestList(ctx context.Context, userID uuid.UUID) ([]*models.Relationship, error) {
	const op = "repository.postgres.FriendRequestList"

	query := `
		SELECT user_id, target_id, status, create_at, update_at
		FROM relationships
		WHERE user_id = $1 AND status = 3;
	`
	rows, err := repo.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: query failed: %w", op, err)
	}
	defer rows.Close()

	var list []*models.Relationship
	for rows.Next() {
		var r models.Relationship
		if err := rows.Scan(&r.UserID, &r.TargetID, &r.Status, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("%s: scan failed: %w", op, err)
		}
		list = append(list, &r)
	}

	return list, nil
}

func (repo *Repository) AcceptFriendRequest(ctx context.Context, actorID, targetID uuid.UUID) error {
	const op = "repository.postgres.AcceptFriendRequest"

	query := `
		UPDATE relationships 
		SET status = 1, update_at = CURRENT_TIMESTAMP 
		WHERE (user_id = $1 AND target_id = $2 AND status = 3) 
		   OR (user_id = $2 AND target_id = $1 AND status = 2);
	`
	res, err := repo.pool.Exec(ctx, query, actorID, targetID)
	if err != nil {
		return fmt.Errorf("%s: update failed: %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (repo *Repository) DenyFriendRequest(ctx context.Context, actorID, targetID uuid.UUID) error {
	const op = "repository.postgres.DenyFriendRequest"

	query := `
		DELETE FROM relationships 
		WHERE (user_id = $1 AND target_id = $2 AND status = 3) 
		   OR (user_id = $2 AND target_id = $1 AND status = 2);
	`
	res, err := repo.pool.Exec(ctx, query, actorID, targetID)
	if err != nil {
		return fmt.Errorf("%s: delete failed: %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (repo *Repository) BlockUser(ctx context.Context, actorID, targetID uuid.UUID) error {
	const op = "repository.postgres.BlockUser"

	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: begin tx failed: %w", op, err)
	}
	defer tx.Rollback(ctx)

	queryDelete := `
		DELETE FROM relationships 
		WHERE (user_id = $1 AND target_id = $2) OR (user_id = $2 AND target_id = $1);
	`
	_, err = tx.Exec(ctx, queryDelete, actorID, targetID)
	if err != nil {
		return fmt.Errorf("%s: delete old relationships failed: %w", op, err)
	}

	queryInsert := `
		INSERT INTO relationships (user_id, target_id, status, create_at, update_at)
		VALUES ($1, $2, 4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
	`
	_, err = tx.Exec(ctx, queryInsert, actorID, targetID)
	if err != nil {
		return fmt.Errorf("%s: insert block failed: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: commit failed: %w", op, err)
	}

	return nil
}

func (repo *Repository) RemoveFriend(ctx context.Context, actorID, targetID uuid.UUID) error {
	const op = "repository.postgres.RemoveFriend"

	tx, err := repo.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: begin tx failed: %w", op, err)
	}
	defer tx.Rollback(ctx)

	// 1. Get the relationship from actor's perspective
	var status models.RelationshipStatus
	querySelect := `
		SELECT status FROM relationships 
		WHERE user_id = $1 AND target_id = $2;
	`
	err = tx.QueryRow(ctx, querySelect, actorID, targetID).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ErrNotFound
		}
		return fmt.Errorf("%s: select status failed: %w", op, err)
	}

	// 2. Perform deletion based on status
	if status == models.StatusBlocked {
		// Blocker unblocks the target: delete only the block row
		queryDelete := `
			DELETE FROM relationships 
			WHERE user_id = $1 AND target_id = $2 AND status = 4;
		`
		res, err := tx.Exec(ctx, queryDelete, actorID, targetID)
		if err != nil {
			return fmt.Errorf("%s: delete block failed: %w", op, err)
		}
		if res.RowsAffected() == 0 {
			return models.ErrNotFound
		}
	} else {
		// Friends or pending request: delete both directional rows
		queryDelete := `
			DELETE FROM relationships 
			WHERE ((user_id = $1 AND target_id = $2) OR (user_id = $2 AND target_id = $1))
			  AND status IN (1, 2, 3);
		`
		res, err := tx.Exec(ctx, queryDelete, actorID, targetID)
		if err != nil {
			return fmt.Errorf("%s: delete friendship/request failed: %w", op, err)
		}
		if res.RowsAffected() == 0 {
			return models.ErrNotFound
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: commit failed: %w", op, err)
	}
	return nil
}
