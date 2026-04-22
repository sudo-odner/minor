package cassandra

import (
	"context"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/chat_service/internal/models"
)

func (r *Repository) SaveMessage(ctx context.Context, msg *models.Message) error { return nil }

func (r *Repository) GetMessages(ctx context.Context, channelID uuid.UUID, limit int, beforeID uuid.UUID) ([]models.Message, error) {
	return nil, nil
}
