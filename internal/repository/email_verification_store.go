package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/repository/database"
	"github.com/google/uuid"
)

type EmailVerificationStore interface {
	Create(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (*models.EmailVerificationToken, error)
	GetByToken(ctx context.Context, token string) (*models.EmailVerificationToken, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.EmailVerificationToken, error)
	MarkAsUsed(ctx context.Context, token string) error
	CleanupExpired(ctx context.Context) error
}

type SQLEmailVerificationStore struct {
	db *database.Queries
}

func NewSQLEmailVerificationStore(db *database.Queries) EmailVerificationStore {
	return &SQLEmailVerificationStore{db: db}
}

func (s *SQLEmailVerificationStore) Create(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (*models.EmailVerificationToken, error) {
	dbEvt, err := s.db.CreateEmailVerificationTokens(ctx, database.CreateEmailVerificationTokensParams{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, err
	}
	return fromDbEVT(&dbEvt), nil
}
func (s *SQLEmailVerificationStore) GetByToken(ctx context.Context, token string) (*models.EmailVerificationToken, error) {
	dbEvt, err := s.db.GetEmailVerificationTokensByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return fromDbEVT(&dbEvt), nil
}
func (s *SQLEmailVerificationStore) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.EmailVerificationToken, error) {
	dbEvt, err := s.db.GetEmailVerificationTokensByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return fromDbEVT(&dbEvt), nil
}
func (s *SQLEmailVerificationStore) MarkAsUsed(ctx context.Context, token string) error {
	dbEvt, err := s.db.UpdateEmailVerificationTokensUsedAtByToken(ctx, token)
	if err != nil {
		return err
	}
	if !dbEvt.UsedAt.Valid {
		return fmt.Errorf("could not set EmailVerificationTokens as UsedAt")
	}
	return nil
}
func (s *SQLEmailVerificationStore) CleanupExpired(ctx context.Context) error {
	return s.db.CleanupExpiredEmailVerificationTokens(ctx)
}

func fromDbEVT(dbEvt *database.EmailVerificationToken) *models.EmailVerificationToken {
	return &models.EmailVerificationToken{
		ID:        dbEvt.ID,
		UserID:    dbEvt.UserID,
		Token:     dbEvt.Token,
		ExpiresAt: dbEvt.ExpiresAt,
		UsedAt:    timeFromNullTime(dbEvt.UsedAt),
		CreatedAt: dbEvt.CreatedAt,
	}
}

func timeFromNullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}
