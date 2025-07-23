package models

import (
	"time"

	"github.com/google/uuid"
)

// User model
type User struct {
	ID             uuid.UUID  `json:"id"`
	Email          string     `json:"email"`
	HashedPassword string     `json:"-"` // Never serialize passwords
	DisplayName    string     `json:"display_name"`
	AvatarURL      string     `json:"avatar_url"`
	EmailVerified  bool       `json:"email_verified"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LastLoginAt    *time.Time `json:"last_login_at"`
}

// EmailVerificationToken represents a token for email verification
type EmailVerificationToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// PasswordResetToken represents a token for password reset
type PasswordResetToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// Represents tokens for user session
type UserSession struct {
	ID        string     `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Request/Response DTOs for API endpoints

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	User      *User     `json:"user"`
	SessionID string    `json:"session_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// VerifyEmailRequest represents an email verification request
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// ResendVerificationRequest represents a request to resend verification email
type ResendVerificationRequest struct {
	Email string `json:"email"`
}
