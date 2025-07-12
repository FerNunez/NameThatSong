package spotify

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/FerNunez/NameThatSong/internal/config"
	"github.com/FerNunez/NameThatSong/internal/pkg/utils"
	"github.com/FerNunez/NameThatSong/internal/repository"
	"github.com/FerNunez/NameThatSong/internal/services/cache"
	"github.com/google/uuid"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// SpotifyAuthService handles OAuth authentication with Spotify
type SpotifyAuthService struct {
	config     *config.SpotifyConfig
	tokenStore repository.SpotifyTokenStore
	cache      cache.SpotifyCache
}

// NewSpotifyAuthService creates a new Spotify auth service
func NewSpotifyAuthService(config *config.SpotifyConfig, tokenStore repository.SpotifyTokenStore, cache cache.SpotifyCache) *SpotifyAuthService {
	return &SpotifyAuthService{
		config:     config,
		tokenStore: tokenStore,
		cache:      cache,
	}
}

// AuthRequestURL generates the Spotify authorization URL with internally managed state
func (s *SpotifyAuthService) AuthRequestURL(userID string) (string, error) {
	// Generate a secure random state
	state, err := utils.GenerateState(16) // 16 bytes as example in spotify
	if err != nil {
		return "", fmt.Errorf("failed to generate OAuth state: %v", err)
	}

	// Store state in cache with 5 minute TTL
	s.cache.SetOAuthState(userID, state)

	u, err := url.Parse(s.config.GetAuthURL())
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", s.config.ClientID)
	q.Set("scope", "user-read-private user-read-email streaming user-modify-playback-state user-read-playback-state")
	q.Set("redirect_uri", s.config.RedirectURI)
	q.Set("state", state)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// TokenExchange exchanges authorization code for access tokens with state validation
func (s *SpotifyAuthService) TokenExchange(userID, code, receivedState string) (TokenResponse, error) {
	// Validate OAuth state first
	storedState, found := s.cache.GetOAuthState(userID)
	if !found {
		return TokenResponse{}, errors.New("OAuth state not found or expired")
	}

	if receivedState != storedState {
		return TokenResponse{}, errors.New("OAuth state validation failed")
	}

	// State is valid, proceed with token exchange
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", s.config.RedirectURI)

	req, err := http.NewRequest("POST", s.config.GetTokenURL(), strings.NewReader(data.Encode()))
	if err != nil {
		return TokenResponse{}, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	auth := base64.StdEncoding.EncodeToString([]byte(s.config.ClientID + ":" + s.config.ClientSecret))
	req.Header.Set("Authorization", "Basic "+auth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("error getting token: %v", err)
	}
	defer resp.Body.Close()

	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return TokenResponse{}, fmt.Errorf("error parsing token response: %v", err)
	}

	return tokenResponse, nil
}

// RefreshToken regenerates access token from refresh token
func (s *SpotifyAuthService) refreshToken(refreshToken string) (TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", s.config.ClientID)

	req, err := http.NewRequest("POST", s.config.GetTokenURL(), strings.NewReader(data.Encode()))
	if err != nil {
		return TokenResponse{}, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	auth := base64.StdEncoding.EncodeToString([]byte(s.config.ClientID + ":" + s.config.ClientSecret))
	req.Header.Set("Authorization", "Basic "+auth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("error refreshing token: %v", err)
	}
	defer resp.Body.Close()

	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return TokenResponse{}, fmt.Errorf("error parsing token response: %v", err)
	}

	return tokenResponse, nil
}

// GetValidToken retrieves a valid access token for the user (refreshing if necessary)
func (s *SpotifyAuthService) GetValidToken(ctx context.Context, userID string) (string, error) {
	// Get token from storage
	token, err := s.tokenStore.Get(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("no token found for user %s: %v", userID, err)
	}

	// Check if token is still valid
	if time.Now().Before(token.ExpiresAt) {
		return token.AccessToken, nil
	}

	// Token is expired, refresh it
	newTokens, err := s.refreshToken(token.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %v", err)
	}

	// Update the token in storage
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return "", fmt.Errorf("invalid user ID: %v", err)
	}

	expiresAt := time.Now().Add(time.Duration(newTokens.ExpiresIn) * time.Second)
	err = s.tokenStore.Update(ctx, userUUID, newTokens.AccessToken, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to update token: %v", err)
	}

	return newTokens.AccessToken, nil
}
