package config

import (
	"fmt"
	"os"
)

// SpotifyConfig holds all Spotify-related configuration
type SpotifyConfig struct {
	ClientID      string
	ClientSecret  string
	RedirectURI   string
	BaseURL       string
	AuthURL       string
	TokenURL      string
	EncryptionKey string
}

// NewSpotifyConfig creates a new Spotify configuration from environment variables
func NewSpotifyConfig() (*SpotifyConfig, error) {
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	if clientID == "" {
		return nil, fmt.Errorf("SPOTIFY_CLIENT_ID environment variable is required")
	}

	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	if clientSecret == "" {
		return nil, fmt.Errorf("SPOTIFY_CLIENT_SECRET environment variable is required")
	}

	redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")
	if redirectURI == "" {
		redirectURI = "http://localhost:8080/auth/callback" // Default fallback
	}

	encryptionKey := os.Getenv("SPOTIFY_TOKEN_ENCRYPTION_KEY")
	if encryptionKey == "" {
		return nil, fmt.Errorf("SPOTIFY_TOKEN_ENCRYPTION_KEY environment variable is required")
	}

	return &SpotifyConfig{
		ClientID:      clientID,
		ClientSecret:  clientSecret,
		RedirectURI:   redirectURI,
		BaseURL:       "https://api.spotify.com/v1",
		AuthURL:       "https://accounts.spotify.com/authorize",
		TokenURL:      "https://accounts.spotify.com/api/token",
		EncryptionKey: encryptionKey,
	}, nil
}

// GetAuthURL returns the Spotify authorization URL
func (c *SpotifyConfig) GetAuthURL() string {
	return c.AuthURL
}

// GetTokenURL returns the Spotify token exchange URL
func (c *SpotifyConfig) GetTokenURL() string {
	return c.TokenURL
}

// GetAPIBaseURL returns the Spotify API base URL
func (c *SpotifyConfig) GetAPIBaseURL() string {
	return c.BaseURL
}

