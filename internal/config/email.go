package config

import (
	"fmt"
	"os"
)

// EmailConfig holds all email-related configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	BaseURL      string
}

// NewEmailConfig creates a new email configuration from environment variables
func NewEmailConfig() (*EmailConfig, error) {
	smtpHost := os.Getenv("SMTP_HOST")
	if smtpHost == "" {
		return nil, fmt.Errorf("SMTP_HOST environment variable is required")
	}

	smtpPort := os.Getenv("SMTP_PORT")
	if smtpPort == "" {
		smtpPort = "587" // Default SMTP port
	}

	smtpUsername := os.Getenv("SMTP_USERNAME")
	if smtpUsername == "" {
		return nil, fmt.Errorf("SMTP_USERNAME environment variable is required")
	}

	smtpPassword := os.Getenv("SMTP_PASSWORD")
	if smtpPassword == "" {
		return nil, fmt.Errorf("SMTP_PASSWORD environment variable is required")
	}

	fromEmail := os.Getenv("FROM_EMAIL")
	if fromEmail == "" {
		fromEmail = smtpUsername // Default to SMTP username
	}

	fromName := os.Getenv("FROM_NAME")
	if fromName == "" {
		fromName = "NameThatSong" // Default application name
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Default for development
	}

	return &EmailConfig{
		SMTPHost:     smtpHost,
		SMTPPort:     smtpPort,
		SMTPUsername: smtpUsername,
		SMTPPassword: smtpPassword,
		FromEmail:    fromEmail,
		FromName:     fromName,
		BaseURL:      baseURL,
	}, nil
}

// GetSMTPAddress returns the full SMTP server address
func (c *EmailConfig) GetSMTPAddress() string {
	return c.SMTPHost + ":" + c.SMTPPort
}
