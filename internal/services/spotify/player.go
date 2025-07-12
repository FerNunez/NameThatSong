package spotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/config"
)

// SpotifyPlayerService handles playback controls
type SpotifyPlayerService struct {
	config      *config.SpotifyConfig
	authService *SpotifyAuthService
	httpClient  *http.Client
}

// NewSpotifyPlayerService creates a new Spotify player service
func NewSpotifyPlayerService(config *config.SpotifyConfig, authService *SpotifyAuthService, httpClient *http.Client) *SpotifyPlayerService {
	return &SpotifyPlayerService{
		config:      config,
		authService: authService,
		httpClient:  httpClient,
	}
}

// PlaySong starts playing a specific song on the user's active device
func (s *SpotifyPlayerService) PlaySong(ctx context.Context, userID, songID string) error {
	// Get access token for user
	accessToken, err := s.authService.GetValidToken(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get access token: %v", err)
	}
	type PlaySongRequest struct {
		Uris       []string `json:"uris"`
		PositionMs int      `json:"position_ms"`
	}

	psr := PlaySongRequest{
		Uris:       []string{fmt.Sprintf("spotify:track:%v", songID)},
		PositionMs: 0,
	}

	dat, err := json.Marshal(psr)
	if err != nil {
		return fmt.Errorf("could not marshal request: %v", err)
	}

	// Set request
	url := s.config.GetAPIBaseURL() + "/me/player/play"
	//if c.DeviceID != "" {
	//	url = fmt.Sprintf("%s?device_id=%s", url, c.DeviceID)
	//}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(dat))
	if err != nil {
		return fmt.Errorf("could not create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	return nil
}

// PausePlayback pauses the current playback on the user's active device
func (s *SpotifyPlayerService) PausePlayback(ctx context.Context, userID string) error {
	// Get access token for user
	accessToken, err := s.authService.GetValidToken(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get access token: %v", err)
	}
	url := s.config.GetAPIBaseURL() + "/me/player/pause"
	//if c.DeviceID != "" {
	//	url = fmt.Sprintf("%s?device_id=%s", url, c.DeviceID)
	//}

	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return fmt.Errorf("could not create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	return nil
}

// ResumePlayback resumes the current playback on the user's active device
func (s *SpotifyPlayerService) ResumePlayback(ctx context.Context, userID string) error {
	// Get access token for user
	accessToken, err := s.authService.GetValidToken(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get access token: %v", err)
	}
	url := s.config.GetAPIBaseURL() + "/me/player/play"
	// if c.DeviceID != "" {
	// 	url = fmt.Sprintf("%s?device_id=%s", url, c.DeviceID)
	// }
	//
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return fmt.Errorf("could not create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	return nil
}
