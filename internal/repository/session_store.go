package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/repository/database"
	"github.com/google/uuid"
)

type UserSessionStore interface {
	Create(ctx context.Context, userID uuid.UUID, ttl time.Duration) (*models.UserSession, error)
	Get(ctx context.Context, id string) (*models.UserSession, error)
	RevokeAllSessions(ctx context.Context, userID uuid.UUID) error
	Revoke(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
}

type SQLUserSessionStore struct {
	db *database.Queries
}

func NewSQLSessionStore(db *database.Queries) UserSessionStore {
	return &SQLUserSessionStore{
		db: db,
	}
}

func (s *SQLUserSessionStore) Delete(ctx context.Context, id string) error {
	return s.db.DeleteUserSessions(ctx, id)
}

func (s *SQLUserSessionStore) Create(ctx context.Context, userID uuid.UUID, ttl time.Duration) (*models.UserSession, error) {
	id := generateSessionID()
	expiresAt := time.Now().Add(ttl)

	dbSession, err := s.db.CreateSession(ctx, database.CreateSessionParams{
		ID:        id,
		UserID:    userID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return &models.UserSession{}, err
	}

	return &models.UserSession{
		ID:        dbSession.ID,
		UserID:    dbSession.UserID,
		CreatedAt: dbSession.CreatedAt,
		ExpiresAt: dbSession.ExpiresAt,
		RevokedAt: nullTimeToPointer(dbSession.RevokedAt),
	}, nil
}

func (s *SQLUserSessionStore) Get(ctx context.Context, id string) (*models.UserSession, error) {
	dbSession, err := s.db.GetSession(ctx, id)
	if err != nil {
		return nil, err
	}

	return &models.UserSession{
		ID:        dbSession.ID,
		UserID:    dbSession.UserID,
		CreatedAt: dbSession.CreatedAt,
		ExpiresAt: dbSession.ExpiresAt,
		RevokedAt: nullTimeToPointer(dbSession.RevokedAt),
	}, nil
}

func (s *SQLUserSessionStore) Revoke(ctx context.Context, id string) error {
	return s.db.Revoke(ctx, id)
}
func (s *SQLUserSessionStore) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	return s.db.RevokeUserSessionsByUserID(ctx, userID)
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
