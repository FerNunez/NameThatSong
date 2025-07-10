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
	Create(ctx context.Context, user_id uuid.UUID, refresh_token, access_token, token_type, scope string, expires_at time.Time) error
	Get(ctx context.Context, user_id uuid.UUID) (SpotifyToken, error)
	IsValid(ctx context.Context, user_id uuid.UUID) (bool, error)
	Update(ctx context.Context, user_id uuid.UUID, new_refresh_token string, expires_at time.Time) error
	GetValidAccessToken(ctx context.Context, userID string) (string, error)
	StoreTokens(ctx context.Context, userID, accessToken, refreshToken string, expiresIn int) error
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
func (s *SQLSpotifyTokenStore) Get(ctx context.Context, user_id uuid.UUID) (SpotifyToken, error) {
	dbSpotifyToken, err := s.db.GetSpotifyTokenByID(ctx, user_id)
	if err != nil {
		return SpotifyToken{}, err
	}

	refresh_token_decrypted, err := s.encryptor.Decrypt(dbSpotifyToken.RefreshToken)
	if err != nil {
		return SpotifyToken{}, err
	}

	// TODO: Add proper structured logging for token operations

	access_token_decrypted, err := s.encryptor.Decrypt(dbSpotifyToken.AccessToken)
	if err != nil {
		return SpotifyToken{}, err
	}
	// TODO: Add proper structured logging for access token operations

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

// GetValidAccessToken returns a valid access token from storage
func (s *SQLSpotifyTokenStore) GetValidAccessToken(ctx context.Context, userID string) (string, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return "", fmt.Errorf("invalid user ID: %v", err)
	}

	// Check if current token is valid
	isValid, err := s.IsValid(ctx, userUUID)
	if err != nil {
		return "", err
	}

	if isValid {
		// Return current access token
		token, err := s.Get(ctx, userUUID)
		if err != nil {
			return "", err
		}
		return token.AccessToken, nil
	}

	// Token is expired - return error to let the auth service handle refresh
	return "", fmt.Errorf("token expired for user %s", userID)
}

// StoreTokens stores tokens from a TokenResponse
func (s *SQLSpotifyTokenStore) StoreTokens(ctx context.Context, userID, accessToken, refreshToken string, expiresIn int) error {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %v", err)
	}

	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
	
	return s.Create(ctx, userUUID, refreshToken, accessToken, "Bearer", "user-read-private user-read-email streaming user-modify-playback-state user-read-playback-state", expiresAt)
}
