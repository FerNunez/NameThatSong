package user

import (
	"context"
	"fmt"
	"time"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/pkg/utils"
	"github.com/FerNunez/NameThatSong/internal/pkg/validation"
	"github.com/FerNunez/NameThatSong/internal/repository"
	"github.com/google/uuid"
)

type User struct {
	userStore              repository.UserStore
	emailVerificationStore repository.EmailVerificationStore
	passwordResetStore     repository.PasswordResetStore
	sessionStore           repository.UserSessionStore
	emailService           EmailService
}

func NewUserService(
	userStore repository.UserStore,
	emailVerificationStore repository.EmailVerificationStore,
	passwordResetStore repository.PasswordResetStore,
	sessionStore repository.UserSessionStore,
	emailService EmailService,
) *User {
	return &User{
		userStore:              userStore,
		emailVerificationStore: emailVerificationStore,
		passwordResetStore:     passwordResetStore,
		sessionStore:           sessionStore,
		emailService:           emailService,
	}
}

// Authentication
func (u *User) Register(ctx context.Context, req models.RegisterRequest) (*models.User, error) {
	// Validate registration request
	if err := u.validateRegisterRequest(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	
	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	// Create user
	dbUser, err := u.userStore.Create(ctx, req.Email, hashedPassword)
	if err != nil {
		return nil, err
	}
	return dbUser, nil
}
func (u *User) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	// Validate login request
	if err := u.validateLoginRequest(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	
	// Get user
	dbUser, err := u.userStore.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	// Check password with stored
	if err := utils.CheckPasswordHash(req.Password, dbUser.HashedPassword); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	// SessionToken creation
	dbSession, err := u.sessionStore.Create(ctx, dbUser.ID, time.Duration(24)*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to create user session: %w", err)
	}
	// Last login at
	if err := u.userStore.UpdateLastLogin(ctx, dbUser.ID); err != nil {
		return nil, err
	}
	return &models.LoginResponse{
		User:      dbUser,
		SessionID: dbSession.ID,
		ExpiresAt: dbSession.ExpiresAt,
	}, nil
}
func (u *User) Logout(ctx context.Context, sessionID string) error {
	return u.sessionStore.Revoke(ctx, sessionID)
}

// Email Verification
func (u *User) SendEmailVerification(ctx context.Context, userID uuid.UUID) error {
	// Get user
	dbUser, err := u.userStore.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	// Verification Email token
	token, err := utils.GenerateState(16)
	if err != nil {
		return err
	}
	// Store in Database
	evt, err := u.emailVerificationStore.Create(ctx, userID, token, time.Now().Add(time.Duration(10)*time.Minute))
	if err != nil {
		return err
	}
	// Send email to user
	return u.emailService.SendVerificationEmail(ctx, dbUser.Email, evt.Token, dbUser.DisplayName)
}
func (u *User) VerifyEmail(ctx context.Context, token string) error {
	// Validate token
	if err := validation.ValidateToken(token); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	
	// Get verification token
	evt, err := u.emailVerificationStore.GetByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid verification token: %w", err)
	}
	// Check if token has expired
	if time.Now().After(evt.ExpiresAt) {
		return fmt.Errorf("verification token has expired")
	}

	// Check if token was already used
	if evt.UsedAt != nil {
		return fmt.Errorf("verification token has already been used")
	}
	// Get user
	dbUser, err := u.userStore.GetByID(ctx, evt.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if user is already verified
	if dbUser.EmailVerified {
		return fmt.Errorf("email is already verified")
	}

	// Mark EmailVerification as used
	if err := u.emailVerificationStore.MarkAsUsed(ctx, token); err != nil {
		return fmt.Errorf("failed to mark verification token as used: %w", err)
	}

	// Update user's email_verified status
	if err := u.userStore.VerifyUserEmail(ctx, evt.UserID); err != nil {
		return fmt.Errorf("failed to verify user email: %w", err)
	}

	// Send welcome email
	if err := u.emailService.SendWelcomeEmail(ctx, dbUser.Email, dbUser.DisplayName); err != nil {
		return fmt.Errorf("failed to send welcome email: %w", err)
	}
	return nil
}
func (u *User) ResendVerification(ctx context.Context, email string) error {
	// Validate email
	if err := validation.ValidateEmail(email); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	
	// Get user
	dbUser, err := u.userStore.GetByEmail(ctx, email)
	if err != nil {
		return nil
	}

	// Check if user is already verified
	if dbUser.EmailVerified {
		return fmt.Errorf("email is already verified")
	}

	// Generate new verification token
	token, err := utils.GenerateState(16)
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}

	// Create new verification token
	expiresAt := time.Now().Add(10 * time.Minute)
	_, err = u.emailVerificationStore.Create(ctx, dbUser.ID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create verification token: %w", err)
	}

	// Send verification email
	if err := u.emailService.SendVerificationEmail(ctx, dbUser.Email, token, dbUser.DisplayName); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}
	return nil
}

// Password Reset
func (u *User) InitiatePasswordReset(ctx context.Context, email string) error {
	// Validate email
	if err := validation.ValidateEmail(email); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	
	// Get user by email
	dbUser, err := u.userStore.GetByEmail(ctx, email)
	if err != nil {
		// For security, don't reveal if email exists or not
		return nil
	}

	// Generate secure reset token
	token, err := utils.GenerateState(32) // Longer token for password reset
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Store token
	expiresAt := time.Now().Add(10 * time.Minute)

	_, err = u.passwordResetStore.Create(ctx, dbUser.ID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}

	// Send password reset email
	if err := u.emailService.SendPasswordResetEmail(ctx, dbUser.Email, token, dbUser.DisplayName); err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}
func (u *User) ResetPassword(ctx context.Context, token string, newPassword string) error {
	// Validate inputs
	if err := validation.ValidateToken(token); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := validation.ValidatePassword(newPassword); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	
	// Get password reset token
	prt, err := u.passwordResetStore.GetByToken(ctx, token)
	if err != nil {
		return fmt.Errorf("invalid password reset token: %w", err)
	}

	// Check if token has expired
	if time.Now().After(prt.ExpiresAt) {
		return fmt.Errorf("password reset token has expired")
	}

	// Check if token was already used
	if prt.UsedAt != nil {
		return fmt.Errorf("password reset token has already been used")
	}

	// Get user
	dbUser, err := u.userStore.GetByID(ctx, prt.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user's password
	if err := u.userStore.UpdatePasswordByID(ctx, dbUser.ID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	if err := u.passwordResetStore.MarkAsUsed(ctx, token); err != nil {
		return fmt.Errorf("failed to mark password reset token as used: %w", err)
	}

	// Revoke all user sessions for security
	if err := u.sessionStore.RevokeAllSessions(ctx, dbUser.ID); err != nil {
		return fmt.Errorf("failed to revoke user sessions: %w", err)
	}

	// Send confirmation email
	if err := u.emailService.SendPasswordResetConfirmation(ctx, dbUser.Email, dbUser.DisplayName); err != nil {
		return fmt.Errorf("failed to send password reset confirmation: %w", err)
	}

	return nil
}

// User Management
func (u *User) GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	dbUser, err := u.userStore.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return dbUser, nil
}
func (u *User) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	// Validate email
	if err := validation.ValidateEmail(email); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	
	dbUser, err := u.userStore.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return dbUser, nil
}
func (u *User) UpdateProfile(ctx context.Context, userID uuid.UUID, req models.UpdateProfileRequest) (*models.User, error) {
	// Validate profile update request
	if err := u.validateUpdateProfileRequest(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}
	
	// Get current user to verify they exist
	dbUser, err := u.userStore.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Prepare update values - use existing values if not provided
	displayName := req.DisplayName
	if displayName == "" {
		displayName = dbUser.DisplayName
	}

	avatarURL := req.AvatarURL
	if avatarURL == "" {
		avatarURL = dbUser.AvatarURL
	}

	// Update profile
	if err := u.userStore.UpdateProfileByID(ctx, userID, displayName, avatarURL); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Return updated user
	updatedUser, err := u.userStore.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated user: %w", err)
	}

	return updatedUser, nil
}
func (u *User) ChangePassword(ctx context.Context, userID uuid.UUID, req models.ChangePasswordRequest) error {
	// Validate password change request
	if err := u.validateChangePasswordRequest(req); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	
	// Get current user
	dbUser, err := u.userStore.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify current password
	if err := utils.CheckPasswordHash(req.CurrentPassword, dbUser.HashedPassword); err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	if err := u.userStore.UpdatePasswordByID(ctx, userID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// For security, revoke all sessions except current one
	if err := u.sessionStore.RevokeAllSessions(ctx, userID); err != nil {
		return fmt.Errorf("failed to revoke sessions: %w", err)
	}

	return nil
}

// Session Management
func (u *User) CreateSession(ctx context.Context, userID uuid.UUID, deviceInfo, ipAddress, userAgent string) (string, error) {
	// Verify user exists
	_, err := u.userStore.GetByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	// Create session with 24-hour expiration
	dbSession, err := u.sessionStore.Create(ctx, userID, 24*time.Hour)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	return dbSession.ID, nil
}
func (u *User) ValidateSession(ctx context.Context, sessionID string) (*models.User, error) {
	// Get session
	dbSession, err := u.sessionStore.Get(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	// Check if session is expired
	if time.Now().After(dbSession.ExpiresAt) {
		return nil, fmt.Errorf("session has expired")
	}

	// Check if session is revoked
	if dbSession.RevokedAt != nil {
		return nil, fmt.Errorf("session has been revoked")
	}

	// Get user associated with session
	dbUser, err := u.userStore.GetByID(ctx, dbSession.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return dbUser, nil
}

// Validation helper methods
func (u *User) validateRegisterRequest(req models.RegisterRequest) error {
	if err := validation.ValidateEmail(req.Email); err != nil {
		return err
	}
	if err := validation.ValidatePassword(req.Password); err != nil {
		return err
	}
	return nil
}

func (u *User) validateLoginRequest(req models.LoginRequest) error {
	if err := validation.ValidateEmail(req.Email); err != nil {
		return err
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func (u *User) validateUpdateProfileRequest(req models.UpdateProfileRequest) error {
	if err := validation.ValidateDisplayName(req.DisplayName); err != nil {
		return err
	}
	if err := validation.ValidateAvatarURL(req.AvatarURL); err != nil {
		return err
	}
	return nil
}

func (u *User) validateChangePasswordRequest(req models.ChangePasswordRequest) error {
	if req.CurrentPassword == "" {
		return fmt.Errorf("current password is required")
	}
	if err := validation.ValidatePassword(req.NewPassword); err != nil {
		return err
	}
	if req.CurrentPassword == req.NewPassword {
		return fmt.Errorf("new password must be different from current password")
	}
	return nil
}
