package user

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

type Email struct {
	host      string
	port      int
	username  string
	password  string
	from      string
	fromName  string
	templates *EmailTemplates
}

func (e *Email) SendVerificationEmail(ctx context.Context, email, token, displayName string) error {
	subject := "Verify your email address"
	body := e.templates.GetVerificationEmailTemplate(displayName, token)
	return e.sendEmail(email, subject, body)
}
func (e *Email) SendPasswordResetEmail(ctx context.Context, email, token, displayName string) error {
	subject := "Reset your password"
	body := e.templates.GetPasswordResetEmailTemplate(displayName, token)
	return e.sendEmail(email, subject, body)
}
func (e *Email) SendWelcomeEmail(ctx context.Context, email, displayName string) error {
	subject := "Welcome to NameThatSong!"
	body := e.templates.GetWelcomeEmailTemplate(displayName)
	return e.sendEmail(email, subject, body)
}
func (e *Email) SendPasswordResetConfirmation(ctx context.Context, email, displayName string) error {
	subject := "Password successfully reset"
	body := e.templates.GetPasswordResetConfirmationTemplate(displayName)
	return e.sendEmail(email, subject, body)
}

// sendEmail is a helper method that handles the actual SMTP sending
func (e *Email) sendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(e.from, e.fromName))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// TODO: add as interface? so it can be configured
	d := gomail.NewDialer(e.host, e.port, e.username, e.password)
	return d.DialAndSend(m)
}

// EmailConfig holds email service configuration
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
	BaseURL  string
}

// ValidateEmailConfig validates email configuration
func ValidateEmailConfig(config *EmailConfig) error {
	if config.Host == "" {
		return fmt.Errorf("EMAIL_HOST is required")
	}
	if config.Port == 0 {
		return fmt.Errorf("EMAIL_PORT is required and must be a valid port number")
	}
	if config.Username == "" {
		return fmt.Errorf("EMAIL_USERNAME is required")
	}
	if config.Password == "" {
		return fmt.Errorf("EMAIL_PASSWORD is required")
	}
	if config.From == "" {
		return fmt.Errorf("EMAIL_FROM is required")
	}
	if config.FromName == "" {
		return fmt.Errorf("EMAIL_FROM_NAME is required")
	}
	if config.BaseURL == "" {
		return fmt.Errorf("BASE_URL is required")
	}
	return nil
}

// NewEmailServiceWithConfig creates a new email service with provided configuration
func NewEmailServiceWithConfig(config *EmailConfig) (*Email, error) {
	if err := ValidateEmailConfig(config); err != nil {
		return nil, fmt.Errorf("invalid email configuration: %w", err)
	}

	return &Email{
		host:      config.Host,
		port:      config.Port,
		username:  config.Username,
		password:  config.Password,
		from:      config.From,
		fromName:  config.FromName,
		templates: NewEmailTemplates(config.BaseURL),
	}, nil
}

// NewEmailService creates a new email service with configuration from environment variables
func NewEmailService() (*Email, error) {
	port, err := strconv.Atoi(os.Getenv("EMAIL_PORT"))
	if err != nil {
		return nil, fmt.Errorf("EMAIL_PORT must be a valid integer: %w", err)
	}

	config := &EmailConfig{
		Host:     os.Getenv("EMAIL_HOST"),
		Port:     port,
		Username: os.Getenv("EMAIL_USERNAME"),
		Password: os.Getenv("EMAIL_PASSWORD"),
		From:     os.Getenv("EMAIL_FROM"),
		FromName: os.Getenv("EMAIL_FROM_NAME"),
		BaseURL:  os.Getenv("BASE_URL"),
	}

	// Fallback for BASE_URL if not set
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:8080"
	}

	return NewEmailServiceWithConfig(config)
}
