package permissions_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor-shared/pkg/authz"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/models"
	"github.com/sudo-odner/minor/backend/services/community_service/internal/service/permissions"
	"go.uber.org/zap"
)

type mockRepo struct {
	server           *models.Server
	serverErr        error
	channel          *models.Channel
	channelErr       error
	role             *models.Role
	roleErr          error
	memberRoles      []models.Role
	memberRolesErr   error
	channelOverrides []models.ChannelPermissionOverride
	overridesErr     error
}

func (m *mockRepo) GetServer(ctx context.Context, id uuid.UUID) (*models.Server, error) {
	return m.server, m.serverErr
}

func (m *mockRepo) GetChannel(ctx context.Context, id uuid.UUID) (*models.Channel, error) {
	return m.channel, m.channelErr
}

func (m *mockRepo) GetRole(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	return m.role, m.roleErr
}

func (m *mockRepo) GetMemberRoles(ctx context.Context, serverID, userID uuid.UUID) ([]models.Role, error) {
	return m.memberRoles, m.memberRolesErr
}

func (m *mockRepo) GetChannelOverrides(ctx context.Context, id uuid.UUID) ([]models.ChannelPermissionOverride, error) {
	return m.channelOverrides, m.overridesErr
}

func TestFetchPermissions_OwnerHasAllPermissions(t *testing.T) {
	userID := uuid.New()
	serverID := uuid.New()
	channelID := uuid.New()

	repo := &mockRepo{
		channel: &models.Channel{ID: channelID, ServerID: serverID},
		server:  &models.Server{ID: serverID, OwnerID: userID},
	}

	logger := zap.NewNop()
	svc := permissions.New(logger, repo)

	perms, err := svc.FetchPermissions(context.Background(), userID, channelID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := authz.Permission(0xFFFFFFFFFFFFFFFF)
	if perms != expected {
		t.Errorf("expected permissions %b, got %b", expected, perms)
	}
}

func TestFetchPermissions_CombineRolesAndApplyOverrides(t *testing.T) {
	userID := uuid.New()
	serverID := uuid.New()
	channelID := uuid.New()

	role1 := models.Role{ID: uuid.New(), Permission: authz.Permission(0x01)} // e.g. ViewChannel
	role2 := models.Role{ID: uuid.New(), Permission: authz.Permission(0x02)} // e.g. SendMessages

	repo := &mockRepo{
		channel: &models.Channel{ID: channelID, ServerID: serverID},
		server:  &models.Server{ID: serverID, OwnerID: uuid.New()}, // Different owner
		role:    &models.Role{ID: serverID, Permission: authz.Permission(0x00)}, // everyone role
		memberRoles: []models.Role{role1, role2},
		channelOverrides: []models.ChannelPermissionOverride{
			{
				ChannelID:  channelID,
				TargetType: models.OverrideTypeUser,
				TargetID:   userID,
				Allow:      authz.Permission(0x04), // e.g. EmbedLinks
				Deny:       authz.Permission(0x02), // deny SendMessages
			},
		},
	}

	logger := zap.NewNop()
	svc := permissions.New(logger, repo)

	perms, err := svc.FetchPermissions(context.Background(), userID, channelID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Initial roles combined: 0x01 | 0x02 = 0x03
	// User override applied: (0x03 & ~0x02) | 0x04 = 0x01 | 0x04 = 0x05
	expected := authz.Permission(0x05)
	if perms != expected {
		t.Errorf("expected permissions %v, got %v", expected, perms)
	}
}

func TestFetchPermissions_GetChannelError(t *testing.T) {
	repo := &mockRepo{
		channelErr: errors.New("database connection failed"),
	}

	logger := zap.NewNop()
	svc := permissions.New(logger, repo)

	_, err := svc.FetchPermissions(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
