package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/repository/database"
	"github.com/google/uuid"
)

type UserStore interface {
	Create(ctx context.Context, email, hashed_password, displayName string) (*models.User, error)
	Delete(ctx context.Context, userID uuid.UUID) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	UpdatePasswordByID(ctx context.Context, id uuid.UUID, hashedPassword string) error
	UpdateProfileByID(ctx context.Context, id uuid.UUID, displayName string, avatarUrl string) error
	VerifyUserEmail(ctx context.Context, id uuid.UUID) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	UpdateSpotifyConnectionStatus(ctx context.Context, id uuid.UUID, connected bool) error
	Reset(ctx context.Context) error
}

type SQLUserStore struct {
	db *database.Queries
}

func NewSQLUserStore(db *database.Queries) UserStore {
	return &SQLUserStore{
		db: db,
	}
}

func (s *SQLUserStore) Create(ctx context.Context, email, hashed_password, displayName string) (*models.User, error) {
	dbUser, err := s.db.CreateUser(ctx, database.CreateUserParams{
		ID:             uuid.New(),
		Email:          email,
		HashedPassword: hashed_password,
		EmailVerified:  sql.NullBool{Bool: false, Valid: true},
		DisplayName:    displayName,
		AvatarUrl:      sql.NullString{String: "", Valid: false},
	})
	if err != nil {
		return &models.User{}, err
	}
	return &models.User{
		ID:               dbUser.ID,
		Email:            dbUser.Email,
		HashedPassword:   dbUser.HashedPassword,
		DisplayName:      "",
		AvatarURL:        "",
		EmailVerified:    false,
		SpotifyConnected: false,
		CreatedAt:        dbUser.CreatedAt,
		UpdatedAt:        dbUser.CreatedAt,
		LastLoginAt:      nil,
	}, nil
}

func (s *SQLUserStore) Delete(ctx context.Context, userID uuid.UUID) error {
	return s.db.DeleteUser(ctx, userID)
}

func (s SQLUserStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	dbUser, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		return &models.User{}, err
	}
	return fromDbUser(&dbUser), nil
}
func (s SQLUserStore) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	dbUser, err := s.db.GetUserByID(ctx, id)
	if err != nil {
		return &models.User{}, err
	}
	return fromDbUser(&dbUser), nil
}

func (s *SQLUserStore) VerifyUserEmail(ctx context.Context, id uuid.UUID) error {
	return s.db.VerifyUserEmail(ctx, database.VerifyUserEmailParams{
		ID:        id,
		UpdatedAt: time.Now(),
	})
}
func (s *SQLUserStore) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {

	return s.db.UpdateLastLogin(ctx, database.UpdateLastLoginParams{
		ID: id,
		LastLoginAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: time.Now(),
	})
}

func (s *SQLUserStore) UpdateSpotifyConnectionStatus(ctx context.Context, id uuid.UUID, connected bool) error {
	return s.db.UpdateSpotifyConnectionStatus(ctx, database.UpdateSpotifyConnectionStatusParams{
		ID:               id,
		SpotifyConnected: sql.NullBool{Bool: connected, Valid: true},
	})
}

func (s *SQLUserStore) UpdatePasswordByID(ctx context.Context, id uuid.UUID, hashedPass string) error {
	return s.db.UpdateUserPassword(ctx, database.UpdateUserPasswordParams{
		ID:             id,
		HashedPassword: hashedPass,
	})
}

func (s *SQLUserStore) UpdateProfileByID(ctx context.Context, id uuid.UUID, displayName string, avatarUrl string) error {
	var avatarURLns sql.NullString
	if avatarUrl == "" {
		avatarURLns.Valid = false
	} else {
		avatarURLns.Valid = true
		avatarURLns.String = avatarUrl
	}

	return s.db.UpdateUserProfile(ctx, database.UpdateUserProfileParams{
		ID:          id,
		DisplayName: displayName,
		AvatarUrl:   avatarURLns,
		UpdatedAt:   time.Now(),
	})
}

func (s *SQLUserStore) Reset(ctx context.Context) error {
	return s.db.ResetUsers(ctx)
}

func fromDbUser(dbUser *database.User) *models.User {
	var avatarURL string
	if !dbUser.AvatarUrl.Valid {
		avatarURL = ""
	} else {
		avatarURL = dbUser.AvatarUrl.String
	}
	var emailVerified bool
	if !dbUser.EmailVerified.Valid {
		emailVerified = false
	} else {
		emailVerified = dbUser.EmailVerified.Bool
	}
	var spotifyConnected bool
	if !dbUser.SpotifyConnected.Valid {
		spotifyConnected = false
	} else {
		spotifyConnected = dbUser.SpotifyConnected.Bool
	}
	var lastLoginAt *time.Time
	if !dbUser.LastLoginAt.Valid {
		lastLoginAt = nil
	} else {
		lastLoginAt = &dbUser.LastLoginAt.Time
	}
	return &models.User{
		ID:               dbUser.ID,
		Email:            dbUser.Email,
		HashedPassword:   dbUser.HashedPassword,
		DisplayName:      dbUser.DisplayName,
		AvatarURL:        avatarURL,
		EmailVerified:    emailVerified,
		SpotifyConnected: spotifyConnected,
		CreatedAt:        dbUser.CreatedAt,
		UpdatedAt:        dbUser.UpdatedAt,
		LastLoginAt:      lastLoginAt,
	}

}
