package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	// "github.com/sudo-odner/minor/backend/services/auth_service/internal/lib/otp"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidResetToken = errors.New("reset token is invalid or expired")
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

	// code, err := otp.GenerateRandomCode(6)
	// if err != nil {
	// 	log.Warn("failed to generate reset code", zap.Error(err))

	// 	return fmt.Errorf("%s: %w", path, err)
	// }

	resetToken := uuid.New().String()

	err = as.resetRepository.SetResetCode(ctx, payload.Email, resetToken, 15*time.Minute)
	if err != nil {
		log.Warn("failed to set reset code in redis", zap.Error(err))

		return fmt.Errorf("%s: %w", path, err)
	}

	err = as.eventPublisher.PublishPasswordResetRequested(ctx, payload.Email, resetToken, user.Username)
	if err != nil {
		log.Warn("failed to publish event", zap.Error(err))

		return fmt.Errorf("%s: %w", path, err)
	}

	log.Info("SUCCESS: reset event published to NATS")

	return nil
}

func (as *AuthService) ResetPassword(ctx context.Context, payload *models.ResetPasswordPayload) error {
	const path = "service.auth.ResetPassword"

	log := as.log.With(
		zap.String("path", path),
	)

	email, err := as.resetRepository.GetEmailByResetToken(ctx, payload.Token)
	if err != nil {
		log.Warn("invalid reset token attempt", zap.String("token", payload.Token), zap.String("path", path))
		return ErrInvalidResetToken
	}

	user, err := as.authRepository.GetByEmail(ctx, email)
	if err != nil {
		log.Warn("failed to get user by email", zap.Error(err))

		return fmt.Errorf("%s: %w", path, err)
	}

	fmt.Println("new Password before hashing:", payload.Password)
	
	hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Warn("failed to generate hash from password", zap.Error(err))

		return fmt.Errorf("%s: %w", path, err)
	}

	fmt.Println("SAVE TO DB HASH:", string(hash))
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

	// err = as.eventPublisher.PublishPasswordChanged(ctx, user.ID.String())
	// if err != nil {
	// 	log.Error("failed to publish password changed event", zap.Error(err))
	// }

	return nil
}
