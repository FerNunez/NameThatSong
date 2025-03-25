package provider

import (
	"encoding/json"
	"fmt"
	"goth/internal/player"
	"net/http"
	"strings"
	"sync"
	"time"
)

// SpotifySongProvider implements the SongProvider interface for Spotify
type SpotifySongProvider struct {
	AccessToken  string
	RefreshToken string
	ClientID     string
	ClientSecret string
	Cache        *SongCache
}

// SongCache provides caching for songs to reduce API calls
type SongCache struct {
	Songs      map[string]player.Song
	Albums     map[string][]string // AlbumID -> []SongID
	LastUpdate map[string]time.Time
	mu         sync.RWMutex
}

// NewSongCache creates a new song cache
func NewSongCache() *SongCache {
	return &SongCache{
		Songs:      make(map[string]player.Song),
		Albums:     make(map[string][]string),
		LastUpdate: make(map[string]time.Time),
		mu:         sync.RWMutex{},
	}
}

// NewSpotifySongProvider creates a new SpotifySongProvider
func NewSpotifySongProvider(accessToken, refreshToken, clientID, clientSecret string) *SpotifySongProvider {
	return &SpotifySongProvider{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Cache:        NewSongCache(),
	}
}

// GetSongsByIDs retrieves songs by their IDs
func (p *SpotifySongProvider) GetSongsByIDs(ids []string) ([]player.Song, error) {
	result := make([]player.Song, 0, len(ids))
	missingIDs := make([]string, 0)

	// Check cache first
	p.Cache.mu.RLock()
	for _, id := range ids {
		if song, exists := p.Cache.Songs[id]; exists {
			// Check if cache entry is still valid (e.g., less than 24 hours old)
			if time.Since(p.Cache.LastUpdate[id]) < 24*time.Hour {
				result = append(result, song)
				continue
			}
		}
		missingIDs = append(missingIDs, id)
	}
	p.Cache.mu.RUnlock()

	// If all songs were in cache, return immediately
	if len(missingIDs) == 0 {
		return result, nil
	}

	// Otherwise, fetch the missing songs from the API
	// We can fetch up to 50 songs at once
	for i := 0; i < len(missingIDs); i += 50 {
		end := min(i+50, len(missingIDs))

		batchIDs := missingIDs[i:end]
		songs, err := p.fetchSongs(batchIDs)
		if err != nil {
			return nil, err
		}

		result = append(result, songs...)
	}

	return result, nil
}

// GetSongsByAlbumID retrieves all songs from a specific album
func (p *SpotifySongProvider) GetSongsByAlbumID(albumID string) ([]player.Song, error) {
	// Check if we have this album's tracks cached
	p.Cache.mu.RLock()
	if songIDs, exists := p.Cache.Albums[albumID]; exists {
		// If all songs are in cache, return them
		allInCache := true
		for _, id := range songIDs {
			if _, exists := p.Cache.Songs[id]; !exists {
				allInCache = false
				break
			}
		}

		if allInCache {
			songs := make([]player.Song, len(songIDs))
			for i, id := range songIDs {
				songs[i] = p.Cache.Songs[id]
			}
			p.Cache.mu.RUnlock()
			return songs, nil
		}
	}
	p.Cache.mu.RUnlock()

	// If not cached, fetch from API
	requestURL := fmt.Sprintf("https://api.spotify.com/v1/albums/%s/tracks", albumID)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", p.AccessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	var albumTracksResponse struct {
		Items []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			URI         string `json:"uri"`
			DurationMs  int    `json:"duration_ms"`
			DiscNumber  int    `json:"disc_number"`
			TrackNumber int    `json:"track_number"`
			Artists     []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"artists"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&albumTracksResponse); err != nil {
		return nil, fmt.Errorf("could not decode response: %v", err)
	}

	// Convert to our Song type
	songs := make([]player.Song, 0, len(albumTracksResponse.Items))
	songIDs := make([]string, 0, len(albumTracksResponse.Items))

	for _, item := range albumTracksResponse.Items {
		artistIDs := make([]string, len(item.Artists))
		for i, artist := range item.Artists {
			artistIDs[i] = artist.ID
		}

		song := player.Song{
			ID:       item.ID,
			Name:     item.Name,
			URI:      item.URI,
			Duration: item.DurationMs,
			AlbumID:  albumID,
			Artists:  artistIDs,
		}

		songs = append(songs, song)
		songIDs = append(songIDs, item.ID)

		// Update cache
		p.Cache.mu.Lock()
		p.Cache.Songs[item.ID] = song
		p.Cache.LastUpdate[item.ID] = time.Now()
		p.Cache.mu.Unlock()
	}

	// Update album cache
	p.Cache.mu.Lock()
	p.Cache.Albums[albumID] = songIDs
	p.Cache.mu.Unlock()

	return songs, nil
}

// GetSongsByAlbumIDs retrieves songs for multiple albums
func (p *SpotifySongProvider) GetSongsByAlbumIDs(albumIDs []string) (map[string][]player.Song, error) {
	result := make(map[string][]player.Song, len(albumIDs))

	// We can parallelize the album requests for better performance
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstError error

	for _, albumID := range albumIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			songs, err := p.GetSongsByAlbumID(id)
			if err != nil {
				mu.Lock()
				if firstError == nil {
					firstError = err
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			result[id] = songs
			mu.Unlock()
		}(albumID)
	}

	wg.Wait()

	if firstError != nil {
		return nil, firstError
	}

	return result, nil
}

// fetchSongs fetches song details from Spotify API
func (p *SpotifySongProvider) fetchSongs(ids []string) ([]player.Song, error) {
	if len(ids) == 0 {
		return []player.Song{}, nil
	}

	idsParam := strings.Join(ids, ",")
	requestURL := fmt.Sprintf("https://api.spotify.com/v1/tracks?ids=%s", idsParam)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", p.AccessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	var tracksResponse struct {
		Tracks []struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			URI        string `json:"uri"`
			DurationMs int    `json:"duration_ms"`
			Album      struct {
				ID string `json:"id"`
			} `json:"album"`
			Artists []struct {
				ID string `json:"id"`
			} `json:"artists"`
		} `json:"tracks"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tracksResponse); err != nil {
		return nil, fmt.Errorf("could not decode response: %v", err)
	}

	songs := make([]player.Song, 0, len(tracksResponse.Tracks))

	p.Cache.mu.Lock()
	defer p.Cache.mu.Unlock()

	for _, track := range tracksResponse.Tracks {
		artistIDs := make([]string, len(track.Artists))
		for i, artist := range track.Artists {
			artistIDs[i] = artist.ID
		}

		song := player.Song{
			ID:       track.ID,
			Name:     track.Name,
			URI:      track.URI,
			Duration: track.DurationMs,
			AlbumID:  track.Album.ID,
			Artists:  artistIDs,
		}

		songs = append(songs, song)

		// Update cache
		p.Cache.Songs[track.ID] = song
		p.Cache.LastUpdate[track.ID] = time.Now()
	}

	return songs, nil
}
