package spotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	m "github.com/FerNunez/NameThatSong/internal/models"
)

func (s *Spotify) GetUserPlaylists(ctx context.Context, userID string) ([]m.PlaylistData, error) {
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	playlists, err := s.getUserPlaylistsFromAPI(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	return playlists, nil
}
func (s *Spotify) CreatePlaylist(ctx context.Context, userID, name, description string, isPublic bool) (m.PlaylistData, error) {
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		return m.PlaylistData{}, fmt.Errorf("failed to get access token: %v", err)
	}

	playlist, err := s.createPlaylistOnSpotify(ctx, accessToken, userID, name, description, isPublic)
	if err != nil {
		return m.PlaylistData{}, err
	}

	return playlist, nil
}
func (s *Spotify) AddTracksToPlaylist(ctx context.Context, userID, playlistID string, trackIDs []string) error {
	if len(trackIDs) == 0 {
		return fmt.Errorf("empty track list")
	}

	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get access token: %v", err)
	}

	return s.addTracksToSpotifyPlaylist(ctx, accessToken, playlistID, trackIDs)
}

func (s *Spotify) RemoveTracksFromPlaylist(ctx context.Context, userID, playlistID string, trackIDs []string) error {
	if len(trackIDs) == 0 {
		return nil
	}

	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get access token: %v", err)
	}

	return s.removeTracksFromSpotifyPlaylist(ctx, accessToken, playlistID, trackIDs)
}

func (s *Spotify) getUserPlaylistsFromAPI(ctx context.Context, accessToken string) ([]m.PlaylistData, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.spotify.com/v1/me/playlists", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("spotify API error: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	type spotifyPlaylistsResponse struct {
		Items []struct {
			Description string `json:"description"`
			Followers   struct {
				Total int `json:"total"`
			} `json:"followers"`
			ID     string `json:"id"`
			Images []struct {
				URL string `json:"url"`
			} `json:"images"`
			Name   string `json:"name"`
			Public bool   `json:"public"`
			Tracks struct {
				Total int `json:"total"`
			} `json:"tracks"`
		} `json:"items"`
	}

	var response spotifyPlaylistsResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	playlists := make([]m.PlaylistData, len(response.Items))
	for i, item := range response.Items {
		imageUrl := ""
		if len(item.Images) > 0 {
			imageUrl = item.Images[0].URL
		}

		playlists[i] = m.PlaylistData{
			Description:    item.Description,
			FollowersTotal: item.Followers.Total,
			ID:             item.ID,
			ImageUrl:       imageUrl,
			Name:           item.Name,
			Public:         item.Public,
			TotalTracks:    item.Tracks.Total,
		}
	}

	return playlists, nil
}

func (s *Spotify) createPlaylistOnSpotify(ctx context.Context, accessToken, userID, name, description string, isPublic bool) (m.PlaylistData, error) {
	requestBody := map[string]any{
		"name":        name,
		"description": description,
		"public":      isPublic,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return m.PlaylistData{}, fmt.Errorf("failed to marshal request body: %v", err)
	}

	requestURL := fmt.Sprintf("https://api.spotify.com/v1/users/%s/playlists", userID)
	req, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return m.PlaylistData{}, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return m.PlaylistData{}, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return m.PlaylistData{}, fmt.Errorf("spotify API error: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return m.PlaylistData{}, fmt.Errorf("failed to read response body: %v", err)
	}

	type spotifyCreatePlaylistResponse struct {
		Description string `json:"description"`
		Followers   struct {
			Total int `json:"total"`
		} `json:"followers"`
		ID     string `json:"id"`
		Images []struct {
			URL string `json:"url"`
		} `json:"images"`
		Name   string `json:"name"`
		Public bool   `json:"public"`
		Tracks struct {
			Total int `json:"total"`
		} `json:"tracks"`
	}
	var response spotifyCreatePlaylistResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return m.PlaylistData{}, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	imageUrl := ""
	if len(response.Images) > 0 {
		imageUrl = response.Images[0].URL
	}

	return m.PlaylistData{
		Description:    response.Description,
		FollowersTotal: response.Followers.Total,
		ID:             response.ID,
		ImageUrl:       imageUrl,
		Name:           response.Name,
		Public:         response.Public,
		TotalTracks:    response.Tracks.Total,
	}, nil
}

func (s *Spotify) addTracksToSpotifyPlaylist(ctx context.Context, accessToken, playlistID string, trackIDs []string) error {
	// Convert track IDs to Spotify URIs
	uris := make([]string, len(trackIDs))
	for i, trackID := range trackIDs {
		uris[i] = "spotify:track:" + trackID
	}

	// Spotify API limits to 100 tracks per request, so batch them
	const batchSize = 100
	for i := 0; i < len(uris); i += batchSize {
		end := min(i+batchSize, len(uris))

		batch := uris[i:end]
		if err := s.addTrackBatchToPlaylist(ctx, accessToken, playlistID, batch); err != nil {
			return fmt.Errorf("failed to add track batch %d-%d: %v", i, end-1, err)
		}
	}

	return nil
}

func (s *Spotify) addTrackBatchToPlaylist(ctx context.Context, accessToken, playlistID string, uris []string) error {
	requestBody := map[string]any{
		"uris": uris,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	requestURL := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistID)
	req, err := http.NewRequestWithContext(ctx, "POST", requestURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("spotify API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

func (s *Spotify) removeTracksFromSpotifyPlaylist(ctx context.Context, accessToken, playlistID string, trackIDs []string) error {
	// Convert track IDs to Spotify URIs
	tracks := make([]map[string]string, len(trackIDs))
	for i, trackID := range trackIDs {
		tracks[i] = map[string]string{
			"uri": "spotify:track:" + trackID,
		}
	}

	// Spotify API limits to 100 tracks per request, so batch them
	const batchSize = 100
	for i := 0; i < len(tracks); i += batchSize {
		end := min(i+batchSize, len(tracks))

		batch := tracks[i:end]
		if err := s.removeTrackBatchFromPlaylist(ctx, accessToken, playlistID, batch); err != nil {
			return fmt.Errorf("failed to remove track batch %d-%d: %v", i, end-1, err)
		}
	}

	return nil
}

func (s *Spotify) removeTrackBatchFromPlaylist(ctx context.Context, accessToken, playlistID string, tracks []map[string]string) error {
	requestBody := map[string]any{
		"tracks": tracks,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	requestURL := fmt.Sprintf("https://api.spotify.com/v1/playlists/%s/tracks", playlistID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", requestURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("spotify API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}
