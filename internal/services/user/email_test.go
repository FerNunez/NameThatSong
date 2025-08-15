package user_test

import (
	"context"
	"strings"
	"testing"

	"github.com/FerNunez/NameThatSong/internal/config"
	"github.com/FerNunez/NameThatSong/internal/services/user"
)

func TestNewEmailService(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.EmailConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &config.EmailConfig{
				SMTPHost:     "smtp.gmail.com",
				SMTPPort:     "587",
				SMTPUsername: "test@gmail.com",
				SMTPPassword: "password",
				FromEmail:    "test@gmail.com",
				FromName:     "Test User",
				BaseURL:      "https://example.com",
			},
			expectError: false,
		},
		{
			name: "nil config",
			config: nil,
			expectError: true,
			errorMsg:    "email config is required",
		},
		{
			name: "invalid port",
			config: &config.EmailConfig{
				SMTPHost:     "smtp.gmail.com",
				SMTPPort:     "invalid",
				SMTPUsername: "test@gmail.com",
				SMTPPassword: "password",
				FromEmail:    "test@gmail.com",
				FromName:     "Test User",
				BaseURL:      "https://example.com",
			},
			expectError: true,
			errorMsg:    "invalid SMTP port:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := user.NewEmailService(tt.config)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
				if service != nil {
					t.Errorf("Expected nil service but got: %v", service)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if service == nil {
					t.Errorf("Expected service but got nil")
				}
			}
		})
	}
}


func TestEmailTemplateIntegration(t *testing.T) {
	config := &config.EmailConfig{
		SMTPHost:     "smtp.example.com",
		SMTPPort:     "587",
		SMTPUsername: "test@example.com",
		SMTPPassword: "password",
		FromEmail:    "test@example.com",
		FromName:     "Test Service",
		BaseURL:      "https://example.com",
	}

	service, err := user.NewEmailService(config)
	if err != nil {
		t.Fatalf("Failed to create email service: %v", err)
	}

	ctx := context.Background()
	testEmail := "user@example.com"
	testToken := "test-token-123"
	testName := "Test User"

	t.Run("SendVerificationEmail method exists", func(t *testing.T) {
		err := service.SendVerificationEmail(ctx, testEmail, testToken, testName)
		if err == nil {
			t.Error("Expected SMTP error since we're not actually sending, but method should exist")
		}
	})

	t.Run("SendPasswordResetEmail method exists", func(t *testing.T) {
		err := service.SendPasswordResetEmail(ctx, testEmail, testToken, testName)
		if err == nil {
			t.Error("Expected SMTP error since we're not actually sending, but method should exist")
		}
	})

	t.Run("SendWelcomeEmail method exists", func(t *testing.T) {
		err := service.SendWelcomeEmail(ctx, testEmail, testName)
		if err == nil {
			t.Error("Expected SMTP error since we're not actually sending, but method should exist")
		}
	})

	t.Run("SendPasswordResetConfirmation method exists", func(t *testing.T) {
		err := service.SendPasswordResetConfirmation(ctx, testEmail, testName)
		if err == nil {
			t.Error("Expected SMTP error since we're not actually sending, but method should exist")
		}
	})
}

