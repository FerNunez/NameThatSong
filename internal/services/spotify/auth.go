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

	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
	"github.com/FerNunez/NameThatSong/internal/pkg/utils"
	"github.com/FerNunez/NameThatSong/internal/repository"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// =============================================================================
// AUTHENTICATION METHODS
// =============================================================================

// AuthRequestURL generates the Spotify authorization URL with internally managed state
func (s *Spotify) AuthRequestURL(userID string) (string, error) {
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
func (s *Spotify) TokenExchange(ctx context.Context, userID, code, receivedState string) (TokenResponse, error) {
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

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.GetTokenURL(), strings.NewReader(data.Encode()))
	if err != nil {
		return TokenResponse{}, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	auth := base64.StdEncoding.EncodeToString([]byte(s.config.ClientID + ":" + s.config.ClientSecret))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("error getting token: %v", err)
	}
	defer resp.Body.Close()

	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return TokenResponse{}, fmt.Errorf("error parsing token response: %v", err)
	}

	// Save in db
	token := repository.SpotifyToken{
		RefreshToken: tokenResponse.RefreshToken,
		AccessToken:  tokenResponse.AccessToken,
		TokenType:    tokenResponse.TokenType,
		Scope:        tokenResponse.Scope,
		ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
	}
	err = s.tokenStore.Create(context.Background(), userID, token)
	if err != nil {
		return TokenResponse{}, fmt.Errorf("could not create token in db: %v", err)
	}
	return tokenResponse, nil
}

// RefreshToken regenerates access token from refresh token
func (s *Spotify) refreshToken(ctx context.Context, refreshToken string) (TokenResponse, error) {
	logger.Debug(ctx, "making token refresh request to Spotify",
		logger.F("has_refresh_token", refreshToken != ""))
	
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", s.config.ClientID)

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.GetTokenURL(), strings.NewReader(data.Encode()))
	if err != nil {
		logger.Error(ctx, "failed to create token refresh request",
			logger.F("error", err))
		return TokenResponse{}, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	auth := base64.StdEncoding.EncodeToString([]byte(s.config.ClientID + ":" + s.config.ClientSecret))
	req.Header.Set("Authorization", "Basic "+auth)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error(ctx, "HTTP request to Spotify token endpoint failed",
			logger.F("error", err))
		return TokenResponse{}, fmt.Errorf("error refreshing token: %v", err)
	}
	defer resp.Body.Close()

	logger.Debug(ctx, "received token refresh response from Spotify",
		logger.F("status_code", resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		logger.Error(ctx, "Spotify token refresh returned non-200 status",
			logger.F("status_code", resp.StatusCode))
		return TokenResponse{}, fmt.Errorf("token refresh failed with status code: %d", resp.StatusCode)
	}

	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		logger.Error(ctx, "failed to parse token refresh response",
			logger.F("error", err))
		return TokenResponse{}, fmt.Errorf("error parsing token response: %v", err)
	}

	logger.Debug(ctx, "token refresh response parsed successfully",
		logger.F("token_type", tokenResponse.TokenType),
		logger.F("expires_in", tokenResponse.ExpiresIn),
		logger.F("has_new_refresh_token", tokenResponse.RefreshToken != ""))

	return tokenResponse, nil
}

// GetValidToken retrieves a valid access token for the user (refreshing if necessary)
func (s *Spotify) GetValidToken(ctx context.Context, userID string) (string, error) {
	logger.Debug(ctx, "retrieving valid token for user",
		logger.F("user_id", userID))
	
	// Get token from storage
	token, err := s.tokenStore.Get(ctx, userID)
	if err != nil {
		logger.Error(ctx, "failed to retrieve token from storage",
			logger.F("user_id", userID),
			logger.F("error", err))
		return "", fmt.Errorf("no token found for user %s: %v", userID, err)
	}

	// Check if token is still valid
	now := time.Now()
	timeUntilExpiry := token.ExpiresAt.Sub(now)
	
	if now.Before(token.ExpiresAt) {
		logger.Debug(ctx, "using existing valid token",
			logger.F("user_id", userID),
			logger.F("expires_at", token.ExpiresAt),
			logger.F("time_until_expiry_minutes", int(timeUntilExpiry.Minutes())))
		return token.AccessToken, nil
	}

	// Token is expired, refresh it
	logger.Info(ctx, "token expired, attempting refresh",
		logger.F("user_id", userID),
		logger.F("expired_at", token.ExpiresAt),
		logger.F("expired_since_minutes", int(timeUntilExpiry.Abs().Minutes())))
	
	newTokens, err := s.refreshToken(ctx, token.RefreshToken)
	if err != nil {
		logger.Error(ctx, "failed to refresh token",
			logger.F("user_id", userID),
			logger.F("error", err))
		return "", fmt.Errorf("failed to refresh token: %v", err)
	}

	// Update the token in storage
	expiresAt := time.Now().Add(time.Duration(newTokens.ExpiresIn) * time.Second)
	err = s.tokenStore.Update(ctx, userID, newTokens.AccessToken, expiresAt)
	if err != nil {
		logger.Error(ctx, "failed to update refreshed token in storage",
			logger.F("user_id", userID),
			logger.F("error", err))
		return "", fmt.Errorf("failed to update token: %v", err)
	}

	logger.Info(ctx, "token refresh successful",
		logger.F("user_id", userID),
		logger.F("new_expires_at", expiresAt),
		logger.F("expires_in_minutes", newTokens.ExpiresIn/60))

	return newTokens.AccessToken, nil
}
