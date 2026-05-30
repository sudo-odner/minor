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
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// NATS interface
type EventPublisher interface {
	PublishUserRegistered(ctx context.Context, user *models.User) error
	PublishLoginSuccess(ctx context.Context, userID, ip, userAgent string) error
    PublishUserLoggedOut(ctx context.Context, userID, tokenID string) error
}

// Redis interface
type SessionRepository interface {
	SetRefreshToken(ctx context.Context, userID string, tokenID string, ttl time.Duration) error
	GetUserIDByRefreshToken(ctx context.Context, tokenID string) (string, error)
	DeleteRefreshToken(ctx context.Context, tokenID string) error
	DeleteAllUserSessions(ctx context.Context, userID string) error
}

// Postgres interface
type AuthRepository interface {
	Create(ctx context.Context, newUser *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id string) (*models.User, error)
}

type AuthService struct {
	authRepository    AuthRepository
	sessionRepository SessionRepository
	eventPublisher    EventPublisher
	log               *zap.Logger
	authConfig       config.AuthConfig
}

func New(authRepository AuthRepository, sessionRepository SessionRepository, eventPublisher EventPublisher, log *zap.Logger, tokenConfig config.AuthConfig) *AuthService {
	return &AuthService{
		authRepository: authRepository,
		sessionRepository: sessionRepository,
		eventPublisher: eventPublisher,
		log:            log,
		authConfig: tokenConfig,
	}
}

func (as *AuthService) Login(ctx context.Context, newUser *models.LoginUser, ip, userAgent string) (authResponse *models.AuthResponse, err error) {
	const op = "service.auth.Login"

	log := as.log.With(
		zap.String("op", op),
	)

	log.Info("attempting to login user")

	user, err := as.authRepository.GetByEmail(ctx, newUser.Email)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(newUser.Password))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	accessToken, err := jwt.GenerateAccessToken(as.authConfig, user)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refreshToken := uuid.New().String()

	err = as.sessionRepository.SetRefreshToken(ctx, user.ID.String(), refreshToken, 30*24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = as.eventPublisher.PublishLoginSuccess(ctx, user.ID.String(), ip, userAgent)
    if err != nil {
        log.Error("failed to publish login event", zap.Error(err))
    }

	return &models.AuthResponse{
		User: &models.NormalizedUser{
			ID:    user.ID,
			Email: user.Email,
		},
		AccessToken:   accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (as *AuthService) Register(ctx context.Context, newUser *models.RegisterUser) (user *models.AuthResponse, err error) {
	const op = "service.auth.Register"

	log := as.log.With(
		zap.String("op", op),
	)

	log.Info("attempting to register new user")

	existingUser, err := as.authRepository.GetByEmail(ctx, newUser.Email)
	if err == nil && existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	id, err := uuid.NewV7()
	if err != nil {
		log.Error("%s: %w", zap.String("path:", op), zap.Error(err))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = as.authRepository.Create(ctx, &models.User{ID: id, Email: newUser.Email, PasswordHash: string(hashedPassword)})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	accessToken, err := jwt.GenerateAccessToken(as.authConfig, &models.User{ID: id, Email: newUser.Email, Username: newUser.Username})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	refreshToken := uuid.New().String()

	err = as.sessionRepository.SetRefreshToken(ctx, id.String(), refreshToken, 30*24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = as.eventPublisher.PublishUserRegistered(ctx, &models.User{ID: id, Email: newUser.Email})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &models.AuthResponse{
		User:         &models.NormalizedUser{ID: id, Email: newUser.Email},
		AccessToken:   accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (as *AuthService) Logout(ctx context.Context, refreshToken string) error {
	const path = "service.auth.Logout"
	
	log := as.log.With(
		zap.String("path", path),
	)
	
	userID, err := as.sessionRepository.GetUserIDByRefreshToken(ctx, refreshToken)
	if err != nil {
		log.Warn("logout attempted with non-existent token", zap.String("token", refreshToken))
		return nil 
	}

	err = as.sessionRepository.DeleteRefreshToken(ctx, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Это позволит Gateway Service мгновенно разорвать WebSocket соединение
	err = as.eventPublisher.PublishUserLoggedOut(ctx, userID, refreshToken)
	if err != nil {
		log.Error("failed to publish logout event", zap.Error(err))
	}

	log.Info("user logged out successfully", zap.String("user_id", userID))
	return nil
}

func (as *AuthService) RefreshAccessToken(ctx context.Context, oldRefreshToken string) (*models.AuthResponse, error) {
	userID, err := as.sessionRepository.GetUserIDByRefreshToken(ctx, oldRefreshToken)
	if err != nil {
		return nil, errors.New("session expired or invalid")
	}

	_ = as.sessionRepository.DeleteRefreshToken(ctx, oldRefreshToken)

	user, err := as.authRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if !user.IsActive {
		return nil, errors.New("user account is disabled")
	}

	newAccessToken, err := jwt.GenerateAccessToken(as.authConfig, user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	err = as.sessionRepository.SetRefreshToken(
		ctx, 
		user.ID.String(), 
		newRefreshToken.String(), 
		as.authConfig.RefreshTokenTTL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save new session: %w", err)
	}

	return &models.AuthResponse{
		User:         &models.NormalizedUser{ID: user.ID, Email: user.Email},
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken.String(),
	}, nil
}