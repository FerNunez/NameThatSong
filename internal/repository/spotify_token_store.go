package repository

import (
	"context"
	"fmt"
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
	Create(ctx context.Context, user_id uuid.UUID, refresh_token, access_token, token_type, score string, expires_at time.Time) error
	Get(ctx context.Context, user_id uuid.UUID) (SpotifyToken, error)
	IsValid(ctx context.Context, user_id uuid.UUID) (bool, error)
	Update(ctx context.Context, user_id uuid.UUID, new_refresh_token string, expires_at time.Time) error
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

	fmt.Println("[Create] refresh_token", refresh_token)
	fmt.Println("[Create] refresh_token_crypte", refresh_token_crypted)

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
func (s *SQLSpotifyTokenStore) Get(ctx context.Context, user_id uuid.UUID) (SpotifyToken, error) {
	dbSpotifyToken, err := s.db.GetSpotifyTokenByID(ctx, user_id)
	if err != nil {
		return SpotifyToken{}, nil
	}

	refresh_token_decrypted, err := s.encryptor.Decrypt(dbSpotifyToken.RefreshToken)
	if err != nil {
		return SpotifyToken{}, err
	}

	fmt.Printf("[GET] Fetching refresh token: %v, encrypted: %v\n", dbSpotifyToken.RefreshToken, refresh_token_decrypted)

	access_token_decrypted, err := s.encryptor.Decrypt(dbSpotifyToken.AccessToken)
	if err != nil {
		return SpotifyToken{}, err
	}
	fmt.Printf("[GET] Fetching access token: %v, encrypted: %v\n", dbSpotifyToken.AccessToken, access_token_decrypted)

	return SpotifyToken{
		RefreshToken: refresh_token_decrypted,
		AccessToken:  access_token_decrypted,
		TokenType:    dbSpotifyToken.TokenType,
		Scope:        dbSpotifyToken.Scope,
		ExpiresAt:    dbSpotifyToken.ExpiresAt,
	}, nil

}
func (s *SQLSpotifyTokenStore) IsValid(ctx context.Context, user_id uuid.UUID) (bool, error) {

	dbSpotifyToken, err := s.Get(ctx, user_id)
	if err != nil {
		_ = fmt.Errorf("[IsValid]: user_id %v  not found", user_id)
		return false, err
	}
	if time.Now().After(dbSpotifyToken.ExpiresAt) {
		_ = fmt.Errorf("[IsValid]: AccessToken expired for user_id: %v", user_id)
		return false, nil
	}
	return true, nil
}

func (s *SQLSpotifyTokenStore) Update(ctx context.Context, user_id uuid.UUID, new_access_token string, expires_at time.Time) error {
	fmt.Printf("[SQLSpotifyTokenStore] Update: new access_token %v\n\n", new_access_token)

	access_token_crypted, err := s.encryptor.Encrypt(new_access_token)
	if err != nil {
		return err
	}
	fmt.Printf("[SQLSpotifyTokenStore] Update: new access_token encrypted %v added in db\n\n", access_token_crypted)

	return s.db.UpdateSpotifyAccessToken(ctx, database.UpdateSpotifyAccessTokenParams{
		AccessToken: access_token_crypted,
		ExpiresAt:   expires_at,
		UserID:      user_id,
	})
}
