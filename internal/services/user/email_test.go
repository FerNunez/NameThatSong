package user_test

import (
	"context"
	"testing"

	"github.com/FerNunez/NameThatSong/internal/services/user"
)

func TestValidateEmailConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *user.EmailConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &user.EmailConfig{
				Host:     "smtp.gmail.com",
				Port:     587,
				Username: "test@gmail.com",
				Password: "password",
				From:     "test@gmail.com",
				FromName: "Test User",
				BaseURL:  "https://example.com",
			},
			expectError: false,
		},
		{
			name: "missing host",
			config: &user.EmailConfig{
				Port:     587,
				Username: "test@gmail.com",
				Password: "password",
				From:     "test@gmail.com",
				FromName: "Test User",
				BaseURL:  "https://example.com",
			},
			expectError: true,
			errorMsg:    "EMAIL_HOST is required",
		},
		{
			name: "missing port",
			config: &user.EmailConfig{
				Host:     "smtp.gmail.com",
				Username: "test@gmail.com",
				Password: "password",
				From:     "test@gmail.com",
				FromName: "Test User",
				BaseURL:  "https://example.com",
			},
			expectError: true,
			errorMsg:    "EMAIL_PORT is required and must be a valid port number",
		},
		{
			name: "missing username",
			config: &user.EmailConfig{
				Host:     "smtp.gmail.com",
				Port:     587,
				Password: "password",
				From:     "test@gmail.com",
				FromName: "Test User",
				BaseURL:  "https://example.com",
			},
			expectError: true,
			errorMsg:    "EMAIL_USERNAME is required",
		},
		{
			name: "missing password",
			config: &user.EmailConfig{
				Host:     "smtp.gmail.com",
				Port:     587,
				Username: "test@gmail.com",
				From:     "test@gmail.com",
				FromName: "Test User",
				BaseURL:  "https://example.com",
			},
			expectError: true,
			errorMsg:    "EMAIL_PASSWORD is required",
		},
		{
			name: "missing from",
			config: &user.EmailConfig{
				Host:     "smtp.gmail.com",
				Port:     587,
				Username: "test@gmail.com",
				Password: "password",
				FromName: "Test User",
				BaseURL:  "https://example.com",
			},
			expectError: true,
			errorMsg:    "EMAIL_FROM is required",
		},
		{
			name: "missing from name",
			config: &user.EmailConfig{
				Host:     "smtp.gmail.com",
				Port:     587,
				Username: "test@gmail.com",
				Password: "password",
				From:     "test@gmail.com",
				BaseURL:  "https://example.com",
			},
			expectError: true,
			errorMsg:    "EMAIL_FROM_NAME is required",
		},
		{
			name: "missing base URL",
			config: &user.EmailConfig{
				Host:     "smtp.gmail.com",
				Port:     587,
				Username: "test@gmail.com",
				Password: "password",
				From:     "test@gmail.com",
				FromName: "Test User",
			},
			expectError: true,
			errorMsg:    "BASE_URL is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := user.ValidateEmailConfig(tt.config)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestNewEmailServiceWithConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *user.EmailConfig
		expectError bool
	}{
		{
			name: "valid config creates service",
			config: &user.EmailConfig{
				Host:     "smtp.gmail.com",
				Port:     587,
				Username: "test@gmail.com",
				Password: "password",
				From:     "test@gmail.com",
				FromName: "Test User",
				BaseURL:  "https://example.com",
			},
			expectError: false,
		},
		{
			name: "invalid config returns error",
			config: &user.EmailConfig{
				Host: "smtp.gmail.com",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := user.NewEmailServiceWithConfig(tt.config)
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
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
	config := &user.EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		FromName: "Test Service",
		BaseURL:  "https://example.com",
	}

	service, err := user.NewEmailServiceWithConfig(config)
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