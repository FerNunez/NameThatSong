package handlers

import (
	"fmt"
	"github.com/joho/godotenv"
	"goth/internal/music_player"
	"goth/internal/service"
	"goth/internal/spotify_api"
	"goth/internal/templates"
	"goth/internal/utils"
	"net/http"
	"os"
	"sort"
)

// GameHandler handles the game flow
type GameHandler struct {
	GameService *service.GameService
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

	// Create song provider
	songProvider := spotify_api.NewSpotifySongProvider(clientID, clientSecret, redirectURI, state)
	// Create music service client
	musicPlayer := player.NewMusicPlayer()
	// Create game service
	gameService := service.NewGameService(musicPlayer, songProvider)
	return &GameHandler{
		GameService: gameService,
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

	urlString, err := h.GameService.SpotifyApi.AuthRequestURL()
	if err != nil {
		fmt.Printf("error generating state: %v", err)
		http.Error(w, "error generating state", http.StatusBadRequest)
		return
	}
	// Redirect to Spotify
	http.Redirect(w, r, urlString, http.StatusFound)
}

// AuthCallbackHandler handles the callback from Spotify after login
func (h *GameHandler) AuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Check state to prevent CSRF
	state := r.URL.Query().Get("state")
	err := h.GameService.SpotifyApi.ValidateState(state)
	if err != nil {
		http.Error(w, "State unvalidated", http.StatusInternalServerError)
		return
	}

	// Get the authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "No code provided", http.StatusBadRequest)
		return
	}
	h.GameService.SpotifyApi.TokenExchange(code)

	http.Redirect(w, r, "/", http.StatusFound)
}

// getFirstAvailableDevice gets the first available Spotify device
//func (h *GameHandler) getFirstAvailableDevice() string { }

// SearchArtists handles artist search requests
func (h *GameHandler) SearchArtists(w http.ResponseWriter, r *http.Request) {
	// Get query term
	query := r.URL.Query().Get("search")
	if query == "" {
		// Return empty results component
		component := templates.SearchResults([]spotify_api.ArtistData{})
		component.Render(r.Context(), w)
		return
	}
	fmt.Println("search artist: ", query)

	artists, err := h.GameService.SearchArtists(query)
	if err != nil || len(artists) == 0 {
		// TODO: SEND ERROR TO REQUEST
		component := templates.SearchResults([]spotify_api.ArtistData{})
		component.Render(r.Context(), w)
		return
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
	fmt.Println("got artist ID", artistID)
	if artistID == "" {
		http.Error(w, "Artist ID is required", http.StatusBadRequest)
		return
	}

	albums, err := h.GameService.GetArtistsAlbum(artistID)
	//albums, err := h.GameService.SpotifyApi.FetchAlbumByArtistID(artistID)
	if err != nil {
		http.Error(w, "Cant retrieve Artist ID albums", http.StatusBadRequest)
		fmt.Println("no album")
		return
	}

	component := templates.AlbumDropdown(albums, h.GameService.AlbumSelection, artistID)
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
	artistID := r.Form.Get("artistID")
	if artistID == "" {
		http.Error(w, "Album ID is required", http.StatusBadRequest)
		return
	}
	fmt.Println("artistID", artistID)

	// Toggle album selection
	toggle := h.GameService.ToggleAlbumSelection(albumID, artistID)

	album, ok := h.GameService.Cache.AlbumMap[albumID]
	if !ok {
		panic("album should be in cache")
	}
	component := templates.AlbumCard(album, toggle, artistID)
	component.Render(r.Context(), w)
}

// StartGame handles game start
func (h *GameHandler) StartGame(w http.ResponseWriter, r *http.Request) {
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
	err := h.GameService.SpotifyApi.PausePlayback()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error play game: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Playback started"))
}

// GuessHelper handles searching for tracks to guess
func (h *GameHandler) GuessHelper(w http.ResponseWriter, r *http.Request) {
	// // TODO: FIX THIS
	// // Get the query term
	// query := r.URL.Query().Get("guess")
	// if query == "" {
	// 	// Return empty results
	// 	w.Write([]byte(""))
	// 	return
	// }
	//
	// // for _, track := range h.GameService.TracksOptions {
	// // 	fmt.Println("t:", track.Name)
	// // }
	//
	// guessOptions, err := h.GameService.FilterTrackOptions(query)
	// if err != nil {
	// 	return
	// }
	//
	// // for _, track := range guessOptions {
	// // 	fmt.Printf("o: %v, id: %v,\n", track.Name, track.ID)
	// //
	// // }
	//
	// //
	// // // Sort tracks by popularity in descending order
	// // sort.Slice(tracks, func(i, j int) bool {
	// // 	return tracks[i].Popularity > tracks[j].Popularity
	// // })
	// //
	// // Return results component
	//
	// artists := make([]spotify_api.ArtistData, 0, len(guessOptions))
	// albums := make([]spotify_api.AlbumData, 0, len(guessOptions))
	// for _, track := range guessOptions {
	//
	// 	albumId, ok := h.GameService.Cache.TrackIdToAlbumId[track.ID]
	// 	if !ok {
	// 		panic("album id should already have in cahce")
	// 	}
	// 	album, ok := h.GameService.Cache.AlbumMap[albumId]
	// 	if !ok {
	// 		panic("album id should already have in cahce")
	// 	}
	// 	albums = append(albums, album)
	//
	// 	_artistId, ok := h.GameService.Cache.AlbumIdToArtistId[albumId]
	// 	if !ok {
	// 		panic("artist id should already have in cahce")
	// 	}
	// 	// album, ok := h.GameService.Cache.ArtistToAlbumsMap[albumId]
	// 	// if !ok {
	// 	// 	panic("album id should already have in cahce")
	// 	// }
	// 	albums = append(albums, album)
	// 	artists = append(artists, spotify_api.ArtistData{})
	// }
	//
	// artist := h.GameService.Cache.AlbumIdToArtistId
	// component := templates.GuessHelperList(guessOptions, artists, albums)
	// component.Render(r.Context(), w)
}

// parseTracks converts API response to track data
// SelectTrack handles track selection for guessing
func (h *GameHandler) SelectTrack(w http.ResponseWriter, r *http.Request) {
	// if err := r.ParseForm(); err != nil {
	// 	http.Error(w, "Failed to parse form", http.StatusBadRequest)
	// 	return
	// }
	//
	// trackID := r.Form.Get("trackID")
	// if trackID == "" {
	// 	http.Error(w, "Track ID is required", http.StatusBadRequest)
	// 	return
	// }
	//
	// // Get the current track being played
	// currentSong := h.GameService.Player.GetCurrentSong()
	// if currentSong == nil {
	// 	http.Error(w, "No song is currently playing", http.StatusBadRequest)
	// 	return
	// }
	//
	// // Check if the guess is correct
	// correct := currentSong.ID == trackID
	//
	// // Get track info from cache
	// track, exists := h.Cache.TracksCache[trackID]
	// if !exists {
	// 	http.Error(w, "Track not found in cache", http.StatusNotFound)
	// 	return
	// }
	//
	// // Update track selection state
	// track.Track.Selected = true
	// h.Cache.TracksCache[trackID] = track
	//
	// // Return a customized response based on correctness
	// var component templ.Component
	// var artistName string
	//
	// // Find the artist name
	// for _, artistID := range track.Track.ArtistsID {
	// 	if artist, ok := h.Cache.ArtistsCache[artistID]; ok {
	// 		if artistName != "" {
	// 			artistName += ", "
	// 		}
	// 		artistName += artist.Artist.Name
	// 	}
	// }
	//
	// // Get album image URL
	// albumURL := ""
	// if album, ok := h.Cache.AlbumsCache[track.Track.AlbumID]; ok {
	// 	if len(album.Album.Images) > 0 {
	// 		albumURL = album.Album.Images[0].URL
	// 	}
	// }
	//
	// // If the albumURL is empty or invalid, fetch it from the Spotify API
	// if albumURL == "" {
	// 	// Try to get the album image URL from the Spotify API
	// 	apiURL := fmt.Sprintf("https://api.spotify.com/v1/albums/%s", track.Track.AlbumID)
	// 	req, err := http.NewRequest("GET", apiURL, nil)
	// 	if err == nil {
	// 		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.AccessToken))
	// 		client := &http.Client{}
	// 		resp, err := client.Do(req)
	// 		if err == nil && resp.StatusCode == http.StatusOK {
	// 			defer resp.Body.Close()
	//
	// 			var albumResponse struct {
	// 				Images []struct {
	// 					URL string `json:"url"`
	// 				} `json:"images"`
	// 			}
	//
	// 			if err := json.NewDecoder(resp.Body).Decode(&albumResponse); err == nil {
	// 				if len(albumResponse.Images) > 0 {
	// 					albumURL = albumResponse.Images[0].URL
	//
	// 					// Update the cache with this album image
	// 					if album, ok := h.Cache.AlbumsCache[track.Track.AlbumID]; ok {
	// 						if len(album.Album.Images) == 0 {
	// 							images := make([]base.Image, len(albumResponse.Images))
	// 							for i, img := range albumResponse.Images {
	// 								images[i] = base.Image{URL: img.URL}
	// 							}
	// 							album.Album.Images = images
	// 							h.Cache.AlbumsCache[track.Track.AlbumID] = album
	// 						}
	// 					}
	// 				}
	// 			}
	// 		}
	// 	}
	//
	// 	// If we still don't have an album URL, use a placeholder
	// 	if albumURL == "" {
	// 		albumURL = "https://via.placeholder.com/300"
	// 	}
	// }
	//
	// if correct {
	// 	// Render correct guess component (to be created)
	// 	// For now, just render the track with a different style
	// 	component = templates.GuessOptionCard(track.Track.ID, track.Track.Name+" - CORRECT!", albumURL, artistName, true)
	//
	// 	// Skip to next song after a delay
	// 	go func() {
	// 		time.Sleep(3 * time.Second)
	// 		h.GameService.SkipSong()
	// 	}()
	// } else {
	// 	// Render incorrect guess component
	// 	component = templates.GuessOptionCard(track.Track.ID, track.Track.Name+" - Try again!", albumURL, artistName, false)
	// }
	//
	// component.Render(r.Context(), w)
}

// GuessTrack is a legacy handler still needed for the frontend
func (h *GameHandler) GuessTrack(w http.ResponseWriter, r *http.Request) {
	guess := r.FormValue("guess")
	if guess == "" {
		http.Error(w, "Guess is required", http.StatusBadRequest)
		return
	}

	result, err := h.GameService.UserGuess(guess)
	if err != nil {
		http.Error(w, "Guess user error", http.StatusBadRequest)
		return
	}
	w.Write([]byte(result))

}

// SkipSong handles skip button
// func (h *GameHandler) SkipSong(w http.ResponseWriter, r *http.Request) { }

// ClearQueue handles clearing the music player queue
// func (h *GameHandler) ClearQueue(w http.ResponseWriter, r *http.Request) { }
