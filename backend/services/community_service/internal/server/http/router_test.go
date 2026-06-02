package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/server/http/handler"
	"go.uber.org/zap"
)

type mockServerService struct {
	handler.ServerService
}
func (m *mockServerService) CreateServer(ctx context.Context, name string, ownerID uuid.UUID, avatarURL string) (*models.Server, error) {
	return &models.Server{}, nil
}
func (m *mockServerService) GetUserServers(ctx context.Context, userID uuid.UUID) ([]models.Server, error) {
	return []models.Server{}, nil
}
func (m *mockServerService) GetServer(ctx context.Context, serverID uuid.UUID) (*models.Server, error) {
	return &models.Server{}, nil
}

type mockChannelService struct {
	handler.ChannelService
}
func (m *mockChannelService) GetServerChannel(ctx context.Context, serverID uuid.UUID) ([]models.Channel, error) {
	return []models.Channel{}, nil
}

func TestRouter_POST_Servers(t *testing.T) {
	log := zap.NewNop()
	// Mock handlers
	handlers := Handlers{
		Server:  *handler.NewServerHandler(log, &mockServerService{}),
		Channel: *handler.NewChannelHandler(log, &mockChannelService{}),
		Member:  *handler.NewMemberHandler(log, nil),
		Role:    *handler.NewRoleHandler(log, nil),
	}

	router := NewRouter(log, handlers)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "POST /api/v1/servers/ (with slash)",
			method:         "POST",
			path:           "/api/v1/servers/",
			expectedStatus: http.StatusUnauthorized, // Unauthorized because of ParseUserID in handler
		},
		{
			name:           "POST /api/v1/servers (no slash)",
			method:         "POST",
			path:           "/api/v1/servers",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "POST /api/v1/servers/123/channels (no slash)",
			method:         "POST",
			path:           "/api/v1/servers/69c8167e-61c0-40e8-8a42-9f3792036c64/channels",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "GET /api/v1/servers/123/channels (no slash)",
			method:         "GET",
			path:           "/api/v1/servers/69c8167e-61c0-40e8-8a42-9f3792036c64/channels",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tt.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}
