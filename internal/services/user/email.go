package user

import (
	"context"
	"fmt"
	"strconv"

	"github.com/FerNunez/NameThatSong/internal/config"
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

// NewEmailService creates a new email service with configuration from EmailConfig
func NewEmailService(emailConfig *config.EmailConfig) (*Email, error) {
	if emailConfig == nil {
		return nil, fmt.Errorf("email config is required")
	}

	// Convert port from string to int
	port, err := strconv.Atoi(emailConfig.SMTPPort)
	if err != nil {
		return nil, fmt.Errorf("invalid SMTP port: %w", err)
	}

	return &Email{
		host:      emailConfig.SMTPHost,
		port:      port,
		username:  emailConfig.SMTPUsername,
		password:  emailConfig.SMTPPassword,
		from:      emailConfig.FromEmail,
		fromName:  emailConfig.FromName,
		templates: NewEmailTemplates(emailConfig.BaseURL),
	}, nil
}
