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

	"github.com/FerNunez/NameThatSong/internal/config"
	"github.com/FerNunez/NameThatSong/internal/repository"
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
}

// NewSpotifyAuthService creates a new Spotify auth service
func NewSpotifyAuthService(config *config.SpotifyConfig, tokenStore repository.SpotifyTokenStore) *SpotifyAuthService {
	return &SpotifyAuthService{
		config:     config,
		tokenStore: tokenStore,
	}
}

// AuthRequestURL generates the Spotify authorization URL
func (s *SpotifyAuthService) AuthRequestURL(state string) (string, error) {
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

// ValidateState validates the OAuth state parameter
func (s *SpotifyAuthService) ValidateState(receivedState, expectedState string) error {
	if receivedState != expectedState {
		return errors.New("state validation failed")
	}
	return nil
}

// TokenExchange exchanges authorization code for access tokens
func (s *SpotifyAuthService) TokenExchange(code string) (TokenResponse, error) {
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
func (s *SpotifyAuthService) RefreshToken(refreshToken string) (TokenResponse, error) {
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

// StoreTokens stores user tokens in the database
func (s *SpotifyAuthService) StoreTokens(ctx context.Context, userID string, tokens TokenResponse) error {
	return s.tokenStore.StoreTokens(ctx, userID, tokens.AccessToken, tokens.RefreshToken, tokens.ExpiresIn)
}

// GetValidAccessToken retrieves a valid access token for the user (refreshing if necessary)
func (s *SpotifyAuthService) GetValidAccessToken(ctx context.Context, userID string) (string, error) {
	// Try to get valid token from store
	accessToken, err := s.tokenStore.GetValidAccessToken(ctx, userID)
	if err == nil {
		return accessToken, nil
	}

	// If token is expired, try to refresh it
	userUUID, parseErr := uuid.Parse(userID)
	if parseErr != nil {
		return "", fmt.Errorf("invalid user ID: %v", parseErr)
	}

	// Get the refresh token
	token, getErr := s.tokenStore.Get(ctx, userUUID)
	if getErr != nil {
		return "", fmt.Errorf("failed to get refresh token: %v", getErr)
	}

	// Use the existing refresh token logic
	newTokens, refreshErr := s.RefreshToken(token.RefreshToken)
	if refreshErr != nil {
		return "", fmt.Errorf("failed to refresh token: %v", refreshErr)
	}

	// Store the new tokens
	storeErr := s.StoreTokens(ctx, userID, newTokens)
	if storeErr != nil {
		return "", fmt.Errorf("failed to store refreshed tokens: %v", storeErr)
	}

	return newTokens.AccessToken, nil
}

