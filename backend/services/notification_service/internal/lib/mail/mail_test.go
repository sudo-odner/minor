package mail_test

import (
	"context"
	"os"
	"testing"

	"github.com/sudo-odner/minor/backend/services/notification_service/internal/lib/mail"
)

func TestSMTPProvider_Send_EmptyConfig(t *testing.T) {
	// Clear SMTP envs
	os.Setenv("SMTP_HOST", "")
	os.Setenv("SMTP_PORT", "")
	os.Setenv("SMTP_USER", "")
	os.Setenv("SMTP_PASS", "")
	os.Setenv("SMTP_FROM", "")

	provider := mail.NewSMTPProvider()

	// Should return error since SMTP_HOST is empty
	err := provider.Send(context.Background(), "user@novsu.ru", "Title", "Body")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedErr := "SMTP host or recipient email is empty"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}

func TestSMTPProvider_Send_EmptyRecipient(t *testing.T) {
	os.Setenv("SMTP_HOST", "smtp.novsu.ru")
	os.Setenv("SMTP_PORT", "587")

	provider := mail.NewSMTPProvider()

	// Should return error since recipient email is empty
	err := provider.Send(context.Background(), "", "Title", "Body")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	expectedErr := "SMTP host or recipient email is empty"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}
