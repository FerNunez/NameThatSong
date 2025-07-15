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
	IPAddress string     `json:"ip_address"`
	UserAgent string     `json:"user_agent"`
	CreatedAt time.Time  `json:"created_at"`
}

// GameSession represents a completed game session
type GameSession struct {
	ID              uuid.UUID `json:"id"` // game session's ID
	UserID          uuid.UUID `json:"user_id"`
	GameMode        string    `json:"game_mode"` // 'guess_song', 'guess_artist', etc.
	TotalQuestions  int       `json:"total_questions"`
	CorrectAnswers  int       `json:"correct_answers"`
	TotalScore      int       `json:"total_score"`
	DurationSeconds int       `json:"duration_seconds"`
	TracksPlayed    []string  `json:"tracks_played"` // Array of Spotify track IDs
	CompletedAt     time.Time `json:"completed_at"`
}

// Accuracy calculates the accuracy percentage for the game session
func (g *GameSession) Accuracy() float64 {
	if g.TotalQuestions == 0 {
		return 0.0
	}
	return float64(g.CorrectAnswers) / float64(g.TotalQuestions) * 100.0
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
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=8"`
	DisplayName string `json:"display_name" validate:"required,min=2,max=50"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	User      *User     `json:"user"`
	SessionID string    `json:"session_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name,omitempty" validate:"omitempty,min=2,max=50"`
	AvatarURL   string `json:"avatar_url,omitempty" validate:"omitempty,url"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents a password reset request
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// VerifyEmailRequest represents an email verification request
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// ResendVerificationRequest represents a request to resend verification email
type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}
