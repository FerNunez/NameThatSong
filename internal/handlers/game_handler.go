package handlers

import (
	"encoding/base64"
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

	"slices"

	"github.com/a-h/templ"
	"github.com/joho/godotenv"
)

// GameHandler handles the game flow
type GameHandler struct {
	GameService *service.GameService
	Cache       *SessionCache
	// Authentication config
	ClientID     string
	ClientSecret string
	RedirectURI  string
	State        string
	AccessToken  string
	RefreshToken string
	DeviceID     string
}

// SessionCache caches artists and albums for the UI
type SessionCache struct {
	ArtistsCache map[string]ArtistCache
	AlbumsCache  map[string]AlbumCache
	TracksCache  map[string]TrackCache
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

// TrackCache caches track details
type TrackCache struct {
	Track      base.Track
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
	redirectURI := "http://127.0.0.1:8080/auth/callback"

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("missing Spotify credentials in .env file")
	}

	// Generate a random state for OAuth
	state, err := utils.GenerateState(16)
	if err != nil {
		return nil, fmt.Errorf("error generating state: %v", err)
	}

	// Create a default session
	cache := &SessionCache{
		ArtistsCache: make(map[string]ArtistCache),
		AlbumsCache:  make(map[string]AlbumCache),
		TracksCache:  make(map[string]TrackCache),
	}

	// For now, create with empty tokens - they'll be set after auth
	accessToken := ""
	refreshToken := ""
	deviceID := ""

	// Create music service client
	musicServiceClient := provider.NewSpotifyMusicServiceClient(accessToken, refreshToken, clientID, clientSecret, deviceID)

	// Create music player
	musicPlayer := player.NewSpotifyPlayer(musicServiceClient)

	// Create song provider
	songProvider := provider.NewSpotifySongProvider(accessToken, refreshToken, clientID, clientSecret)

	// Create game service
	gameService := service.NewGameService(musicPlayer, songProvider)

	return &GameHandler{
		GameService:  gameService,
		Cache:        cache,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
		State:        state,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		DeviceID:     deviceID,
	}, nil
}

// IndexHandler renders the main page
func (h *GameHandler) IndexHandler(w http.ResponseWriter, r *http.Request) {
	component := templates.IndexPage()
	layout := templates.Layout(component, "NameThatSong")
	layout.Render(r.Context(), w)
}

// AuthHandler handles the Spotify OAuth login flow
func (h *GameHandler) AuthHandler(w http.ResponseWriter, r *http.Request) {
	// Build the authorization URL
	authURL := "https://accounts.spotify.com/authorize"
	u, err := url.Parse(authURL)
	if err != nil {
		http.Error(w, "Error parsing URL", http.StatusInternalServerError)
		return
	}

	// Add query parameters
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", h.ClientID)
	q.Set("scope", "user-read-private user-read-email streaming user-modify-playback-state user-read-playback-state")
	q.Set("redirect_uri", h.RedirectURI)
	q.Set("state", h.State)
	u.RawQuery = q.Encode()

	// Redirect to Spotify
	http.Redirect(w, r, u.String(), http.StatusFound)
}

// AuthCallbackHandler handles the callback from Spotify after login
func (h *GameHandler) AuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Check state to prevent CSRF
	state := r.URL.Query().Get("state")
	if state != h.State {
		http.Error(w, "State mismatch", http.StatusBadRequest)
		return
	}

	// Get the authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code provided", http.StatusBadRequest)
		return
	}

	// Exchange code for tokens
	tokenURL := "https://accounts.spotify.com/api/token"
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", h.RedirectURI)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating request: %v", err), http.StatusInternalServerError)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	auth := base64.StdEncoding.EncodeToString([]byte(h.ClientID + ":" + h.ClientSecret))
	req.Header.Set("Authorization", "Basic "+auth)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting token: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Parse response
	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing token response: %v", err), http.StatusInternalServerError)
		return
	}

	// Store tokens
	h.AccessToken = tokenResponse.AccessToken
	h.RefreshToken = tokenResponse.RefreshToken

	fmt.Println("AccessToken: ", h.AccessToken)
	fmt.Println("RefreshToken: ", h.RefreshToken)

	// Get available devices
	h.DeviceID = h.getFirstAvailableDevice()

	// Create a new music service client with updated tokens
	musicServiceClient := provider.NewSpotifyMusicServiceClient(
		h.AccessToken,
		h.RefreshToken,
		h.ClientID,
		h.ClientSecret,
		h.DeviceID,
	)

	// Create a new player with the updated music service client
	musicPlayer := player.NewSpotifyPlayer(musicServiceClient)

	// Update the GameService with the new player
	h.GameService.Player = musicPlayer

	// Update the provider as well
	h.GameService.SongProvider.(*provider.SpotifySongProvider).AccessToken = h.AccessToken
	h.GameService.SongProvider.(*provider.SpotifySongProvider).RefreshToken = h.RefreshToken

	// Redirect back to the main page
	http.Redirect(w, r, "/", http.StatusFound)
}

// getFirstAvailableDevice gets the first available Spotify device
func (h *GameHandler) getFirstAvailableDevice() string {
	// Instead of directly calling the Spotify API, create a temporary music service client
	tempClient := provider.NewSpotifyMusicServiceClient(
		h.AccessToken,
		h.RefreshToken,
		h.ClientID,
		h.ClientSecret,
		"", // Empty device ID
	)

	// Use the client to get the active device
	deviceID, err := tempClient.GetActiveDevice()
	if err != nil {
		fmt.Printf("Error getting active device: %v\n", err)
		return ""
	}

	return deviceID
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

	// Set Authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.AccessToken))

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
	artistID := r.URL.Query().Get("artist-id")
	if artistID == "" {
		http.Error(w, "Artist ID is required", http.StatusBadRequest)
		fmt.Println("no album")
		return
	}

	// Call Spotify API
	apiURL := fmt.Sprintf("https://api.spotify.com/v1/artists/%s/albums", artistID)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating request: %v", err), http.StatusInternalServerError)
		return
	}

	// Set Authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.AccessToken))

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
		selected := slices.Contains(h.GameService.GetSelectedAlbums(), item.ID)

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

	// Return albums grid
	component := templates.AlbumGrid(albums)
	component.Render(r.Context(), w)
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
	// Get all selected albums
	selectedAlbums := h.GameService.GetSelectedAlbums()
	if len(selectedAlbums) == 0 {
		http.Error(w, "No albums selected", http.StatusBadRequest)
		return
	}

	// Start the game
	err := h.GameService.StartGame()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error starting game: %v", err), http.StatusInternalServerError)
		return
	}

	// Send successful response
	w.Write([]byte("Game started successfully! Press Play to begin."))
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

// GuessHelper handles searching for tracks to guess
func (h *GameHandler) GuessHelper(w http.ResponseWriter, r *http.Request) {
	// Get the query term
	query := r.URL.Query().Get("guess")
	if query == "" {
		// Return empty results
		w.Write([]byte(""))
		return
	}

	// Format the query for track search
	cleanedQuery := utils.ParseGuessText(query)

	// Call Spotify API to search for tracks
	apiURL, err := url.Parse("https://api.spotify.com/v1/search")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing URL: %v", err), http.StatusInternalServerError)
		return
	}

	q := apiURL.Query()
	q.Set("type", "track")
	q.Set("q", cleanedQuery)
	q.Set("limit", "5") // Limit to 5 results
	apiURL.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", apiURL.String(), nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating request: %v", err), http.StatusInternalServerError)
		return
	}

	// Set Authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.AccessToken))

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
	var searchResponse base.SearchTrackResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding response: %v", err), http.StatusInternalServerError)
		return
	}

	// Process the tracks
	tracks, artistsNames, albumUrls := h.parseTracks(searchResponse)

	// Sort tracks by popularity in descending order
	sort.Slice(tracks, func(i, j int) bool {
		return tracks[i].Popularity > tracks[j].Popularity
	})

	// Return results component
	component := templates.GuessHelperList(tracks, artistsNames, albumUrls)
	component.Render(r.Context(), w)
}

// parseTracks converts API response to track data
func (h *GameHandler) parseTracks(response base.SearchTrackResponse) ([]base.Track, []string, []string) {
	var tracks []base.Track
	var albumUrls []string
	var artistsNames []string

	for _, item := range response.Tracks.Items {
		// Process album
		albumJson := item.Album
		images := make([]base.Image, len(albumJson.Images))
		for _, img := range albumJson.Images {
			images = append(images, base.NewImage(img.URL, img.Height, img.Width))
		}

		releaseDate, _ := utils.ParseReleaseDate(albumJson.ReleaseDate, albumJson.ReleaseDatePrecision)
		artistIDs := make([]string, len(item.Artists))
		for i, artist := range item.Artists {
			artistIDs[i] = artist.ID
		}

		album := base.Album{
			AlbumType:        albumJson.AlbumType,
			TotalTracks:      0,
			AvailableMarkets: albumJson.AvailableMarkets,
			Href:             albumJson.Href,
			ID:               albumJson.ID,
			Images:           images,
			Name:             albumJson.Name,
			ReleaseDate:      releaseDate,
			URI:              albumJson.URI,
			ArtistsFeatureID: artistIDs,
			Selected:         false,
		}

		// Update album cache
		h.Cache.AlbumsCache[album.ID] = AlbumCache{
			Album:      album,
			LastUpdate: time.Now(),
		}

		// Process artists
		for _, artistJson := range item.Artists {
			artist := base.Artist{
				TotalFollowers: 0,
				Genres:         []string{"Unknown"},
				Href:           artistJson.Href,
				ID:             artistJson.ID,
				Images:         nil,
				Name:           artistJson.Name,
				Popularity:     0,
				Type:           artistJson.Type,
				URI:            artistJson.URI,
			}

			// Update artist cache
			h.Cache.ArtistsCache[artistJson.ID] = ArtistCache{
				Artist:     artist,
				LastUpdate: time.Now(),
			}
		}

		// Collect artist names and album image
		artistsName := ""
		for _, artist := range item.Artists {
			if artistsName != "" {
				artistsName += ", "
			}
			artistsName += artist.Name
		}

		if len(item.Album.Images) <= 0 {
			// Use a placeholder if no image
			albumUrls = append(albumUrls, "https://via.placeholder.com/300")
		} else {
			albumUrls = append(albumUrls, item.Album.Images[0].URL)
		}
		artistsNames = append(artistsNames, artistsName)

		// Create track
		track := base.Track{
			DiscNumber:  item.DiscNumber,
			DurationMs:  item.DurationMs,
			Href:        item.Href,
			ID:          item.ID,
			IsPlayable:  item.IsPlayable,
			Name:        item.Name,
			Popularity:  item.Popularity,
			TrackNumber: item.TrackNumber,
			URI:         item.URI,
			AlbumID:     item.Album.ID,
			ArtistsID:   artistIDs,
			Selected:    false,
		}

		// Update track cache
		h.Cache.TracksCache[track.ID] = TrackCache{
			Track:      track,
			LastUpdate: time.Now(),
		}

		tracks = append(tracks, track)
	}

	return tracks, artistsNames, albumUrls
}

// SelectTrack handles track selection for guessing
func (h *GameHandler) SelectTrack(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	trackID := r.Form.Get("trackID")
	if trackID == "" {
		http.Error(w, "Track ID is required", http.StatusBadRequest)
		return
	}

	// Get the current track being played
	currentSong := h.GameService.Player.GetCurrentSong()
	if currentSong == nil {
		http.Error(w, "No song is currently playing", http.StatusBadRequest)
		return
	}

	// Check if the guess is correct
	correct := currentSong.ID == trackID

	// Get track info from cache
	track, exists := h.Cache.TracksCache[trackID]
	if !exists {
		http.Error(w, "Track not found in cache", http.StatusNotFound)
		return
	}

	// Update track selection state
	track.Track.Selected = true
	h.Cache.TracksCache[trackID] = track

	// Return a customized response based on correctness
	var component templ.Component
	var artistName string

	// Find the artist name
	for _, artistID := range track.Track.ArtistsID {
		if artist, ok := h.Cache.ArtistsCache[artistID]; ok {
			if artistName != "" {
				artistName += ", "
			}
			artistName += artist.Artist.Name
		}
	}

	// Get album image URL
	albumURL := ""
	if album, ok := h.Cache.AlbumsCache[track.Track.AlbumID]; ok {
		if len(album.Album.Images) > 0 {
			albumURL = album.Album.Images[0].URL
		}
	}

	if correct {
		// Render correct guess component (to be created)
		// For now, just render the track with a different style
		component = templates.GuessOptionCard(track.Track.ID, track.Track.Name+" - CORRECT!", albumURL, artistName, true)

		// Skip to next song after a delay
		go func() {
			time.Sleep(3 * time.Second)
			h.GameService.SkipSong()
		}()
	} else {
		// Render incorrect guess component
		component = templates.GuessOptionCard(track.Track.ID, track.Track.Name+" - Try again!", albumURL, artistName, false)
	}

	component.Render(r.Context(), w)
}

// GuessTrack is a legacy handler still needed for the frontend
func (h *GameHandler) GuessTrack(w http.ResponseWriter, r *http.Request) {
	guess := r.FormValue("guess")
	if guess == "" {
		http.Error(w, "Guess is required", http.StatusBadRequest)
		return
	}

	// Process the guess using the new service
	correct, actualSong, err := h.GameService.MakeGuess(guess)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error processing guess: %v", err), http.StatusInternalServerError)
		return
	}

	// Return a simple response
	var message string
	if correct {
		message = fmt.Sprintf("Correct! The song is '%s'", actualSong)

		// Skip to next song after a delay
		go func() {
			time.Sleep(3 * time.Second)
			h.GameService.SkipSong()
		}()
	} else {
		message = "Incorrect. Try again!"
	}

	w.Write([]byte(message))
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
