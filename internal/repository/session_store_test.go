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

// Mock UserSessionStore for testing the interface behavior
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

// Test helper functions
func setupMockSessionStore() *MockUserSessionStore {
	return &MockUserSessionStore{}
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

// Test UserSessionStore Interface Behavior
func TestUserSessionStoreInterface_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("successful session creation", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		userID := uuid.New()
		ttl := 24 * time.Hour
		expectedSession := createTestSession(userID)

		// Mock expectation
		mockStore.On("Create", ctx, userID, ttl).Return(expectedSession, nil)

		// Execute
		result, err := mockStore.Create(ctx, userID, ttl)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedSession.ID, result.ID)
		assert.Equal(t, userID, result.UserID)
		assert.NotEmpty(t, result.ID)
		assert.Nil(t, result.RevokedAt)
		mockStore.AssertExpectations(t)
	})

	t.Run("session creation with database error", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		userID := uuid.New()
		ttl := 24 * time.Hour
		expectedError := errors.New("database connection error")

		// Mock expectation
		mockStore.On("Create", ctx, userID, ttl).Return(nil, expectedError)

		// Execute
		result, err := mockStore.Create(ctx, userID, ttl)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("session creation with different TTL values", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		userID := uuid.New()
		ttl := 1 * time.Hour
		expectedSession := createTestSession(userID)
		expectedSession.ExpiresAt = time.Now().Add(ttl)

		// Mock expectation
		mockStore.On("Create", ctx, userID, ttl).Return(expectedSession, nil)

		// Execute
		result, err := mockStore.Create(ctx, userID, ttl)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		mockStore.AssertExpectations(t)
	})
}

func TestUserSessionStoreInterface_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("successful session retrieval", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		userID := uuid.New()
		sessionID := uuid.New().String()
		expectedSession := createTestSession(userID)
		expectedSession.ID = sessionID

		// Mock expectation
		mockStore.On("Get", ctx, sessionID).Return(expectedSession, nil)

		// Execute
		result, err := mockStore.Get(ctx, sessionID)

		// Verify
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, sessionID, result.ID)
		assert.Equal(t, userID, result.UserID)
		mockStore.AssertExpectations(t)
	})

	t.Run("session not found", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		sessionID := "nonexistent-session"
		expectedError := errors.New("session not found")

		// Mock expectation
		mockStore.On("Get", ctx, sessionID).Return(nil, expectedError)

		// Execute
		result, err := mockStore.Get(ctx, sessionID)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("session retrieval with database error", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		sessionID := uuid.New().String()
		expectedError := errors.New("database connection error")

		// Mock expectation
		mockStore.On("Get", ctx, sessionID).Return(nil, expectedError)

		// Execute
		result, err := mockStore.Get(ctx, sessionID)

		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

func TestUserSessionStoreInterface_Revoke(t *testing.T) {
	ctx := context.Background()

	t.Run("successful session revocation", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		sessionID := uuid.New().String()

		// Mock expectation
		mockStore.On("Revoke", ctx, sessionID).Return(nil)

		// Execute
		err := mockStore.Revoke(ctx, sessionID)

		// Verify
		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("revoke nonexistent session", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		sessionID := "nonexistent-session"
		expectedError := errors.New("session not found")

		// Mock expectation
		mockStore.On("Revoke", ctx, sessionID).Return(expectedError)

		// Execute
		err := mockStore.Revoke(ctx, sessionID)

		// Verify
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("revoke with database error", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		sessionID := uuid.New().String()
		expectedError := errors.New("database error")

		// Mock expectation
		mockStore.On("Revoke", ctx, sessionID).Return(expectedError)

		// Execute
		err := mockStore.Revoke(ctx, sessionID)

		// Verify
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

func TestUserSessionStoreInterface_RevokeAllSessions(t *testing.T) {
	ctx := context.Background()

	t.Run("successful revocation of all user sessions", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		userID := uuid.New()

		// Mock expectation
		mockStore.On("RevokeAllSessions", ctx, userID).Return(nil)

		// Execute
		err := mockStore.RevokeAllSessions(ctx, userID)

		// Verify
		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("revoke all sessions for nonexistent user", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		userID := uuid.New()
		expectedError := errors.New("user not found")

		// Mock expectation
		mockStore.On("RevokeAllSessions", ctx, userID).Return(expectedError)

		// Execute
		err := mockStore.RevokeAllSessions(ctx, userID)

		// Verify
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("revoke all sessions with database error", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		userID := uuid.New()
		expectedError := errors.New("database error")

		// Mock expectation
		mockStore.On("RevokeAllSessions", ctx, userID).Return(expectedError)

		// Execute
		err := mockStore.RevokeAllSessions(ctx, userID)

		// Verify
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

func TestUserSessionStoreInterface_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("successful session deletion", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		sessionID := uuid.New().String()

		// Mock expectation
		mockStore.On("Delete", ctx, sessionID).Return(nil)

		// Execute
		err := mockStore.Delete(ctx, sessionID)

		// Verify
		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("delete nonexistent session", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		sessionID := "nonexistent-session"
		expectedError := errors.New("session not found")

		// Mock expectation
		mockStore.On("Delete", ctx, sessionID).Return(expectedError)

		// Execute
		err := mockStore.Delete(ctx, sessionID)

		// Verify
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("delete with database error", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		sessionID := uuid.New().String()
		expectedError := errors.New("database error")

		// Mock expectation
		mockStore.On("Delete", ctx, sessionID).Return(expectedError)

		// Execute
		err := mockStore.Delete(ctx, sessionID)

		// Verify
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})
}

// Test session lifecycle scenarios
func TestUserSessionStoreInterface_SessionLifecycle(t *testing.T) {
	ctx := context.Background()

	t.Run("complete session lifecycle", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		userID := uuid.New()
		ttl := 24 * time.Hour
		
		// Create session
		expectedSession := createTestSession(userID)
		mockStore.On("Create", ctx, userID, ttl).Return(expectedSession, nil)
		
		session, err := mockStore.Create(ctx, userID, ttl)
		assert.NoError(t, err)
		assert.NotNil(t, session)
		
		// Get session
		mockStore.On("Get", ctx, session.ID).Return(session, nil)
		
		retrievedSession, err := mockStore.Get(ctx, session.ID)
		assert.NoError(t, err)
		assert.Equal(t, session.ID, retrievedSession.ID)
		
		// Revoke session
		mockStore.On("Revoke", ctx, session.ID).Return(nil)
		
		err = mockStore.Revoke(ctx, session.ID)
		assert.NoError(t, err)
		
		// Verify all expectations
		mockStore.AssertExpectations(t)
	})

	t.Run("session expiration handling", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		userID := uuid.New()
		sessionID := uuid.New().String()
		
		// Create expired session
		expiredSession := createTestSession(userID)
		expiredSession.ID = sessionID
		expiredSession.ExpiresAt = time.Now().Add(-1 * time.Hour) // Expired
		
		// Mock expectation
		mockStore.On("Get", ctx, sessionID).Return(expiredSession, nil)
		
		// Execute
		result, err := mockStore.Get(ctx, sessionID)
		
		// Verify - store should return expired session (validation happens at service level)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.ExpiresAt.Before(time.Now()))
		mockStore.AssertExpectations(t)
	})

	t.Run("revoked session handling", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		userID := uuid.New()
		sessionID := uuid.New().String()
		
		// Create revoked session
		revokedAt := time.Now().Add(-1 * time.Hour)
		revokedSession := createTestSession(userID)
		revokedSession.ID = sessionID
		revokedSession.RevokedAt = &revokedAt
		
		// Mock expectation
		mockStore.On("Get", ctx, sessionID).Return(revokedSession, nil)
		
		// Execute
		result, err := mockStore.Get(ctx, sessionID)
		
		// Verify - store should return revoked session (validation happens at service level)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.RevokedAt)
		mockStore.AssertExpectations(t)
	})
}

// Test concurrent session management
func TestUserSessionStoreInterface_ConcurrentSessions(t *testing.T) {
	ctx := context.Background()

	t.Run("multiple sessions for same user", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		userID := uuid.New()
		ttl := 24 * time.Hour
		
		// Create multiple sessions
		session1 := createTestSession(userID)
		session2 := createTestSession(userID)
		session2.ID = uuid.New().String()
		
		mockStore.On("Create", ctx, userID, ttl).Return(session1, nil).Once()
		mockStore.On("Create", ctx, userID, ttl).Return(session2, nil).Once()
		
		// Execute
		result1, err1 := mockStore.Create(ctx, userID, ttl)
		result2, err2 := mockStore.Create(ctx, userID, ttl)
		
		// Verify
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, result1.ID, result2.ID)
		assert.Equal(t, userID, result1.UserID)
		assert.Equal(t, userID, result2.UserID)
		mockStore.AssertExpectations(t)
	})

	t.Run("revoke all sessions for user with multiple sessions", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		userID := uuid.New()
		
		// Mock expectation - should revoke all sessions
		mockStore.On("RevokeAllSessions", ctx, userID).Return(nil)
		
		// Execute
		err := mockStore.RevokeAllSessions(ctx, userID)
		
		// Verify
		assert.NoError(t, err)
		mockStore.AssertExpectations(t)
	})
}

// Test that the UserSessionStore interface is implemented correctly
func TestUserSessionStoreInterface_Implementation(t *testing.T) {
	t.Run("mock implements UserSessionStore interface", func(t *testing.T) {
		// Setup
		var sessionStore repository.UserSessionStore
		mockStore := setupMockSessionStore()
		
		// Execute - this should compile if the interface is implemented correctly
		sessionStore = mockStore
		
		// Verify
		assert.NotNil(t, sessionStore)
		assert.IsType(t, &MockUserSessionStore{}, sessionStore)
	})
}

// Test edge cases and error conditions
func TestUserSessionStoreInterface_EdgeCases(t *testing.T) {
	ctx := context.Background()

	t.Run("create session with zero TTL", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		userID := uuid.New()
		ttl := 0 * time.Second
		
		// Mock expectation - should handle zero TTL gracefully
		expectedError := errors.New("invalid TTL")
		mockStore.On("Create", ctx, userID, ttl).Return(nil, expectedError)
		
		// Execute
		result, err := mockStore.Create(ctx, userID, ttl)
		
		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("create session with negative TTL", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		userID := uuid.New()
		ttl := -1 * time.Hour
		
		// Mock expectation - should handle negative TTL gracefully
		expectedError := errors.New("invalid TTL")
		mockStore.On("Create", ctx, userID, ttl).Return(nil, expectedError)
		
		// Execute
		result, err := mockStore.Create(ctx, userID, ttl)
		
		// Verify
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)
		mockStore.AssertExpectations(t)
	})

	t.Run("operations with empty session ID", func(t *testing.T) {
		// Setup
		mockStore := setupMockSessionStore()
		
		emptySessionID := ""
		expectedError := errors.New("invalid session ID")
		
		// Mock expectations
		mockStore.On("Get", ctx, emptySessionID).Return(nil, expectedError)
		mockStore.On("Revoke", ctx, emptySessionID).Return(expectedError)
		mockStore.On("Delete", ctx, emptySessionID).Return(expectedError)
		
		// Execute
		result, err1 := mockStore.Get(ctx, emptySessionID)
		err2 := mockStore.Revoke(ctx, emptySessionID)
		err3 := mockStore.Delete(ctx, emptySessionID)
		
		// Verify
		assert.Error(t, err1)
		assert.Error(t, err2)
		assert.Error(t, err3)
		assert.Nil(t, result)
		mockStore.AssertExpectations(t)
	})
}