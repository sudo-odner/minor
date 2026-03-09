package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sudo-odner/min/backend/auth-service/internal/config"
	"github.com/sudo-odner/min/backend/auth-service/internal/lib/jwt"
	"github.com/sudo-odner/min/backend/auth-service/internal/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthorizationRepository interface {
	Create(ctx context.Context, input models.User) (error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	// VerifyEmail(ctx context.Context, code string) error
}

type AuthorizationService struct {
	authRepository AuthorizationRepository
	log *zap.Logger
	tokenConfig config.TokenConfig
}

func New(authRepository AuthorizationRepository, log *zap.Logger) *AuthorizationService {
	return &AuthorizationService{
		authRepository: authRepository,
		log: log,
	}
}

func (as *AuthorizationService) Login(ctx context.Context, email string, password string) (user *models.User, accessToken string, refreshToken string, err error) {
	const op = "service.auth.Login"

	log := as.log.With(
		zap.String("op", op),
	)

	log.Info("attempting to login user")

	user, err = as.authRepository.GetByEmail(ctx, email)
	if err != nil {
		return nil, "", "", fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, "", "", fmt.Errorf("%s: %w", op, err)
	}

	accessToken, refreshToken, err = jwt.GenerateTokens(as.tokenConfig, user.ID, user.Email)
	if err != nil {
		return nil, "", "", fmt.Errorf("%s: %w", op, err)
	}

	return user, accessToken, refreshToken, nil
}

func (as *AuthorizationService) Register(ctx context.Context, email string, username string, password string) (user *models.User, accessToken string, refreshToken string, err error) {
	const op = "service.auth.Register"

	log := as.log.With(
		zap.String("op", op),
	)

	log.Info("attempting to register new user")

	return
}


