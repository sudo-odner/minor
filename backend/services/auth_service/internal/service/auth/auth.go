package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/lib/jwt"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

// NATS interface
type EventPublisher interface {
	PublishUserRegistered(ctx context.Context, user *models.User) error
}

// Redis interface
type SessionRepository interface {
	SetRefreshToken(ctx context.Context, userID string, token string, ttl time.Duration) error
}

type AuthorizationRepository interface {
	Create(ctx context.Context, input *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type AuthorizationService struct {
	authRepository    AuthorizationRepository
	sessionRepository SessionRepository
	eventPublisher    EventPublisher
	log               *zap.Logger
	tokenConfig       config.TokenConfig
}

func New(authRepository AuthorizationRepository, log *zap.Logger) *AuthorizationService {
	return &AuthorizationService{
		authRepository: authRepository,
		log:            log,
	}
}

func (as *AuthorizationService) Login(ctx context.Context, email string, password string) (authResponse *models.AuthResponse, err error) {
	const op = "service.auth.Login"

	log := as.log.With(
		zap.String("op", op),
	)

	log.Info("attempting to login user")

	user, err := as.authRepository.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	accessToken, err := jwt.GenerateTokens(as.tokenConfig, user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refreshToken := uuid.New().String()

	return &models.AuthResponse{
		User: &models.NormalizedUser{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
		},
		AcceeToken:   accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (as *AuthorizationService) Register(ctx context.Context, email string, username string, password string) (user *models.AuthResponse, err error) {
	const op = "service.auth.Register"

	log := as.log.With(
		zap.String("op", op),
	)

	log.Info("attempting to register new user")

	existingUser, err := as.authRepository.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	id, err := uuid.NewV7()
	if err != nil {
		log.Error("%s: %w", zap.String("path:", op), zap.Error(err))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = as.authRepository.Create(ctx, &models.User{ID: id, Email: email, PasswordHash: string(hashedPassword)})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	accessToken, err := jwt.GenerateTokens(as.tokenConfig, id, email)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refreshToken := uuid.New().String()

	err = as.sessionRepository.SetRefreshToken(ctx, id.String(), refreshToken, 30*24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.AuthResponse{
		User:         &models.NormalizedUser{ID: id, Email: email, Username: username},
		AcceeToken:   accessToken,
		RefreshToken: refreshToken,
	}, nil
}
