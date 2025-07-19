package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/services/user"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock UserStore implementation
type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) Create(ctx context.Context, email, hashedPassword string) (*models.User, error) {
	args := m.Called(ctx, email, hashedPassword)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserStore) Delete(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserStore) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserStore) UpdatePasswordByID(ctx context.Context, id uuid.UUID, hashedPassword string) error {
	args := m.Called(ctx, id, hashedPassword)
	return args.Error(0)
}

func (m *MockUserStore) UpdateProfileByID(ctx context.Context, id uuid.UUID, displayName, avatarURL string) error {
	args := m.Called(ctx, id, displayName, avatarURL)
	return args.Error(0)
}

func (m *MockUserStore) VerifyUserEmail(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserStore) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserStore) Reset(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Mock EmailVerificationStore implementation
type MockEmailVerificationStore struct {
	mock.Mock
}

func (m *MockEmailVerificationStore) Create(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (*models.EmailVerificationToken, error) {
	args := m.Called(ctx, userID, token, expiresAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EmailVerificationToken), args.Error(1)
}

func (m *MockEmailVerificationStore) GetByToken(ctx context.Context, token string) (*models.EmailVerificationToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EmailVerificationToken), args.Error(1)
}

func (m *MockEmailVerificationStore) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.EmailVerificationToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EmailVerificationToken), args.Error(1)
}

func (m *MockEmailVerificationStore) MarkAsUsed(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockEmailVerificationStore) CleanupExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Mock PasswordResetStore implementation
type MockPasswordResetStore struct {
	mock.Mock
}

func (m *MockPasswordResetStore) Create(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (*models.PasswordResetToken, error) {
	args := m.Called(ctx, userID, token, expiresAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PasswordResetToken), args.Error(1)
}

func (m *MockPasswordResetStore) GetByToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PasswordResetToken), args.Error(1)
}

func (m *MockPasswordResetStore) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.PasswordResetToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PasswordResetToken), args.Error(1)
}

func (m *MockPasswordResetStore) DeleteByID(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPasswordResetStore) MarkAsUsed(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockPasswordResetStore) CleanupExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Mock UserSessionStore implementation
type MockUserSessionStore struct {
	mock.Mock
}

func (m *MockUserSessionStore) Create(ctx context.Context, userID uuid.UUID, ttl time.Duration) (*models.UserSession, error) {
	args := m.Called(ctx, userID, ttl)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserSession), args.Error(1)
}

func (m *MockUserSessionStore) Get(ctx context.Context, id string) (*models.UserSession, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserSession), args.Error(1)
}

func (m *MockUserSessionStore) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserSessionStore) Revoke(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserSessionStore) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Mock EmailService implementation
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendVerificationEmail(ctx context.Context, email, token, displayName string) error {
	args := m.Called(ctx, email, token, displayName)
	return args.Error(0)
}

func (m *MockEmailService) SendPasswordResetEmail(ctx context.Context, email, token, displayName string) error {
	args := m.Called(ctx, email, token, displayName)
	return args.Error(0)
}

func (m *MockEmailService) SendWelcomeEmail(ctx context.Context, email, displayName string) error {
	args := m.Called(ctx, email, displayName)
	return args.Error(0)
}

func (m *MockEmailService) SendPasswordResetConfirmation(ctx context.Context, email, displayName string) error {
	args := m.Called(ctx, email, displayName)
	return args.Error(0)
}

// Test setup helper function
func setupUserService() (*user.User, *MockUserStore, *MockEmailVerificationStore, *MockPasswordResetStore, *MockUserSessionStore, *MockEmailService) {
	userStore := &MockUserStore{}
	emailVerificationStore := &MockEmailVerificationStore{}
	passwordResetStore := &MockPasswordResetStore{}
	sessionStore := &MockUserSessionStore{}
	emailService := &MockEmailService{}

	userService := user.NewUserService(
		userStore,
		emailVerificationStore,
		passwordResetStore,
		sessionStore,
		emailService,
	)

	return userService, userStore, emailVerificationStore, passwordResetStore, sessionStore, emailService
}

// Test data helper functions
func createTestUser() *models.User {
	return &models.User{
		ID:             uuid.New(),
		Email:          "test@example.com",
		HashedPassword: "$2a$10$D8CH.c5dsOPX02XHshYcY.UMmi9Iz.F50lhxbYG.cQ9SyAO2O5nha", // hash for "Password123"
		DisplayName:    "Test User",
		AvatarURL:      "",
		EmailVerified:  false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastLoginAt:    nil,
	}
}

func createTestSession(userID uuid.UUID) *models.UserSession {
	return &models.UserSession{
		ID:        uuid.New().String(),
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		RevokedAt: nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Registration Tests
func TestUserService_Register(t *testing.T) {
	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		// Setup
		userService, userStore, _, _, _, _ := setupUserService()

		req := models.RegisterRequest{
			Email:       "test@example.com",
			Password:    "Password123",
			DisplayName: "Test User",
		}

		expectedUser := createTestUser()
		expectedUser.Email = req.Email

		// Mock expectation: userStore.Create() should be called with email and hashed password
		userStore.On("Create", ctx, req.Email, mock.AnythingOfType("string")).Return(expectedUser, nil)

		// Execute
		result, err := userService.Register(ctx, req)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedUser.Email, result.Email)
		assert.Equal(t, expectedUser.ID, result.ID)
		userStore.AssertExpectations(t)
	})

	t.Run("registration with duplicate email", func(t *testing.T) {
		// Setup
		userService, userStore, _, _, _, _ := setupUserService()

		req := models.RegisterRequest{
			Email:       "duplicate@example.com",
			Password:    "Password123",
			DisplayName: "Test User",
		}

		// Mock expectation: userStore.Create() should return error
		userStore.On("Create", ctx, req.Email, mock.AnythingOfType("string")).Return(nil, errors.New("duplicate email"))

		// Execute
		result, err := userService.Register(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "duplicate email")
		userStore.AssertExpectations(t)
	})

	t.Run("registration with invalid email", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		req := models.RegisterRequest{
			Email:       "invalid-email",
			Password:    "any-password", // Not validated due to email error
			DisplayName: "Test User",
		}

		// Execute
		result, err := userService.Register(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "validation error")
	})

	t.Run("registration with weak password", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		req := models.RegisterRequest{
			Email:       "test@example.com",
			Password:    "weak",
			DisplayName: "Test User",
		}

		// Execute
		result, err := userService.Register(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "validation error")
	})

	t.Run("registration with password missing uppercase", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		req := models.RegisterRequest{
			Email:       "test@example.com",
			Password:    "password123",
			DisplayName: "Test User",
		}

		// Execute
		result, err := userService.Register(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "validation error")
	})

	t.Run("registration with password missing digit", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		req := models.RegisterRequest{
			Email:       "test@example.com",
			Password:    "PasswordOnly",
			DisplayName: "Test User",
		}

		// Execute
		result, err := userService.Register(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "validation error")
	})
}

// Login Tests
func TestUserService_Login(t *testing.T) {
	ctx := context.Background()

	t.Run("successful login", func(t *testing.T) {
		// Setup
		userService, userStore, _, _, sessionStore, _ := setupUserService()

		req := models.LoginRequest{
			Email:    "test@example.com",
			Password: "Password123",
		}

		testUser := createTestUser()
		testUser.Email = req.Email
		testSession := createTestSession(testUser.ID)

		// Mock expectations: Login calls 3 methods in sequence
		userStore.On("GetByEmail", ctx, req.Email).Return(testUser, nil)
		sessionStore.On("Create", ctx, testUser.ID, mock.AnythingOfType("time.Duration")).Return(testSession, nil)
		userStore.On("UpdateLastLogin", ctx, testUser.ID).Return(nil)

		// Execute
		result, err := userService.Login(ctx, req)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testUser.ID, result.User.ID)
		assert.Equal(t, testSession.ID, result.SessionID)
		assert.Equal(t, testSession.ExpiresAt, result.ExpiresAt)
		userStore.AssertExpectations(t)
		sessionStore.AssertExpectations(t)
	})

	t.Run("login with invalid credentials", func(t *testing.T) {
		// Setup
		userService, userStore, _, _, _, _ := setupUserService()

		req := models.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		testUser := createTestUser()
		testUser.Email = req.Email

		// Mock expectation: Only GetByEmail should be called
		userStore.On("GetByEmail", ctx, req.Email).Return(testUser, nil)

		// Execute
		result, err := userService.Login(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid credentials")
		userStore.AssertExpectations(t)
	})

	t.Run("login with non-existent user", func(t *testing.T) {
		// Setup
		userService, userStore, _, _, _, _ := setupUserService()

		req := models.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "Password123",
		}

		// Mock expectation: GetByEmail returns error
		userStore.On("GetByEmail", ctx, req.Email).Return(nil, errors.New("user not found"))

		// Execute
		result, err := userService.Login(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		userStore.AssertExpectations(t)
	})

	t.Run("login with invalid email", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		req := models.LoginRequest{
			Email:    "invalid-email",
			Password: "any-password", // Not validated due to email error
		}

		// Execute
		result, err := userService.Login(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "validation error")
	})

	t.Run("login with empty password", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		req := models.LoginRequest{
			Email:    "test@example.com",
			Password: "",
		}

		// Execute
		result, err := userService.Login(ctx, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "validation error")
	})
}

// Email Verification Tests
func TestUserService_SendEmailVerification(t *testing.T) {
	ctx := context.Background()

	t.Run("successful email verification send", func(t *testing.T) {
		// Setup
		userService, userStore, evStore, _, _, emailService := setupUserService()

		testUser := createTestUser()
		testToken := &models.EmailVerificationToken{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Token:     "testverificationtoken123",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			CreatedAt: time.Now(),
		}

		// Mock expectations: SendEmailVerification calls 3 methods
		userStore.On("GetByID", ctx, testUser.ID).Return(testUser, nil)
		evStore.On("Create", ctx, testUser.ID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(testToken, nil)
		emailService.On("SendVerificationEmail", ctx, testUser.Email, testToken.Token, testUser.DisplayName).Return(nil)

		// Execute
		err := userService.SendEmailVerification(ctx, testUser.ID)
		// Get User (Mocked)
		// Verification Email token // not mocked
		// Store EmailToken in Database (Mocked)
		// Send email to user (Mocked)

		// Verify
		assert.NoError(t, err)
		userStore.AssertExpectations(t)
		evStore.AssertExpectations(t)
		emailService.AssertExpectations(t)
	})

	t.Run("send verification for non-existent user", func(t *testing.T) {
		// Setup
		userService, userStore, _, _, _, _ := setupUserService()

		userID := uuid.New()

		// Mock expectation: GetByID returns error
		userStore.On("GetByID", ctx, userID).Return(nil, errors.New("user not found"))

		// Execute
		err := userService.SendEmailVerification(ctx, userID)

		// Verify
		assert.Error(t, err)
		userStore.AssertExpectations(t)
	})
}

func TestUserService_VerifyEmail(t *testing.T) {
	ctx := context.Background()

	t.Run("successful email verification", func(t *testing.T) {
		// Setup
		userService, userStore, evStore, _, _, emailService := setupUserService()

		testUser := createTestUser()
		testToken := &models.EmailVerificationToken{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Token:     "validtokenfortest123",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			UsedAt:    nil,
			CreatedAt: time.Now(),
		}

		// Mock expectations: VerifyEmail calls 5 methods
		evStore.On("GetByToken", ctx, testToken.Token).Return(testToken, nil)
		userStore.On("GetByID", ctx, testUser.ID).Return(testUser, nil)
		evStore.On("MarkAsUsed", ctx, testToken.Token).Return(nil)
		userStore.On("VerifyUserEmail", ctx, testUser.ID).Return(nil)
		emailService.On("SendWelcomeEmail", ctx, testUser.Email, testUser.DisplayName).Return(nil)

		// Execute
		err := userService.VerifyEmail(ctx, testToken.Token)

		// Verify
		assert.NoError(t, err)
		evStore.AssertExpectations(t)
		userStore.AssertExpectations(t)
		emailService.AssertExpectations(t)
	})

	t.Run("verify with expired token", func(t *testing.T) {
		// Setup
		userService, _, evStore, _, _, _ := setupUserService()

		expiredToken := &models.EmailVerificationToken{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Token:     "expiredtokenfortest123",
			ExpiresAt: time.Now().Add(-10 * time.Minute), // Already expired
			UsedAt:    nil,
			CreatedAt: time.Now(),
		}

		// Mock expectation: Only GetByToken should be called
		evStore.On("GetByToken", ctx, expiredToken.Token).Return(expiredToken, nil)

		// Execute
		err := userService.VerifyEmail(ctx, expiredToken.Token)

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
		evStore.AssertExpectations(t)
	})

	t.Run("verify with already used token", func(t *testing.T) {
		// Setup
		userService, _, evStore, _, _, _ := setupUserService()

		usedAt := time.Now().Add(-1 * time.Hour)
		usedToken := &models.EmailVerificationToken{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Token:     "usedtokenfortest123",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			UsedAt:    &usedAt,
			CreatedAt: time.Now(),
		}

		// Mock expectation: Only GetByToken should be called
		evStore.On("GetByToken", ctx, usedToken.Token).Return(usedToken, nil)

		// Execute
		err := userService.VerifyEmail(ctx, usedToken.Token)

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already been used")
		evStore.AssertExpectations(t)
	})

	t.Run("verify with invalid token format", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		// Execute
		err := userService.VerifyEmail(ctx, "short")

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})

	t.Run("verify with empty token", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		// Execute
		err := userService.VerifyEmail(ctx, "")

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})

	t.Run("verify with whitespace-only token", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		// Execute
		err := userService.VerifyEmail(ctx, "   ")

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})
}

// Session Management Tests
func TestUserService_ValidateSession(t *testing.T) {
	ctx := context.Background()

	t.Run("valid session", func(t *testing.T) {
		// Setup
		userService, userStore, _, _, sessionStore, _ := setupUserService()

		testUser := createTestUser()
		testSession := createTestSession(testUser.ID)

		// Mock expectations: ValidateSession calls 2 methods
		sessionStore.On("Get", ctx, testSession.ID).Return(testSession, nil)
		userStore.On("GetByID", ctx, testUser.ID).Return(testUser, nil)

		// Execute
		result, err := userService.ValidateSession(ctx, testSession.ID)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testUser.ID, result.ID)
		sessionStore.AssertExpectations(t)
		userStore.AssertExpectations(t)
	})

	t.Run("expired session", func(t *testing.T) {
		// Setup
		userService, _, _, _, sessionStore, _ := setupUserService()

		expiredSession := createTestSession(uuid.New())
		expiredSession.ExpiresAt = time.Now().Add(-1 * time.Hour) // Already expired

		// Mock expectation: Only Get should be called
		sessionStore.On("Get", ctx, expiredSession.ID).Return(expiredSession, nil)

		// Execute
		result, err := userService.ValidateSession(ctx, expiredSession.ID)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "expired")
		sessionStore.AssertExpectations(t)
	})

	t.Run("revoked session", func(t *testing.T) {
		// Setup
		userService, _, _, _, sessionStore, _ := setupUserService()

		revokedAt := time.Now().Add(-1 * time.Hour)
		revokedSession := createTestSession(uuid.New())
		revokedSession.RevokedAt = &revokedAt

		// Mock expectation: Only Get should be called
		sessionStore.On("Get", ctx, revokedSession.ID).Return(revokedSession, nil)

		// Execute
		result, err := userService.ValidateSession(ctx, revokedSession.ID)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "revoked")
		sessionStore.AssertExpectations(t)
	})
}

// Password Reset Tests
func TestUserService_InitiatePasswordReset(t *testing.T) {
	ctx := context.Background()

	t.Run("successful password reset initiation", func(t *testing.T) {
		// Setup
		userService, userStore, _, prStore, _, emailService := setupUserService()

		testUser := createTestUser()
		testToken := &models.PasswordResetToken{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Token:     "resettoken123456789",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			CreatedAt: time.Now(),
		}

		// Mock expectations: InitiatePasswordReset calls 3 methods
		userStore.On("GetByEmail", ctx, testUser.Email).Return(testUser, nil)
		prStore.On("Create", ctx, testUser.ID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(testToken, nil)
		emailService.On("SendPasswordResetEmail", ctx, testUser.Email, mock.AnythingOfType("string"), testUser.DisplayName).Return(nil)

		// Execute
		err := userService.InitiatePasswordReset(ctx, testUser.Email)

		// Verify
		assert.NoError(t, err)
		userStore.AssertExpectations(t)
		prStore.AssertExpectations(t)
		emailService.AssertExpectations(t)
	})

	t.Run("password reset for non-existent user", func(t *testing.T) {
		// Setup
		userService, userStore, _, _, _, _ := setupUserService()

		// Mock expectation: GetByEmail returns error
		userStore.On("GetByEmail", ctx, "nonexistent@example.com").Return(nil, errors.New("user not found"))

		// Execute
		err := userService.InitiatePasswordReset(ctx, "nonexistent@example.com")

		// Verify - Should not return error for security (don't reveal if email exists)
		assert.NoError(t, err)
		userStore.AssertExpectations(t)
	})

	t.Run("password reset with invalid email", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		// Execute
		err := userService.InitiatePasswordReset(ctx, "invalid-email")

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})
}

func TestUserService_ResetPassword(t *testing.T) {
	ctx := context.Background()

	t.Run("successful password reset", func(t *testing.T) {
		// Setup
		userService, userStore, _, prStore, sessionStore, emailService := setupUserService()

		testUser := createTestUser()
		testToken := &models.PasswordResetToken{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Token:     "validresettokentest123",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			UsedAt:    nil,
			CreatedAt: time.Now(),
		}

		// Mock expectations: ResetPassword calls 6 methods
		prStore.On("GetByToken", ctx, testToken.Token).Return(testToken, nil)
		userStore.On("GetByID", ctx, testUser.ID).Return(testUser, nil)
		userStore.On("UpdatePasswordByID", ctx, testUser.ID, mock.AnythingOfType("string")).Return(nil)
		prStore.On("MarkAsUsed", ctx, testToken.Token).Return(nil)
		sessionStore.On("RevokeAllSessions", ctx, testUser.ID).Return(nil)
		emailService.On("SendPasswordResetConfirmation", ctx, testUser.Email, testUser.DisplayName).Return(nil)

		// Execute
		err := userService.ResetPassword(ctx, testToken.Token, "NewPassword123")

		// Verify
		assert.NoError(t, err)
		prStore.AssertExpectations(t)
		userStore.AssertExpectations(t)
		sessionStore.AssertExpectations(t)
		emailService.AssertExpectations(t)
	})

	t.Run("reset with expired token", func(t *testing.T) {
		// Setup
		userService, _, _, prStore, _, _ := setupUserService()

		expiredToken := &models.PasswordResetToken{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Token:     "expiredresettokentest123",
			ExpiresAt: time.Now().Add(-10 * time.Minute), // Already expired
			UsedAt:    nil,
			CreatedAt: time.Now(),
		}

		// Mock expectation: Only GetByToken should be called
		prStore.On("GetByToken", ctx, expiredToken.Token).Return(expiredToken, nil)

		// Execute
		err := userService.ResetPassword(ctx, expiredToken.Token, "NewPassword123")

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
		prStore.AssertExpectations(t)
	})

	t.Run("reset with already used token", func(t *testing.T) {
		// Setup
		userService, _, _, prStore, _, _ := setupUserService()

		usedAt := time.Now().Add(-1 * time.Hour)
		usedToken := &models.PasswordResetToken{
			ID:        uuid.New(),
			UserID:    uuid.New(),
			Token:     "usedresettokentest123",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			UsedAt:    &usedAt,
			CreatedAt: time.Now(),
		}

		// Mock expectation: Only GetByToken should be called
		prStore.On("GetByToken", ctx, usedToken.Token).Return(usedToken, nil)

		// Execute
		err := userService.ResetPassword(ctx, usedToken.Token, "NewPassword123")

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already been used")
		prStore.AssertExpectations(t)
	})

	t.Run("reset with invalid token format", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		// Execute
		err := userService.ResetPassword(ctx, "short", "any-password") // Not validated due to token error

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})

	t.Run("reset with weak new password", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		// Execute
		err := userService.ResetPassword(ctx, "validtokenfortest123", "weak")

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})
}

// User Management Tests
func TestUserService_UpdateProfile(t *testing.T) {
	ctx := context.Background()

	t.Run("successful profile update", func(t *testing.T) {
		// Setup
		userService, userStore, _, _, _, _ := setupUserService()

		testUser := createTestUser()
		updatedUser := createTestUser()
		updatedUser.ID = testUser.ID
		updatedUser.DisplayName = "Updated Name"
		updatedUser.AvatarURL = "https://example.com/avatar.jpg"

		req := models.UpdateProfileRequest{
			DisplayName: "Updated Name",
			AvatarURL:   "https://example.com/avatar.jpg",
		}

		// Mock expectations: UpdateProfile calls 3 methods
		userStore.On("GetByID", ctx, testUser.ID).Return(testUser, nil).Once()
		userStore.On("UpdateProfileByID", ctx, testUser.ID, req.DisplayName, req.AvatarURL).Return(nil)
		userStore.On("GetByID", ctx, testUser.ID).Return(updatedUser, nil).Once()

		// Execute
		result, err := userService.UpdateProfile(ctx, testUser.ID, req)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, updatedUser.DisplayName, result.DisplayName)
		assert.Equal(t, updatedUser.AvatarURL, result.AvatarURL)
		userStore.AssertExpectations(t)
	})

	t.Run("update profile for non-existent user", func(t *testing.T) {
		// Setup
		userService, userStore, _, _, _, _ := setupUserService()

		userID := uuid.New()
		req := models.UpdateProfileRequest{
			DisplayName: "Valid Name",
			AvatarURL:   "https://example.com/avatar.jpg",
		}

		// Mock expectation: GetByID returns error
		userStore.On("GetByID", ctx, userID).Return(nil, errors.New("user not found"))

		// Execute
		result, err := userService.UpdateProfile(ctx, userID, req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		userStore.AssertExpectations(t)
	})

	t.Run("update profile with invalid display name", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		req := models.UpdateProfileRequest{
			DisplayName: "Invalid@Name#With$Special%Characters",
			AvatarURL:   "https://example.com/avatar.jpg", // Valid to isolate display name error
		}

		// Execute
		result, err := userService.UpdateProfile(ctx, uuid.New(), req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "validation error")
	})

	t.Run("update profile with invalid avatar URL", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		req := models.UpdateProfileRequest{
			DisplayName: "Valid Name",
			AvatarURL:   "invalid-url",
		}

		// Execute
		result, err := userService.UpdateProfile(ctx, uuid.New(), req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "validation error")
	})

	t.Run("update profile with empty display name", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		req := models.UpdateProfileRequest{
			DisplayName: "",
			AvatarURL:   "https://example.com/avatar.jpg",
		}

		// Execute
		result, err := userService.UpdateProfile(ctx, uuid.New(), req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "validation error")
	})

	t.Run("update profile with empty avatar URL", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		req := models.UpdateProfileRequest{
			DisplayName: "Valid Name",
			AvatarURL:   "",
		}

		// Execute
		result, err := userService.UpdateProfile(ctx, uuid.New(), req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "validation error")
	})

	t.Run("update profile with whitespace-only display name", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		req := models.UpdateProfileRequest{
			DisplayName: "   ",
			AvatarURL:   "https://example.com/avatar.jpg",
		}

		// Execute
		result, err := userService.UpdateProfile(ctx, uuid.New(), req)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "validation error")
	})
}

// GetUserByEmail Tests
func TestUserService_GetUserByEmail(t *testing.T) {
	ctx := context.Background()

	t.Run("successful user retrieval by email", func(t *testing.T) {
		// Setup
		userService, userStore, _, _, _, _ := setupUserService()

		testUser := createTestUser()
		email := "test@example.com"

		// Mock expectation
		userStore.On("GetByEmail", ctx, email).Return(testUser, nil)

		// Execute
		result, err := userService.GetUserByEmail(ctx, email)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, testUser.ID, result.ID)
		userStore.AssertExpectations(t)
	})

	t.Run("get user with invalid email", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		// Execute
		result, err := userService.GetUserByEmail(ctx, "invalid-email")

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "validation error")
	})
}

// ResendVerification Tests
func TestUserService_ResendVerification(t *testing.T) {
	ctx := context.Background()

	t.Run("successful resend verification", func(t *testing.T) {
		// Setup
		userService, userStore, evStore, _, _, emailService := setupUserService()

		testUser := createTestUser()
		testUser.EmailVerified = false // Make sure user is not verified
		email := "test@example.com"
		testToken := &models.EmailVerificationToken{
			ID:        uuid.New(),
			UserID:    testUser.ID,
			Token:     "resendtokenfortest123",
			ExpiresAt: time.Now().Add(10 * time.Minute),
			CreatedAt: time.Now(),
		}

		// Mock expectations
		userStore.On("GetByEmail", ctx, email).Return(testUser, nil)
		evStore.On("Create", ctx, testUser.ID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).Return(testToken, nil)
		emailService.On("SendVerificationEmail", ctx, email, mock.AnythingOfType("string"), testUser.DisplayName).Return(nil)

		// Execute
		err := userService.ResendVerification(ctx, email)

		// Verify
		assert.NoError(t, err)
		userStore.AssertExpectations(t)
		evStore.AssertExpectations(t)
		emailService.AssertExpectations(t)
	})

	t.Run("resend verification with invalid email", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		// Execute
		err := userService.ResendVerification(ctx, "invalid-email")

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})

	t.Run("resend verification with whitespace-only email", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		// Execute
		err := userService.ResendVerification(ctx, "   ")

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})
}

// Change Password Tests
func TestUserService_ChangePassword(t *testing.T) {
	ctx := context.Background()

	t.Run("successful password change", func(t *testing.T) {
		// Setup
		userService, userStore, _, _, sessionStore, _ := setupUserService()

		testUser := createTestUser()
		req := models.ChangePasswordRequest{
			CurrentPassword: "Password123",
			NewPassword:     "NewPassword456",
		}

		// Mock expectations
		userStore.On("GetByID", ctx, testUser.ID).Return(testUser, nil)
		userStore.On("UpdatePasswordByID", ctx, testUser.ID, mock.AnythingOfType("string")).Return(nil)
		sessionStore.On("RevokeAllSessions", ctx, testUser.ID).Return(nil)

		// Execute
		err := userService.ChangePassword(ctx, testUser.ID, req)

		// Verify
		assert.NoError(t, err)
		userStore.AssertExpectations(t)
		sessionStore.AssertExpectations(t)
	})

	t.Run("change password with weak new password", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		req := models.ChangePasswordRequest{
			CurrentPassword: "any-current", // Not validated due to new password error
			NewPassword:     "weak",
		}

		// Execute
		err := userService.ChangePassword(ctx, uuid.New(), req)

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})

	t.Run("change password with empty current password", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		req := models.ChangePasswordRequest{
			CurrentPassword: "",
			NewPassword:     "any-new-password", // Not validated due to current password error
		}

		// Execute
		err := userService.ChangePassword(ctx, uuid.New(), req)

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})

	t.Run("change password with same passwords", func(t *testing.T) {
		// Setup
		userService, _, _, _, _, _ := setupUserService()

		req := models.ChangePasswordRequest{
			CurrentPassword: "same-password",
			NewPassword:     "same-password",
		}

		// Execute
		err := userService.ChangePassword(ctx, uuid.New(), req)

		// Verify
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation error")
	})
}

// Logout Tests
func TestUserService_Logout(t *testing.T) {
	ctx := context.Background()

	t.Run("successful logout", func(t *testing.T) {
		// Setup
		userService, _, _, _, sessionStore, _ := setupUserService()

		sessionID := uuid.New().String()

		// Mock expectation: Logout calls Revoke
		sessionStore.On("Revoke", ctx, sessionID).Return(nil)

		// Execute
		err := userService.Logout(ctx, sessionID)

		// Verify
		assert.NoError(t, err)
		sessionStore.AssertExpectations(t)
	})

	t.Run("logout with invalid session", func(t *testing.T) {
		// Setup
		userService, _, _, _, sessionStore, _ := setupUserService()

		sessionID := "invalid-session-id"

		// Mock expectation: Revoke returns error
		sessionStore.On("Revoke", ctx, sessionID).Return(errors.New("session not found"))

		// Execute
		err := userService.Logout(ctx, sessionID)

		// Verify
		assert.Error(t, err)
		sessionStore.AssertExpectations(t)
	})
}

