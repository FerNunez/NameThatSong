package user

import (
	"context"
	"fmt"
	"time"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/pkg/utils"
	"github.com/FerNunez/NameThatSong/internal/repository"
	"github.com/google/uuid"
)

type User struct {
	userStore              repository.UserStore
	emailVerificationStore repository.EmailVerificationStore
	passwordResetStore     repository.PasswordResetStore
	sessionStore           repository.SessionStore
	emailService           EmailService
}

func NewUserService(
	userStore repository.UserStore,
	emailVerificationStore repository.EmailVerificationStore,
	passwordResetStore repository.PasswordResetStore,
	sessionStore repository.SessionStore,
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
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	dbUser, err := u.userStore.Create(ctx, req.Email, hashedPassword)
	if err != nil {
		return nil, err
	}
	return dbUser, nil
}
func (u *User) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	dbUser, err := u.userStore.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if err := utils.CheckPasswordHash(req.Password, dbUser.HashedPassword); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// SessionToken creation:
	dbSession, err := u.sessionStore.Create(ctx, dbUser.ID, time.Duration(24)*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("couldnt create usersession for user: wrong password ")
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
	dbUser, err := u.userStore.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	token, err := utils.GenerateState(16)
	if err != nil {
		return err
	}

	evt, err := u.emailVerificationStore.Create(ctx, userID, token, time.Now().Add(time.Duration(10)*time.Minute))
	if err != nil {
		return err
	}
	return u.emailService.SendVerificationEmail(ctx, dbUser.Email, evt.Token, dbUser.DisplayName)
}
func (u *User) VerifyEmail(ctx context.Context, token string) error {
	evt, err := u.emailVerificationStore.GetByToken(ctx, token)
	if err != nil {
		return err
	}
	dbUser, err := u.userStore.GetByID(ctx, evt.UserID)
	if err != nil {
		return err
	}
	// Mark EmailVerification as used
	if err := u.emailVerificationStore.MarkAsUsed(ctx, token); err != nil {
		return err
	}

	return u.emailService.SendWelcomeEmail(ctx, dbUser.Email, dbUser.DisplayName)
}
func (u *User) ResendVerification(ctx context.Context, email string) error {
	return nil
}

// Password Reset
func (u *User) InitiatePasswordReset(ctx context.Context, email string) error {
	return nil
}
func (u *User) ResetPassword(ctx context.Context, token string, newPassword string) error {
	return nil
}

// User Management
func (u *User) GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	return nil, nil
}
func (u *User) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}
func (u *User) UpdateProfile(ctx context.Context, userID uuid.UUID, req models.UpdateProfileRequest) (*models.User, error) {
	return nil, nil
}
func (u *User) ChangePassword(ctx context.Context, userID uuid.UUID, req models.ChangePasswordRequest) error {
	return nil
}

// Session Management
func (u *User) CreateSession(ctx context.Context, userID uuid.UUID, deviceInfo, ipAddress, userAgent string) (string, error) {
	return "", nil
}
func (u *User) ValidateSession(ctx context.Context, sessionID string) (*models.User, error) {
	return nil, nil
}

// Game History
func (u *User) SaveGameSession(ctx context.Context, session models.GameSession) error {
	return nil
}
func (u *User) GetGameHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.GameSession, error) {
	return []models.GameSession{}, nil
}
func (u *User) GetGameSession(ctx context.Context, sessionID uuid.UUID) (*models.GameSession, error) {
	return nil, nil
}
