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
		return nil, fmt.Errorf("%s: failed select channel by id: %w", op, err)
	}

	return &channel, nil
}

func (r *ChannelRepository) ByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Channel, error) {
	const op = "repository.postgres.ByUserID"

	query := `
		select 
			id, type, name, updated_at, created_at
		from members_channel mc
		join channels c on c.id = mc.channel_id
		where mc.user_id = $1
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to query user channels: %w", op, err)
	}

	var channels []domain.Channel
	for rows.Next() {
		var ch domain.Channel
		if err := rows.Scan(&ch.ID, &ch.Type, &ch.Name, &ch.UpdatedAt, &ch.CreatedAt); err != nil {
			return nil, fmt.Errorf("%s: scan channel error: %w", op, err)
		}
		channels = append(channels, ch)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: rows iteration error: %w", op, err)
	}

	return channels, nil
}

func (r *ChannelRepository) Membres(ctx context.Context, channelID uuid.UUID) ([]uuid.UUID, error) {
	const op = "repository.postgres.Members"

	query := `
		select user_id
		from members_channel where channel_id = $1
	`
	rows, err := r.pool.Query(ctx, query, channelID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to qyery members channel: %w", op, err)
	}
	userIDs, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	if err != nil {
		return nil, fmt.Errorf("%s: failed collect members: %w", op, err)
	}

	return userIDs, err
}

// UserPermission if user in channel his can READ, WRITE, ATTACH FILES and KICK MEMBERS
func (r *ChannelRepository) UserPermission(ctx context.Context, channelID, userID uuid.UUID) (authz.Permission, error) {
	const op = "repository.psotgres.UserPermission"

	query := `
		select 1
		from members_channel
		where user_id = $1 and channel_id = $2 
	`
	var l int
	if err := r.pool.QueryRow(ctx, query, userID, channelID).Scan(&l); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return authz.None, nil
		}
		return authz.None, fmt.Errorf("%s: failed query row: %w", op, err)
	}

	return authz.PermViewChannel | authz.PermSendMessages | authz.PermAttachFiles | authz.PermKickMembers, nil
}

func (r *ChannelRepository) Delete(channelID, userID uuid.UUID) error { return nil }
