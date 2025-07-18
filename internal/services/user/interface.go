package user

import (
	"context"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/google/uuid"
)

// UserService defines the interface for user management operations
type UserService interface {
	// Authentication
	Register(ctx context.Context, req models.RegisterRequest) (*models.User, error)
	Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error)
	Logout(ctx context.Context, sessionID string) error

	// Email Verification
	SendEmailVerification(ctx context.Context, userID uuid.UUID) error
	VerifyEmail(ctx context.Context, token string) error
	ResendVerification(ctx context.Context, email string) error

	// Password Reset
	InitiatePasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token string, newPassword string) error

	// User Management
	GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req models.UpdateProfileRequest) (*models.User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, req models.ChangePasswordRequest) error

	// Session Management
	CreateSession(ctx context.Context, userID uuid.UUID, deviceInfo, ipAddress, userAgent string) (string, error)
	ValidateSession(ctx context.Context, sessionID string) (*models.User, error)
}

// EmailService defines the interface for sending emails
type EmailService interface {
	SendVerificationEmail(ctx context.Context, email, token, displayName string) error
	SendPasswordResetEmail(ctx context.Context, email, token, displayName string) error
	SendWelcomeEmail(ctx context.Context, email, displayName string) error
	SendPasswordResetConfirmation(ctx context.Context, email, displayName string) error
}

