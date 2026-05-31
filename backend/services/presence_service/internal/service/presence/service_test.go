package presence_test

import (
	"context"
	"testing"

	"github.com/sudo-odner/minor/backend/services/presence_service/internal/models"
	"github.com/sudo-odner/minor/backend/services/presence_service/internal/service/presence"
	"go.uber.org/zap"
)

type mockCache struct {
	setStatusFn       func(ctx context.Context, userID string, status models.UserStatus, customStatus string, lastActiveAt int64) error
	getUserStatusesFn func(ctx context.Context, userIDs []string) (map[string]*models.Presence, error)
}

func (m *mockCache) SetStatus(ctx context.Context, userID string, status models.UserStatus, customStatus string, lastActiveAt int64) error {
	return m.setStatusFn(ctx, userID, status, customStatus, lastActiveAt)
}

func (m *mockCache) GetUserStatuses(ctx context.Context, userIDs []string) (map[string]*models.Presence, error) {
	return m.getUserStatusesFn(ctx, userIDs)
}

type mockBroker struct {
	publishFn func(ctx context.Context, p *models.Presence) error
}

func (m *mockBroker) PublishPresenceStatusUpdated(ctx context.Context, p *models.Presence) error {
	return m.publishFn(ctx, p)
}

func TestService_SetStatus(t *testing.T) {
	logger := zap.NewNop()

	var setStatusCalled bool
	var publishCalled bool

	mc := &mockCache{
		setStatusFn: func(ctx context.Context, userID string, status models.UserStatus, customStatus string, lastActiveAt int64) error {
			setStatusCalled = true
			if userID != "user-1" {
				t.Errorf("expected user-1, got %s", userID)
			}
			if status != models.UserStatusOnline {
				t.Errorf("expected status online, got %v", status)
			}
			return nil
		},
	}

	mb := &mockBroker{
		publishFn: func(ctx context.Context, p *models.Presence) error {
			publishCalled = true
			if p.UserID != "user-1" {
				t.Errorf("expected user-1, got %s", p.UserID)
			}
			return nil
		},
	}

	svc := presence.New(logger, mc, mb)

	err := svc.SetStatus(context.Background(), "user-1", models.UserStatusOnline, "Coding...")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !setStatusCalled {
		t.Error("expected SetStatus to be called on cache")
	}
	if !publishCalled {
		t.Error("expected PublishPresenceStatusUpdated to be called on broker")
	}
}

func TestService_GetUserStatuses(t *testing.T) {
	logger := zap.NewNop()

	expectedResult := map[string]*models.Presence{
		"user-1": {
			UserID:       "user-1",
			Status:       models.UserStatusOnline,
			CustomStatus: "Coding...",
			LastActiveAt: 123456,
		},
	}

	mc := &mockCache{
		getUserStatusesFn: func(ctx context.Context, userIDs []string) (map[string]*models.Presence, error) {
			return expectedResult, nil
		},
	}

	svc := presence.New(logger, mc, nil)

	res, err := svc.GetUserStatuses(context.Background(), []string{"user-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	p, ok := res["user-1"]
	if !ok {
		t.Fatal("expected user-1 in results")
	}

	if p.Status != models.UserStatusOnline {
		t.Errorf("expected online status, got %v", p.Status)
	}
}
