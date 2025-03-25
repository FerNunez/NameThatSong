package handlers

import (
	"encoding/json"
	"fmt"
	"goth/internal/base"
	"goth/internal/player"
	"goth/internal/provider"
	"goth/internal/service"
	"goth/internal/templates"
	"goth/internal/utils"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// GameHandler handles the game flow
type GameHandler struct {
	GameService *service.GameService
	Cache       *SessionCache
}

// SessionCache caches artists and albums for the UI
type SessionCache struct {
	ArtistsCache map[string]ArtistCache
	AlbumsCache  map[string]AlbumCache
}

// ArtistCache caches artist details
type ArtistCache struct {
	Artist     base.Artist
	LastUpdate time.Time
}

// AlbumCache caches album details
type AlbumCache struct {
	Album      base.Album
	LastUpdate time.Time
}

// NewGameHandler creates a new game handler
func NewGameHandler() (*GameHandler, error) {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("missing Spotify credentials in .env file")
	}

	// Create a default session
	cache := &SessionCache{
		ArtistsCache: make(map[string]ArtistCache),
		AlbumsCache:  make(map[string]AlbumCache),
	}

	// In a real application, you'd get these from a token exchange/session
	accessToken := ""
	refreshToken := ""
	deviceID := ""

	// Create song provider
	songProvider := provider.NewSpotifySongProvider(
		accessToken,
		refreshToken,
		clientID,
		clientSecret,
	)

	// Create music player
	musicPlayer := player.NewSpotifyPlayer(deviceID, accessToken)

	// Create game service
	gameService := service.NewGameService(musicPlayer, songProvider)

	return &GameHandler{
		GameService: gameService,
		Cache:       cache,
	}, nil
}

// SearchArtists handles artist search requests
func (h *GameHandler) SearchArtists(w http.ResponseWriter, r *http.Request) {
	// Get query term
	query := r.URL.Query().Get("search")
	if query == "" {
		// Return empty results component
		component := templates.SearchResults([]base.Artist{})
		component.Render(r.Context(), w)
		return
	}

	// Clean up the query
	cleanedQuery := "artist:" + strings.ToLower(query)
	fmt.Println("got artist?:", cleanedQuery)

	// Call Spotify API
	apiURL, err := url.Parse("https://api.spotify.com/v1/search")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing URL: %v", err), http.StatusInternalServerError)
		return
	}

	q := apiURL.Query()
	q.Set("type", "artist")
	q.Set("q", cleanedQuery)
	apiURL.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", apiURL.String(), nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating request: %v", err), http.StatusInternalServerError)
		return
	}

	// Get access token from the provider
	// For simplicity, we'll assume there's a method to get the token
	// In a real app, this would come from authentication
	// Let's assume something like this exists:
	accessToken := h.GameService.Player.(*player.SpotifyPlayer).AccessToken

	fmt.Println("accessToken from search artist", accessToken)

	// Set Authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error making request: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Spotify API returned error: %v", resp.Status), resp.StatusCode)
		return
	}

	// Parse response
	var searchResponse struct {
		Artists struct {
			Items []struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				Popularity int    `json:"popularity"`
				Images     []struct {
					URL    string `json:"url"`
					Height int    `json:"height"`
					Width  int    `json:"width"`
				} `json:"images"`
			} `json:"items"`
		} `json:"artists"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding response: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to base.Artist type
	artists := make([]base.Artist, 0, len(searchResponse.Artists.Items))
	for _, item := range searchResponse.Artists.Items {
		images := make([]base.Image, 0, len(item.Images))
		for _, img := range item.Images {
			images = append(images, base.NewImage(img.URL, img.Height, img.Width))
		}

		artist := base.Artist{
			ID:             item.ID,
			Name:           item.Name,
			Images:         images,
			Popularity:     item.Popularity,
			TotalFollowers: 0, // Not available in this API response
			Genres:         []string{},
			Href:           "",
			Type:           "artist",
			URI:            "",
		}

		artists = append(artists, artist)

		// Update cache
		h.Cache.ArtistsCache[artist.ID] = ArtistCache{
			Artist:     artist,
			LastUpdate: time.Now(),
		}
	}

	// Sort by popularity
	sort.Slice(artists, func(i, j int) bool {
		return artists[i].Popularity > artists[j].Popularity
	})

	// Return results component
	component := templates.SearchResults(artists)
	component.Render(r.Context(), w)
}

// GetArtistAlbums gets albums for an artist
func (h *GameHandler) GetArtistAlbums(w http.ResponseWriter, r *http.Request) {
	artistID := r.URL.Query().Get("artist_id")
	if artistID == "" {
		http.Error(w, "Artist ID is required", http.StatusBadRequest)
		return
	}

	// Call Spotify API
	apiURL := fmt.Sprintf("https://api.spotify.com/v1/artists/%s/albums", artistID)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating request: %v", err), http.StatusInternalServerError)
		return
	}

	// Get access token
	accessToken := h.GameService.Player.(*player.SpotifyPlayer).AccessToken

	// Set Authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error making request: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Spotify API returned error: %v", resp.Status), resp.StatusCode)
		return
	}

	// Parse response
	var albumsResponse struct {
		Items []struct {
			ID                   string `json:"id"`
			Name                 string `json:"name"`
			AlbumType            string `json:"album_type"`
			TotalTracks          int    `json:"total_tracks"`
			ReleaseDate          string `json:"release_date"`
			ReleaseDatePrecision string `json:"release_date_precision"`
			URI                  string `json:"uri"`
			Images               []struct {
				URL    string `json:"url"`
				Height int    `json:"height"`
				Width  int    `json:"width"`
			} `json:"images"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&albumsResponse); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding response: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to base.Album type
	albums := make([]base.Album, 0, len(albumsResponse.Items))
	for _, item := range albumsResponse.Items {
		images := make([]base.Image, 0, len(item.Images))
		for _, img := range item.Images {
			images = append(images, base.NewImage(img.URL, img.Height, img.Width))
		}

		// Parse release date
		releaseDate, err := utils.ParseReleaseDate(item.ReleaseDate, item.ReleaseDatePrecision)
		if err != nil {
			// Use a default date if parsing fails
			releaseDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		}

		// Check if the album is selected
		selected := false
		for _, id := range h.GameService.GetSelectedAlbums() {
			if id == item.ID {
				selected = true
				break
			}
		}

		album := base.Album{
			ID:               item.ID,
			Name:             item.Name,
			Images:           images,
			ReleaseDate:      releaseDate,
			AlbumType:        item.AlbumType,
			TotalTracks:      item.TotalTracks,
			URI:              item.URI,
			Selected:         selected,
			AvailableMarkets: []string{},
			ArtistsFeatureID: []string{},
			Href:             "",
		}

		albums = append(albums, album)

		// Update cache
		h.Cache.AlbumsCache[album.ID] = AlbumCache{
			Album:      album,
			LastUpdate: time.Now(),
		}
	}

	// Return results component
	// If templates.AlbumsList doesn't exist yet, you'll need to create it
	// For now we'll use AlbumCard for each album
	for _, album := range albums {
		component := templates.AlbumCard(album)
		component.Render(r.Context(), w)
	}
}

// SelectAlbum handles album selection
func (h *GameHandler) SelectAlbum(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	albumID := r.Form.Get("albumID")
	if albumID == "" {
		http.Error(w, "Album ID is required", http.StatusBadRequest)
		return
	}

	// Toggle album selection
	err := h.GameService.SelectAlbum(albumID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error selecting album: %v", err), http.StatusInternalServerError)
		return
	}

	// Update UI state
	if albumCache, exists := h.Cache.AlbumsCache[albumID]; exists {
		albumCache.Album.Selected = !albumCache.Album.Selected
		h.Cache.AlbumsCache[albumID] = albumCache

		// Return updated album card
		component := templates.AlbumCard(albumCache.Album)
		component.Render(r.Context(), w)
	} else {
		http.Error(w, "Album not found in cache", http.StatusNotFound)
	}
}

// StartGame handles game start
func (h *GameHandler) StartGame(w http.ResponseWriter, r *http.Request) {
	err := h.GameService.StartGame()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error starting game: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Game started successfully"))
}

// PlayGame handles play button
func (h *GameHandler) PlayGame(w http.ResponseWriter, r *http.Request) {
	err := h.GameService.PlayGame()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error playing game: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Playback started"))
}

// MakeGuess handles user guesses
func (h *GameHandler) MakeGuess(w http.ResponseWriter, r *http.Request) {
	guess := r.URL.Query().Get("guess")
	if guess == "" {
		http.Error(w, "Guess is required", http.StatusBadRequest)
		return
	}

	correct, actualSong, err := h.GameService.MakeGuess(guess)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error processing guess: %v", err), http.StatusInternalServerError)
		return
	}

	result := struct {
		Correct    bool   `json:"correct"`
		ActualSong string `json:"actualSong"`
		Score      int    `json:"score"`
	}{
		Correct:    correct,
		ActualSong: actualSong,
		Score:      h.GameService.GetScore(),
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// SkipSong handles skip button
func (h *GameHandler) SkipSong(w http.ResponseWriter, r *http.Request) {
	err := h.GameService.SkipSong()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error skipping song: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Skipped to next song"))
}
