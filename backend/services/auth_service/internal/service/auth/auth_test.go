package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/models"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/service/auth"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type mockAuthRepo struct {
	users map[string]*models.User
}

func (m *mockAuthRepo) Create(ctx context.Context, newUser *models.User) error {
	m.users[newUser.Email] = newUser
	return nil
}

func (m *mockAuthRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	u, ok := m.users[email]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (m *mockAuthRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	for _, u := range m.users {
		if u.ID.String() == id {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *mockAuthRepo) UpdatePassword(ctx context.Context, id string, newPasswordHash string) error {
	if user := m.users[id]; user == nil {
		return errors.New("user not found")
	}
	m.users[id].PasswordHash = newPasswordHash

	return nil
}

type mockSessionRepo struct {
	tokens map[string]string
}

func (m *mockSessionRepo) SetRefreshToken(ctx context.Context, userID string, tokenID string, ttl time.Duration) error {
	m.tokens[tokenID] = userID
	return nil
}

func (m *mockSessionRepo) GetUserIDByRefreshToken(ctx context.Context, tokenID string) (string, error) {
	id, ok := m.tokens[tokenID]
	if !ok {
		return "", errors.New("session not found")
	}
	return id, nil
}

func (m *mockSessionRepo) DeleteRefreshToken(ctx context.Context, tokenID string) error {
	delete(m.tokens, tokenID)
	return nil
}

func (m *mockSessionRepo) DeleteAllUserSessions(ctx context.Context, userID string) error {
	for k, v := range m.tokens {
		if v == userID {
			delete(m.tokens, k)
		}
	}
	return nil
}

type mockResetRepo struct {
	codes map[string]string
}

func (m *mockResetRepo) SetResetCode(ctx context.Context, id string, code string, timeout time.Duration) error {
	m.codes[id] = code
	return nil
}

func (m *mockResetRepo) GetResetCode(ctx context.Context, id string) (string, error) {
	code := m.codes[id]
	return code, nil
}

func (m *mockResetRepo) DeleteResetCode(ctx context.Context, id string) error {
	delete(m.codes, id)
	return nil
}

type mockPublisher struct {
	registered []*models.User
	logins     []string
}

func (m *mockPublisher) PublishUserRegistered(ctx context.Context, user *models.User) error {
	m.registered = append(m.registered, user)
	return nil
}

func (m *mockPublisher) PublishLoginSuccess(ctx context.Context, userID, ip, userAgent string) error {
	m.logins = append(m.logins, userID)
	return nil
}

func (m *mockPublisher) PublishUserLoggedOut(ctx context.Context, userID, tokenID string) error {
	return nil
}

func (m *mockPublisher) PublishPasswordResetRequested(ctx context.Context, userID, code string, username string) error {
	return nil
}

func TestAuthService_Register_Success(t *testing.T) {
	repo := &mockAuthRepo{users: make(map[string]*models.User)}
	session := &mockSessionRepo{tokens: make(map[string]string)}
	reset := &mockResetRepo{codes: make(map[string]string)}
	pub := &mockPublisher{}

	cfg := config.AuthConfig{
		JWTSecret:      "secret_key_secret_key_secret_key",
		AccessTokenTTL: 5 * time.Minute,
	}

	svc := auth.New(repo, session, reset, pub, zap.NewNop(), cfg)

	newUser := &models.RegisterUser{
		Email:    "test@novsu.ru",
		Username: "testuser",
		Password: "password123",
	}

	resp, err := svc.Register(context.Background(), newUser)
	if err != nil {
		t.Fatalf("unexpected error during registration: %v", err)
	}

	if resp.User.Email != newUser.Email {
		t.Errorf("expected email %s, got %s", newUser.Email, resp.User.Email)
	}

	// Verify that password hash is saved in DB and password matches bcrypt hash
	savedUser, err := repo.GetByEmail(context.Background(), newUser.Email)
	if err != nil {
		t.Fatalf("failed to retrieve registered user: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(savedUser.PasswordHash), []byte(newUser.Password))
	if err != nil {
		t.Errorf("password verification failed: %v", err)
	}

	// Verify that event was published
	if len(pub.registered) != 1 {
		t.Errorf("expected 1 registered event, got %d", len(pub.registered))
	}
}

func TestAuthService_Register_AlreadyExists(t *testing.T) {
	repo := &mockAuthRepo{users: make(map[string]*models.User)}
	session := &mockSessionRepo{tokens: make(map[string]string)}
	reset := &mockResetRepo{codes: make(map[string]string)}
	pub := &mockPublisher{}

	cfg := config.AuthConfig{
		JWTSecret:      "secret_key_secret_key_secret_key",
		AccessTokenTTL: 5 * time.Minute,
	}

	// Pre-create user in database
	userID, _ := uuid.NewV7()
	repo.users["test@novsu.ru"] = &models.User{
		ID:           userID,
		Email:        "test@novsu.ru",
		Username:     "existing",
		PasswordHash: "some_hash",
	}

	svc := auth.New(repo, session, reset, pub, zap.NewNop(), cfg)

	newUser := &models.RegisterUser{
		Email:    "test@novsu.ru",
		Username: "newname",
		Password: "password123",
	}

	_, err := svc.Register(context.Background(), newUser)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, auth.ErrUserAlreadyExists) {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := &mockAuthRepo{users: make(map[string]*models.User)}
	session := &mockSessionRepo{tokens: make(map[string]string)}
	reset := &mockResetRepo{codes: make(map[string]string)}
	pub := &mockPublisher{}

	cfg := config.AuthConfig{
		JWTSecret:      "secret_key_secret_key_secret_key",
		AccessTokenTTL: 5 * time.Minute,
	}

	password := "mypassword"
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	userID, _ := uuid.NewV7()

	// Pre-register user
	repo.users["user@novsu.ru"] = &models.User{
		ID:           userID,
		Email:        "user@novsu.ru",
		Username:     "user",
		PasswordHash: string(hash),
	}

	svc := auth.New(repo, session, reset, pub, zap.NewNop(), cfg)

	login := &models.LoginUser{
		Email:    "user@novsu.ru",
		Password: password,
	}

	resp, err := svc.Login(context.Background(), login, "127.0.0.1", "Mozilla")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	if resp.AccessToken == "" {
		t.Error("expected access token to be returned")
	}

	if resp.RefreshToken == "" {
		t.Error("expected refresh token to be returned")
	}

	// Verify session set in Redis
	userIDInSession, err := session.GetUserIDByRefreshToken(context.Background(), resp.RefreshToken)
	if err != nil {
		t.Fatalf("failed to retrieve session from Redis: %v", err)
	}
	if userIDInSession != userID.String() {
		t.Errorf("expected session userID %s, got %s", userID.String(), userIDInSession)
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	repo := &mockAuthRepo{users: make(map[string]*models.User)}
	session := &mockSessionRepo{tokens: make(map[string]string)}
	reset := &mockResetRepo{codes: make(map[string]string)}
	pub := &mockPublisher{}

	cfg := config.AuthConfig{
		JWTSecret:      "secret_key_secret_key_secret_key",
		AccessTokenTTL: 5 * time.Minute,
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	userID, _ := uuid.NewV7()

	repo.users["user@novsu.ru"] = &models.User{
		ID:           userID,
		Email:        "user@novsu.ru",
		Username:     "user",
		PasswordHash: string(hash),
	}

	svc := auth.New(repo, session, reset, pub, zap.NewNop(), cfg)

	login := &models.LoginUser{
		Email:    "user@novsu.ru",
		Password: "wrongpassword",
	}

	_, err := svc.Login(context.Background(), login, "127.0.0.1", "Mozilla")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
