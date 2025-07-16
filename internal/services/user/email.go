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

	d := gomail.NewDialer(e.host, e.port, e.username, e.password)

	return d.DialAndSend(m)
}

// NewEmailService creates a new email service with configuration from environment variables
func NewEmailService() *Email {
	port, _ := strconv.Atoi(os.Getenv("EMAIL_PORT"))
	baseURL := "http://localhost:8080" // TODO: Make this configurable
	
	return &Email{
		host:      os.Getenv("EMAIL_HOST"),
		port:      port,
		username:  os.Getenv("EMAIL_USERNAME"),
		password:  os.Getenv("EMAIL_PASSWORD"),
		from:      os.Getenv("EMAIL_FROM"),
		fromName:  os.Getenv("EMAIL_FROM_NAME"),
		templates: NewEmailTemplates(baseURL),
	}
}
