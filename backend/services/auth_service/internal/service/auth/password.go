package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/sudo-odner/minor/backend/services/auth_service/internal/lib/otp"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func (as *AuthService) ForgotPassword(ctx context.Context, payload *models.ForgotPasswordPayload) error {
	const path = "service.auth.ForgotPassword"

	log := as.log.With(
		zap.String("path", path),
	)

	user, err := as.authRepository.GetByEmail(ctx, payload.Email)
	if err != nil {
		log.Warn("failed to get user by email", zap.Error(err))

		return nil
	}

	code, err := otp.GenerateRandomCode(6)
	if err != nil {
		log.Warn("failed to generate reset code", zap.Error(err))

		return fmt.Errorf("%s: %w", path, err)
	}

	err = as.resetRepository.SetResetCode(ctx, payload.Email, code, 15*time.Minute)
	if err != nil {
		log.Warn("failed to set reset code in redis", zap.Error(err))

		return fmt.Errorf("%s: %w", path, err)
	}

	err = as.eventPublisher.PublishPasswordResetRequested(ctx, payload.Email, code, user.Username)
	if err != nil {
		log.Warn("failed to publish event", zap.Error(err))

		return fmt.Errorf("%s: %w", path, err)
	}

	return nil
}

func (as *AuthService) ResetPassword(ctx context.Context, payload *models.ResetPasswordPayload) error {
	const path = "service.auth.ResetPassword"

	log := as.log.With(
		zap.String("path", path),
	)

	savedCode, err := as.resetRepository.GetResetCode(ctx, payload.Email)
	if err != nil || savedCode != payload.Code {
		log.Warn("failed to compare reset codes", zap.Error(err))

		return fmt.Errorf("%s: %w", path, err)
	}

	user, err := as.authRepository.GetByEmail(ctx, payload.Email)
	if err != nil {
		log.Warn("failed to get user by email", zap.Error(err))

		return fmt.Errorf("%s: %w", path, err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Warn("failed to generate hash from password", zap.Error(err))

		return fmt.Errorf("%s: %w", path, err)
	}

	err = as.authRepository.UpdatePassword(ctx, user.ID.String(), string(hash))
	if err != nil {
		log.Warn("failed to update password", zap.Error(err))

		return fmt.Errorf("%s: %w", path, err)
	}

	err = as.sessionRepository.DeleteAllUserSessions(ctx, user.ID.String())
	if err != nil {
		log.Warn("failed to delete all user sessions")

		return fmt.Errorf("%s: %w", path, err)
	}

	err = as.resetRepository.DeleteResetCode(ctx, payload.Email)
	if err != nil {
		log.Warn("failed to delete reset code from redis", zap.Error(err))

		return fmt.Errorf("%s: %w", path, err)
	}

	return nil
}
