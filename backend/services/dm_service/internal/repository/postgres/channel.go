package channel

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	"github.com/sudo-odner/minor-shared/pkg/channelid"
	"github.com/sudo-odner/minor/backend/service/dm_service/internal/domain"
)

type ChannelRepository struct {
	pool *pgxpool.Pool
}

func NewChannelRepository(pool *pgxpool.Pool) *ChannelRepository {
	return &ChannelRepository{
		pool: pool,
	}
}

func (r *ChannelRepository) Create(ctx context.Context, channelType domain.ChannelType, name *string, userIDs []uuid.UUID) (*domain.Channel, error) {
	const op = "repository.postgres.Create"

	id, err := channelid.New(channelid.TypeDM)
	now := time.Now().UTC()
	if err != nil {
		return nil, fmt.Errorf("%s: failed generate uuid: %w", op, err)
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: failed begin tracsation: %w", op, err)
	}
	defer tx.Rollback(ctx)

	channelQuery := `
		insert into channels (id, type, name, updated_at, created_at)
		values ($1, $2, $3, $4, $5)
	`
	if _, err := tx.Exec(ctx, channelQuery, id, channelType, name, now, now); err != nil {
		return nil, fmt.Errorf("%s: failed insert channel: %w", op, err)
	}

	var rows [][]any
	for _, userID := range userIDs {
		rows = append(rows, []any{userID, id, now})
	}
	if _, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"members_channel"},
		[]string{"user_id", "channel_id", "created_at"},
		pgx.CopyFromRows(rows)); err != nil {
		return nil, fmt.Errorf("%s: failed copy insert members channel: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: failed commit tx: %w", op, err)
	}

	return &domain.Channel{
		ID:        id,
		Type:      channelType,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (r *ChannelRepository) ByID(ctx context.Context, channelID uuid.UUID) (*domain.Channel, error) {
	const op = "repository.postgres.ByID"

	var channel domain.Channel

	query := `
		select
			id, type, name, updated_at, created_at
		from channels where id = $1;
	`
	if err := r.pool.QueryRow(ctx, query, channelID).Scan(&channel.ID, &channel.Type, &channel.Name, &channel.UpdatedAt, &channel.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s (channel id=%s): %w", op, channelID.String(), domain.ErrNotFound)
		}
		return nil, fmt.Errorf("%s: falied select channel by id: %w", op, err)
	}

	return &channel, nil
}

func (r *ChannelRepository) ByUserID(userID uuid.UUID) ([]domain.Channel, error) { return nil, nil }

func (r *ChannelRepository) Membres(channelID uuid.UUID) ([]uuid.UUID, error) { return nil, nil }

func (r *ChannelRepository) UserPermission(channelID, userID uuid.UUID) authz.Permission {
	return 0x0000000000000000
}

func (r *ChannelRepository) Delete(channelID, userID uuid.UUID) error { return nil }
