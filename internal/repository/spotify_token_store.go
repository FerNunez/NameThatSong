package repository

import (
	"context"
	"time"

	"github.com/FerNunez/NameThatSong/internal/pkg/utils"
	"github.com/FerNunez/NameThatSong/internal/repository/database"
	"github.com/google/uuid"
)

type SpotifyToken struct {
	RefreshToken string
	AccessToken  string
	TokenType    string
	Scope        string
	ExpiresAt    time.Time
}

type SpotifyTokenStore interface {
	// Core token operations
	Get(ctx context.Context, user_id string) (SpotifyToken, error)
	Create(ctx context.Context, user_id uuid.UUID, refresh_token, access_token, token_type, scope string, expires_at time.Time) error
	Update(ctx context.Context, user_id uuid.UUID, new_access_token string, expires_at time.Time) error
	Delete(ctx context.Context, user_id uuid.UUID) error
}

// ////////////////////////////////////////////
type SQLSpotifyTokenStore struct {
	db        *database.Queries
	encryptor *utils.TokenEncryptor
}

func NewSQLSpotifyTokenStore(db *database.Queries, encryptor *utils.TokenEncryptor) SpotifyTokenStore {
	return &SQLSpotifyTokenStore{db, encryptor}
}

func (s *SQLSpotifyTokenStore) Create(ctx context.Context, user_id uuid.UUID, refresh_token, access_token, token_type, scope string, expires_at time.Time) error {

	// Encrypt tokens!
	refresh_token_crypted, err := s.encryptor.Encrypt(refresh_token)
	if err != nil {
		return err
	}
	access_token_crypted, err := s.encryptor.Encrypt(access_token)
	if err != nil {
		return err
	}

	// TODO: Add proper structured logging instead of println

	// Store data
	_, err = s.db.CreateSpotifyToken(ctx, database.CreateSpotifyTokenParams{
		RefreshToken: refresh_token_crypted,
		AccessToken:  access_token_crypted,
		TokenType:    token_type,
		Scope:        scope,
		ExpiresAt:    expires_at,
		UserID:       user_id,
	})

	return err
}

// Returns decrypted SpotifyTOken for userId.
func (s *SQLSpotifyTokenStore) Get(ctx context.Context, userId string) (SpotifyToken, error) {
	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return SpotifyToken{}, err
	}

	dbSpotifyToken, err := s.db.GetSpotifyTokenByID(ctx, userUUID)
	if err != nil {
		return SpotifyToken{}, err
	}

	refresh_token_decrypted, err := s.encryptor.Decrypt(dbSpotifyToken.RefreshToken)
	if err != nil {
		return SpotifyToken{}, err
	}

	access_token_decrypted, err := s.encryptor.Decrypt(dbSpotifyToken.AccessToken)
	if err != nil {
		return SpotifyToken{}, err
	}

	return SpotifyToken{
		RefreshToken: refresh_token_decrypted,
		AccessToken:  access_token_decrypted,
		TokenType:    dbSpotifyToken.TokenType,
		Scope:        dbSpotifyToken.Scope,
		ExpiresAt:    dbSpotifyToken.ExpiresAt,
	}, nil

}

func (s *SQLSpotifyTokenStore) Update(ctx context.Context, user_id uuid.UUID, new_access_token string, expires_at time.Time) error {
	// TODO: Add proper structured logging for token updates
	access_token_crypted, err := s.encryptor.Encrypt(new_access_token)
	if err != nil {
		return err
	}
	// TODO: Log encrypted token storage operation

	return s.db.UpdateSpotifyAccessToken(ctx, database.UpdateSpotifyAccessTokenParams{
		AccessToken: access_token_crypted,
		ExpiresAt:   expires_at,
		UserID:      user_id,
	})
}


// Delete removes a token from storage
func (s *SQLSpotifyTokenStore) Delete(ctx context.Context, user_id uuid.UUID) error {
	return s.db.DeleteSpotifyToken(ctx, user_id)
}
