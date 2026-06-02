package jwt_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/config"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/lib/jwt"
	"github.com/sudo-odner/minor/backend/services/auth_service/internal/models"
)

func TestGenerateAndValidateAccessToken(t *testing.T) {
	cfg := config.AuthConfig{
		JWTSecret:      "test_secret_key_1234567890",
		AccessTokenTTL: 5 * time.Minute,
	}

	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@novsu.ru",
		Username: "testuser",
	}

	// 1. Generate Token
	tokenString, err := jwt.GenerateAccessToken(cfg, user)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	if tokenString == "" {
		t.Fatal("expected token string, got empty string")
	}

	// 2. Validate Token
	claims, err := jwt.ValidateAccessToken(cfg, tokenString)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}

	if claims.UserID != user.ID.String() {
		t.Errorf("expected userID %s, got %s", user.ID.String(), claims.UserID)
	}

	if claims.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, claims.Email)
	}

	if claims.Username != user.Username {
		t.Errorf("expected username %s, got %s", user.Username, claims.Username)
	}
}

func TestValidateAccessToken_InvalidSecret(t *testing.T) {
	cfgGen := config.AuthConfig{
		JWTSecret:      "secret_one",
		AccessTokenTTL: 5 * time.Minute,
	}

	cfgVal := config.AuthConfig{
		JWTSecret:      "secret_two",
		AccessTokenTTL: 5 * time.Minute,
	}

	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@novsu.ru",
		Username: "testuser",
	}

	tokenString, err := jwt.GenerateAccessToken(cfgGen, user)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = jwt.ValidateAccessToken(cfgVal, tokenString)
	if err == nil {
		t.Fatal("expected validation to fail due to secret mismatch, but it succeeded")
	}
}
