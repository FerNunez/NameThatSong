package repository

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/repository/database"
	"github.com/google/uuid"
)

type PasswordResetStore interface {
	Create(ctx context.Context, userID uuid.UUID, token string, expire_at time.Time, ipAddress net.IP, userAgent string) (*models.PasswordResetToken, error)
	GetByToken(ctx context.Context, token string) (*models.PasswordResetToken, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) (*models.PasswordResetToken, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
	MarkAsUsed(ctx context.Context, token string) error
	CleanupExpired(ctx context.Context) error
}

type SQLPasswordResetStore struct {
	db *database.Queries
}

func NewSQLPasswordResetStore(db *database.Queries) PasswordResetStore {
	return &SQLPasswordResetStore{db: db}
}

func (s *SQLPasswordResetStore) Create(ctx context.Context, userID uuid.UUID, token string, expire_at time.Time, ipAddress net.IP, userAgent string) (*models.PasswordResetToken, error) {
	// Convert userAgent to nullSqu
	var nullstring sql.NullString
	if userAgent == "" {
		nullstring.Valid = false
	} else {
		nullstring.Valid = true
		nullstring.String = userAgent
	}

	dbPRT, err := s.db.CreatePasswordResetTokens(ctx, database.CreatePasswordResetTokensParams{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     token,
		IpAddress: ipAddress,
		UserAgent: nullstring,
		ExpiresAt: expire_at,
	})
	if err != nil {
		return nil, err
	}
	return fromDbPRT(&dbPRT), nil
}
func (s *SQLPasswordResetStore) GetByToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	dbPRT, err := s.db.GetPasswordResetTokensByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return fromDbPRT(&dbPRT), nil
}
func (s *SQLPasswordResetStore) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.PasswordResetToken, error) {
	dbPRT, err := s.db.GetPasswordResetTokensByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return fromDbPRT(&dbPRT), nil
}
func (s *SQLPasswordResetStore) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return s.db.DeletePasswordResetTokens(ctx, id)
}
func (s *SQLPasswordResetStore) MarkAsUsed(ctx context.Context, token string) error {
	dbPRT, err := s.db.UpdatePasswordResetTokensUsedAtByToken(ctx, token)
	if err != nil {
		return err
	}
	if !dbPRT.UsedAt.Valid {
		return fmt.Errorf("could not set PasswordResetTokens as UsedAt")
	}
	return nil
}
func (s *SQLPasswordResetStore) CleanupExpired(ctx context.Context) error {
	return s.db.CleanupExpiredPasswordResetTokens(ctx)
}

func fromDbPRT(dbPrt *database.PasswordResetToken) *models.PasswordResetToken {
	return &models.PasswordResetToken{
		ID:        dbPrt.ID,
		UserID:    dbPrt.UserID,
		Token:     dbPrt.Token,
		ExpiresAt: dbPrt.ExpiresAt,
		UsedAt:    timeFromNullTime(dbPrt.UsedAt),
		IPAddress: dbPrt.IpAddress.String(),
		UserAgent: stringFromNullString(dbPrt.UserAgent),
		CreatedAt: dbPrt.CreatedAt,
	}
}

func stringFromNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "unknown"
}
