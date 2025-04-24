package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/FerNunez/NameThatSong/internal/store/database"
	"github.com/google/uuid"
)

type Session struct {
	ID        string
	UserID    uuid.UUID
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

type SessionStore interface {
	Create(ctx context.Context, userID uuid.UUID, ttl time.Duration) (Session, error)
	Get(ctx context.Context, id string) (Session, error)
	Revoke(ctx context.Context, id string) error
	IsValid(ctx context.Context, id string) (bool, error)
}

type SQLSessionStore struct {
	db *database.Queries
}

func NewSQLSessionStore(db *database.Queries) SessionStore {
	return &SQLSessionStore{
		db: db,
	}
}

func (s *SQLSessionStore) Create(ctx context.Context, userID uuid.UUID, ttl time.Duration) (Session, error) {
	id := generateSessionID() // You would implement this function
	expiresAt := time.Now().Add(ttl)

	dbSession, err := s.db.CreateSession(ctx, database.CreateSessionParams{
		ID:        id,
		UserID:    userID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return Session{}, err
	}

	// Convert from DB model to domain model
	return Session{
		ID:        dbSession.ID,
		UserID:    dbSession.UserID,
		CreatedAt: dbSession.CreatedAt,
		ExpiresAt: dbSession.ExpiresAt,
		RevokedAt: nullTimeToPointer(dbSession.RevokedAt),
	}, nil
}

// Get retrieves a session by ID
func (s *SQLSessionStore) Get(ctx context.Context, id string) (Session, error) {
	dbSession, err := s.db.GetSession(ctx, id)
	if err != nil {
		return Session{}, err
	}

	return Session{
		ID:        dbSession.ID,
		UserID:    dbSession.UserID,
		CreatedAt: dbSession.CreatedAt,
		ExpiresAt: dbSession.ExpiresAt,
		RevokedAt: nullTimeToPointer(dbSession.RevokedAt),
	}, nil
}

func (s *SQLSessionStore) Revoke(ctx context.Context, id string) error {
	now := time.Now()
	return s.db.UpdateSession(ctx, database.UpdateSessionParams{
		RevokedAt: sql.NullTime{Time: now, Valid: true},
		UpdatedAt: now,
		ID:        id,
	})
}

func (s *SQLSessionStore) IsValid(ctx context.Context, id string) (bool, error) {
	session, err := s.Get(ctx, id)
	if err != nil {
		return false, err
	}

	now := time.Now()

	// Check if session is expired
	if now.After(session.ExpiresAt) {
		return false, nil
	}

	// Check if session is revoked
	if session.RevokedAt != nil && now.After(*session.RevokedAt) {
		return false, nil
	}

	return true, nil
}

func nullTimeToPointer(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

func generateSessionID() string {
	return uuid.New().String()
}
