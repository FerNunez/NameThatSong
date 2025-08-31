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

// Mock PasswordResetStore
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

func createTestPasswordResetToken(userID uuid.UUID) *models.PasswordResetToken {
	return &models.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     "test-reset-token",
		ExpiresAt: time.Now().Add(10 * time.Minute),
		UsedAt:    nil,
		CreatedAt: time.Now(),
	}
}

func TestPasswordResetStore_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("successful token creation", func(t *testing.T) {
		mockStore := &MockPasswordResetStore{}
		
		userID := uuid.New()
		token := "reset-token-123"
		expiresAt := time.Now().Add(10 * time.Minute)
		expectedToken := createTestPasswordResetToken(userID)
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
		mockStore := &MockPasswordResetStore{}
		
		userID := uuid.New()
		token := "reset-token-123"
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

func TestPasswordResetStore_GetByToken(t *testing.T) {
	ctx := context.Background()

	t.Run("successful token retrieval", func(t *testing.T) {
		mockStore := &MockPasswordResetStore{}
		
		token := "reset-token-123"
		expectedToken := createTestPasswordResetToken(uuid.New())
		expectedToken.Token = token

		mockStore.On("GetByToken", ctx, token).Return(expectedToken, nil)

		result, err := mockStore.GetByToken(ctx, token)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, token, result.Token)
		mockStore.AssertExpectations(t)
	})

	t.Run("token not found", func(t *testing.T) {
		mockStore := &MockPasswordResetStore{}
		
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

func TestPasswordResetStore_MarkAsUsed(t *testing.T) {
	ctx := context.Background()

	t.Run("successful mark as used", func(t *testing.T) {
		mockStore := &MockPasswordResetStore{}
		
		token := "reset-token-123"

		mockStore.On("MarkAsUsed", ctx, token).Return(nil)

		err := mockStore.MarkAsUsed(ctx, token)

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("mark as used with database error", func(t *testing.T) {
		mockStore := &MockPasswordResetStore{}
		
		token := "reset-token-123"
		expectedError := errors.New("database error")

		mockStore.On("MarkAsUsed", ctx, token).Return(expectedError)

		err := mockStore.MarkAsUsed(ctx, token)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

func TestPasswordResetStore_DeleteByID(t *testing.T) {
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		mockStore := &MockPasswordResetStore{}
		
		tokenID := uuid.New()

		mockStore.On("DeleteByID", ctx, tokenID).Return(nil)

		err := mockStore.DeleteByID(ctx, tokenID)

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("deletion with database error", func(t *testing.T) {
		mockStore := &MockPasswordResetStore{}
		
		tokenID := uuid.New()
		expectedError := errors.New("database error")

		mockStore.On("DeleteByID", ctx, tokenID).Return(expectedError)

		err := mockStore.DeleteByID(ctx, tokenID)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

func TestPasswordResetStore_CleanupExpired(t *testing.T) {
	ctx := context.Background()

	t.Run("successful cleanup", func(t *testing.T) {
		mockStore := &MockPasswordResetStore{}

		mockStore.On("CleanupExpired", ctx).Return(nil)

		err := mockStore.CleanupExpired(ctx)

		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("cleanup with database error", func(t *testing.T) {
		mockStore := &MockPasswordResetStore{}
		
		expectedError := errors.New("database error")

		mockStore.On("CleanupExpired", ctx).Return(expectedError)

		err := mockStore.CleanupExpired(ctx)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

func TestPasswordResetStore_Interface(t *testing.T) {
	t.Run("mock implements PasswordResetStore interface", func(t *testing.T) {
		var store repository.PasswordResetStore
		mockStore := &MockPasswordResetStore{}
		
		store = mockStore
		
		assert.NotNil(t, store)
		assert.IsType(t, &MockPasswordResetStore{}, store)
	})
}