package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock UserStore for testing the interface behavior
type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) Create(ctx context.Context, email, hashedPassword, displayName string) (*models.User, error) {
	args := m.Called(ctx, email, hashedPassword, displayName)
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

func (m *MockUserStore) UpdateSpotifyConnectionStatus(ctx context.Context, id uuid.UUID, connected bool) error {
	args := m.Called(ctx, id, connected)
	return args.Error(0)
}

// Test helper functions
func setupMockUserStore() *MockUserStore {
	return &MockUserStore{}
}

func createTestUser() *models.User {
	return &models.User{
		ID:             uuid.New(),
		Email:          "test@example.com",
		HashedPassword: "$2a$10$hashedpassword",
		DisplayName:    "Test User",
		AvatarURL:      "",
		EmailVerified:  false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastLoginAt:    nil,
	}
}

// Test UserStore Interface Behavior
func TestUserStoreInterface_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("successful user creation", func(t *testing.T) {
		// Setup
		mockStore := setupMockUserStore()
		
		email := "test@example.com"
		hashedPassword := "$2a$10$hashedpassword"
		expectedUser := createTestUser()
		expectedUser.Email = email
		expectedUser.HashedPassword = hashedPassword

		// Mock expectation
		mockStore.On("Create", ctx, email, hashedPassword, "Test Display Name").Return(expectedUser, nil)

		// Execute
		result, err := mockStore.Create(ctx, email, hashedPassword, "Test Display Name")

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, email, result.Email)
		assert.Equal(t, hashedPassword, result.HashedPassword)
		assert.Equal(t, expectedUser.ID, result.ID)
		mockStore.AssertExpectations(t)
	})

	t.Run("creation with database error", func(t *testing.T) {
		// Setup
		mockStore := setupMockUserStore()
		
		email := "test@example.com"
		hashedPassword := "$2a$10$hashedpassword"
		expectedError := errors.New("database connection error")

		// Mock expectation
		mockStore.On("Create", ctx, email, hashedPassword, "Test Display Name").Return(nil, expectedError)

		// Execute
		result, err := mockStore.Create(ctx, email, hashedPassword, "Test Display Name")

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

func TestUserStoreInterface_GetByEmail(t *testing.T) {
	ctx := context.Background()

	t.Run("successful user retrieval by email", func(t *testing.T) {
		// Setup
		mockStore := setupMockUserStore()
		
		email := "test@example.com"
		expectedUser := createTestUser()
		expectedUser.Email = email

		// Mock expectation
		mockStore.On("GetByEmail", ctx, email).Return(expectedUser, nil)

		// Execute
		result, err := mockStore.GetByEmail(ctx, email)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, email, result.Email)
		assert.Equal(t, expectedUser.ID, result.ID)
		mockStore.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		// Setup
		mockStore := setupMockUserStore()
		
		email := "nonexistent@example.com"
		expectedError := errors.New("user not found")

		// Mock expectation
		mockStore.On("GetByEmail", ctx, email).Return(nil, expectedError)

		// Execute
		result, err := mockStore.GetByEmail(ctx, email)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

func TestUserStoreInterface_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("successful user retrieval by ID", func(t *testing.T) {
		// Setup
		mockStore := setupMockUserStore()
		
		userID := uuid.New()
		expectedUser := createTestUser()
		expectedUser.ID = userID

		// Mock expectation
		mockStore.On("GetByID", ctx, userID).Return(expectedUser, nil)

		// Execute
		result, err := mockStore.GetByID(ctx, userID)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.ID)
		mockStore.AssertExpectations(t)
	})

	t.Run("user not found by ID", func(t *testing.T) {
		// Setup
		mockStore := setupMockUserStore()
		
		userID := uuid.New()
		expectedError := errors.New("user not found")

		// Mock expectation
		mockStore.On("GetByID", ctx, userID).Return(nil, expectedError)

		// Execute
		result, err := mockStore.GetByID(ctx, userID)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

func TestUserStoreInterface_UpdatePasswordByID(t *testing.T) {
	ctx := context.Background()

	t.Run("successful password update", func(t *testing.T) {
		// Setup
		mockStore := setupMockUserStore()
		
		userID := uuid.New()
		newHashedPassword := "$2a$10$newhash"

		// Mock expectation
		mockStore.On("UpdatePasswordByID", ctx, userID, newHashedPassword).Return(nil)

		// Execute
		err := mockStore.UpdatePasswordByID(ctx, userID, newHashedPassword)

		// Verify
		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("password update with database error", func(t *testing.T) {
		// Setup
		mockStore := setupMockUserStore()
		
		userID := uuid.New()
		newHashedPassword := "$2a$10$newhash"
		expectedError := errors.New("database error")

		// Mock expectation
		mockStore.On("UpdatePasswordByID", ctx, userID, newHashedPassword).Return(expectedError)

		// Execute
		err := mockStore.UpdatePasswordByID(ctx, userID, newHashedPassword)

		// Verify
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

func TestUserStoreInterface_UpdateProfileByID(t *testing.T) {
	ctx := context.Background()

	t.Run("successful profile update", func(t *testing.T) {
		// Setup
		mockStore := setupMockUserStore()
		
		userID := uuid.New()
		displayName := "Updated Name"
		avatarURL := "https://example.com/avatar.jpg"

		// Mock expectation
		mockStore.On("UpdateProfileByID", ctx, userID, displayName, avatarURL).Return(nil)

		// Execute
		err := mockStore.UpdateProfileByID(ctx, userID, displayName, avatarURL)

		// Verify
		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("profile update with database error", func(t *testing.T) {
		// Setup
		mockStore := setupMockUserStore()
		
		userID := uuid.New()
		displayName := "Updated Name"
		avatarURL := "https://example.com/avatar.jpg"
		expectedError := errors.New("database error")

		// Mock expectation
		mockStore.On("UpdateProfileByID", ctx, userID, displayName, avatarURL).Return(expectedError)

		// Execute
		err := mockStore.UpdateProfileByID(ctx, userID, displayName, avatarURL)

		// Verify
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

func TestUserStoreInterface_VerifyUserEmail(t *testing.T) {
	ctx := context.Background()

	t.Run("successful email verification", func(t *testing.T) {
		// Setup
		mockStore := setupMockUserStore()
		
		userID := uuid.New()

		// Mock expectation
		mockStore.On("VerifyUserEmail", ctx, userID).Return(nil)

		// Execute
		err := mockStore.VerifyUserEmail(ctx, userID)

		// Verify
		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("email verification with database error", func(t *testing.T) {
		// Setup
		mockStore := setupMockUserStore()
		
		userID := uuid.New()
		expectedError := errors.New("database error")

		// Mock expectation
		mockStore.On("VerifyUserEmail", ctx, userID).Return(expectedError)

		// Execute
		err := mockStore.VerifyUserEmail(ctx, userID)

		// Verify
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

func TestUserStoreInterface_UpdateLastLogin(t *testing.T) {
	ctx := context.Background()

	t.Run("successful last login update", func(t *testing.T) {
		// Setup
		mockStore := setupMockUserStore()
		
		userID := uuid.New()

		// Mock expectation
		mockStore.On("UpdateLastLogin", ctx, userID).Return(nil)

		// Execute
		err := mockStore.UpdateLastLogin(ctx, userID)

		// Verify
		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})
}

func TestUserStoreInterface_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("successful user deletion", func(t *testing.T) {
		// Setup
		mockStore := setupMockUserStore()
		
		userID := uuid.New()

		// Mock expectation
		mockStore.On("Delete", ctx, userID).Return(nil)

		// Execute
		err := mockStore.Delete(ctx, userID)

		// Verify
		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("deletion with database error", func(t *testing.T) {
		// Setup
		mockStore := setupMockUserStore()
		
		userID := uuid.New()
		expectedError := errors.New("database error")

		// Mock expectation
		mockStore.On("Delete", ctx, userID).Return(expectedError)

		// Execute
		err := mockStore.Delete(ctx, userID)

		// Verify
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

func TestUserStoreInterface_Reset(t *testing.T) {
	ctx := context.Background()

	t.Run("successful reset", func(t *testing.T) {
		// Setup
		mockStore := setupMockUserStore()

		// Mock expectation
		mockStore.On("Reset", ctx).Return(nil)

		// Execute
		err := mockStore.Reset(ctx)

		// Verify
		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("reset with database error", func(t *testing.T) {
		// Setup
		mockStore := setupMockUserStore()
		
		expectedError := errors.New("database error")

		// Mock expectation
		mockStore.On("Reset", ctx).Return(expectedError)

		// Execute
		err := mockStore.Reset(ctx)

		// Verify
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

// Test that the UserStore interface is implemented correctly
func TestUserStoreInterface_Implementation(t *testing.T) {
	t.Run("mock implements UserStore interface", func(t *testing.T) {
		// Setup
		var userStore repository.UserStore
		mockStore := setupMockUserStore()
		
		// Execute - this should compile if the interface is implemented correctly
		userStore = mockStore
		
		// Verify
		assert.NotNil(t, userStore)
		assert.IsType(t, &MockUserStore{}, userStore)
	})
}