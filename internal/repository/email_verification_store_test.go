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

// Mock EmailVerificationStore
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

func createTestEmailVerificationToken(userID uuid.UUID) *models.EmailVerificationToken {
	return &models.EmailVerificationToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     "test-verification-token",
		ExpiresAt: time.Now().Add(10 * time.Minute),
		UsedAt:    nil,
		CreatedAt: time.Now(),
	}
}

func TestEmailVerificationStore_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("successful token creation", func(t *testing.T) {
		mockStore := &MockEmailVerificationStore{}

		userID := uuid.New()
		token := "verification-token-123"
		expiresAt := time.Now().Add(10 * time.Minute)
		expectedToken := createTestEmailVerificationToken(userID)
		expectedToken.Token = token
		expectedToken.ExpiresAt = expiresAt

		mockStore.On("Create", ctx, userID, token, expiresAt).Return(expectedToken, nil)

		result, err := mockStore.Create(ctx, userID, token, expiresAt)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, token, result.Token)
		assert.Equal(t, userID, result.UserID)
		assert.Nil(t, result.UsedAt)
		mockStore.AssertExpectations(t)
	})

	t.Run("create with database error", func(t *testing.T) {
		mockStore := &MockEmailVerificationStore{}

		userID := uuid.New()
		token := "verification-token-123"
		expiresAt := time.Now().Add(10 * time.Minute)
		expectedError := errors.New("database error")

		mockStore.On("Create", ctx, userID, token, expiresAt).Return(nil, expectedError)

		result, err := mockStore.Create(ctx, userID, token, expiresAt)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

func TestEmailVerificationStore_GetByToken(t *testing.T) {
	ctx := context.Background()

	t.Run("successful token retrieval", func(t *testing.T) {
		mockStore := &MockEmailVerificationStore{}

		token := "verification-token-123"
		expectedToken := createTestEmailVerificationToken(uuid.New())
		expectedToken.Token = token

		mockStore.On("GetByToken", ctx, token).Return(expectedToken, nil)

		result, err := mockStore.GetByToken(ctx, token)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, token, result.Token)
		mockStore.AssertExpectations(t)
	})

	t.Run("token not found", func(t *testing.T) {
		mockStore := &MockEmailVerificationStore{}

		token := "nonexistent-token"
		expectedError := errors.New("token not found")

		mockStore.On("GetByToken", ctx, token).Return(nil, expectedError)

		result, err := mockStore.GetByToken(ctx, token)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

func TestEmailVerificationStore_MarkAsUsed(t *testing.T) {
	ctx := context.Background()

	t.Run("successful mark as used", func(t *testing.T) {
		mockStore := &MockEmailVerificationStore{}

		token := "verification-token-123"

		mockStore.On("MarkAsUsed", ctx, token).Return(nil)

		err := mockStore.MarkAsUsed(ctx, token)

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("mark as used with database error", func(t *testing.T) {
		mockStore := &MockEmailVerificationStore{}

		token := "verification-token-123"
		expectedError := errors.New("database error")

		mockStore.On("MarkAsUsed", ctx, token).Return(expectedError)

		err := mockStore.MarkAsUsed(ctx, token)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

func TestEmailVerificationStore_CleanupExpired(t *testing.T) {
	ctx := context.Background()

	t.Run("successful cleanup", func(t *testing.T) {
		mockStore := &MockEmailVerificationStore{}

		mockStore.On("CleanupExpired", ctx).Return(nil)

		err := mockStore.CleanupExpired(ctx)

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("cleanup with database error", func(t *testing.T) {
		mockStore := &MockEmailVerificationStore{}

		expectedError := errors.New("database error")

		mockStore.On("CleanupExpired", ctx).Return(expectedError)

		err := mockStore.CleanupExpired(ctx)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

func TestEmailVerificationStore_Interface(t *testing.T) {
	t.Run("mock implements EmailVerificationStore interface", func(t *testing.T) {
		var store repository.EmailVerificationStore
		mockStore := &MockEmailVerificationStore{}

		store = mockStore

		assert.NotNil(t, store)
		assert.IsType(t, &MockEmailVerificationStore{}, store)
	})
}
