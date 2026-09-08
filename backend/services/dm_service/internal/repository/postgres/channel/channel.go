package channel

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	"github.com/sudo-odner/minor/backend/service/dm_service/internal/domain"
)

type ChannelRepository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *ChannelRepository {
	return &ChannelRepository{
		pool: pool,
	}
}

func (r *ChannelRepository) Create(channelType domain.ChannelType, name *string, userIDs []uuid.UUID) (*domain.Channel, error) {
	return nil, nil
}

func (r *ChannelRepository) ByID(channelID uuid.UUID) (*domain.Channel, error) { return nil, nil }

func (r *ChannelRepository) ByUserID(userID uuid.UUID) ([]domain.Channel, error) { return nil, nil }

func (r *ChannelRepository) Membres(channelID uuid.UUID) ([]uuid.UUID, error) { return nil, nil }

func (r *ChannelRepository) UserPermission(channelID, userID uuid.UUID) authz.Permission {
	return 0x0000000000000000
}

func (r *ChannelRepository) Delete(channelID, userID uuid.UUID) error { return nil }
