package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/FerNunez/NameThatSong/internal/api/middleware"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
	"github.com/FerNunez/NameThatSong/web/templates"
	"github.com/go-chi/chi/v5"

	m "github.com/FerNunez/NameThatSong/internal/models"
)

// SimpleSearchHandler handles the simplified HTMX search interface
type SimpleSearchHandler struct {
	spotifyService spotify.SpotifyService
}

// NewSimpleSearchHandler creates a new simplified search handler
func NewSimpleSearchHandler(ss spotify.SpotifyService) *SimpleSearchHandler {
	return &SimpleSearchHandler{
		spotifyService: ss,
	}
}

// ServeHTTP handles the search API endpoint
func (h *SimpleSearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Get the search query
	query := strings.TrimSpace(r.URL.Query().Get("query"))

	// If query is empty or too short, return empty state
	if query == "" {
		component := templates.EmptySearchState()
		component.Render(r.Context(), w)
		return
	}

	if len(query) < 2 {
		component := templates.NoResultsFound(query)
		component.Render(r.Context(), w)
		return
	}

	// Get authenticated user
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		logger.Warn(r.Context(), "search attempted without authenticated user")
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Set reasonable limits for search results
	const maxResults = 5

	// Create channels for concurrent search operations
	tracksChan := make(chan []m.TrackSearch)
	albumsChan := make(chan []m.AlbumSearch)
	artistsChan := make(chan []m.ArtistSearch)
	playlistsChan := make(chan []m.PlaylistSearch)
	errorsChan := make(chan error, 4)

	// Search tracks concurrently
	go func() {
		tracks, err := h.spotifyService.SearchTracks(r.Context(), user.ID.String(), query)
		if err != nil {
			logger.Error(r.Context(), "track search failed", logger.F("error", err), logger.F("query", query))
			errorsChan <- err
			tracksChan <- []m.TrackSearch{}
			return
		}

		// Limit results
		if len(tracks) > maxResults {
			tracks = tracks[:maxResults]
		}
		tracksChan <- tracks
	}()

	// Search albums concurrently
	go func() {
		albums, err := h.spotifyService.SearchAlbums(r.Context(), user.ID.String(), query)
		if err != nil {
			logger.Error(r.Context(), "album search failed", logger.F("error", err), logger.F("query", query))
			errorsChan <- err
			albumsChan <- []m.AlbumSearch{}
			return
		}

		// Limit results
		if len(albums) > maxResults {
			albums = albums[:maxResults]
		}
		albumsChan <- albums
	}()

	// Search artists concurrently
	go func() {
		artists, err := h.spotifyService.SearchArtists(r.Context(), user.ID.String(), query)
		if err != nil {
			logger.Error(r.Context(), "artist search failed", logger.F("error", err), logger.F("query", query))
			errorsChan <- err
			artistsChan <- []m.ArtistSearch{}
			return
		}

		// Limit results
		if len(artists) > maxResults {
			artists = artists[:maxResults]
		}
		artistsChan <- artists
	}()

	// Search playlists concurrently
	go func() {
		playlists, err := h.spotifyService.SearchPlaylists(r.Context(), user.ID.String(), query)
		if err != nil {
			logger.Error(r.Context(), "playlist search failed", logger.F("error", err), logger.F("query", query))
			errorsChan <- err
			playlistsChan <- []m.PlaylistSearch{}
			return
		}

		// Limit results
		if len(playlists) > maxResults {
			playlists = playlists[:maxResults]
		}
		playlistsChan <- playlists
	}()

	// Collect results with timeout
	var (
		tracks    []m.TrackSearch
		albums    []m.AlbumSearch
		artists   []m.ArtistSearch
		playlists []m.PlaylistSearch
	)

	timeout := time.After(3 * time.Second) // Reduced timeout for better UX
	completed := 0

	for completed < 4 {
		select {
		case tracks = <-tracksChan:
			completed++
		case albums = <-albumsChan:
			completed++
		case artists = <-artistsChan:
			completed++
		case playlists = <-playlistsChan:
			completed++
		case err := <-errorsChan:
			logger.Warn(r.Context(), "search operation failed", logger.F("error", err), logger.F("query", query))
			completed++
		case <-timeout:
			logger.Warn(r.Context(), "search timeout", logger.F("query", query), logger.F("completed", completed))
			// Continue with partial results
			goto renderResults
		}
	}

renderResults:
	logger.Info(r.Context(), "search completed",
		logger.F("query", query),
		logger.F("tracks", len(tracks)),
		logger.F("albums", len(albums)),
		logger.F("artists", len(artists)),
		logger.F("playlists", len(playlists)))

	// Render the search results
	component := templates.SimpleSearchResults(tracks, albums, artists, playlists, query)
	component.Render(r.Context(), w)
}

// SearchPageHandler removed - search is now integrated into modern UI sidebar

// Action handlers for the buttons (placeholder implementations)
type ActionHandler struct {
	ss spotify.SpotifyService
}

func NewActionHandler(ss spotify.SpotifyService) *ActionHandler {
	return &ActionHandler{
		ss: ss,
	}
}

// AddTrackHandler handles adding a track to playlist
func (h *ActionHandler) AddTrackHandler(w http.ResponseWriter, r *http.Request) {
	trackID := r.FormValue("trackId")
	logger.Info(r.Context(), "track add requested", logger.F("trackId", trackID))

	// TODO: Implement actual add to playlist logic
	w.Header().Set("HX-Trigger", "track-added")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Track %s added to playlist", trackID)
}

// PlayTrackHandler handles playing a track
func (h *ActionHandler) PlayTrackHandler(w http.ResponseWriter, r *http.Request) {
	trackID := r.FormValue("trackId")
	logger.Info(r.Context(), "track play requested", logger.F("trackId", trackID))

	// TODO: Implement act&ual play logic
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		logger.Warn(r.Context(), "debug play track attempted without authenticated user")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := h.ss.PlaySong(r.Context(), user.ID.String(), trackID)
	if err != nil {
		logger.Warn(r.Context(), "debug couldnt play track", logger.F("err", err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.Header().Set("HX-Trigger", "track-playing")
	w.WriteHeader(http.StatusOK)
}

// AddAlbumHandler handles adding an album to playlist
func (h *ActionHandler) AddAlbumHandler(w http.ResponseWriter, r *http.Request) {
	albumID := r.FormValue("albumId")
	logger.Info(r.Context(), "album add requested", logger.F("albumId", albumID))

	// TODO: Implement actual add album logic
	w.Header().Set("HX-Trigger", "album-added")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Album added to playlist", albumID)
}

// PlayAlbumHandler handles playing an album
func (h *ActionHandler) PlayAlbumHandler(w http.ResponseWriter, r *http.Request) {
	albumID := r.FormValue("albumId")
	logger.Info(r.Context(), "album play requested", logger.F("albumId", albumID))

	// TODO: Implement actual play album logic
	w.Header().Set("HX-Trigger", "album-playing")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Playing album %s", albumID)
}

// PlayArtistHandler handles playing an artist's top tracks
func (h *ActionHandler) PlayArtistHandler(w http.ResponseWriter, r *http.Request) {
	artistID := r.FormValue("artistId")
	logger.Info(r.Context(), "artist play requested", logger.F("artistId", artistID))

	// TODO: Implement actual play artist logic
	w.Header().Set("HX-Trigger", "artist-playing")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Playing top tracks for artist %s", artistID)
}

// PlayPlaylistHandler handles playing a playlist
func (h *ActionHandler) PlayPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	playlistName := r.FormValue("playlistName")
	logger.Info(r.Context(), "playlist play requested", logger.F("playlistName", playlistName))

	// TODO: Implement actual play playlist logic
	w.Header().Set("HX-Trigger", "playlist-playing")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Playing playlist %s", playlistName)
}

// ArtistDetailHandler handles showing artist details
type ArtistDetailHandler struct {
	spotifyService spotify.SpotifyService
}

func NewArtistDetailHandler(ss spotify.SpotifyService) *ArtistDetailHandler {
	return &ArtistDetailHandler{spotifyService: ss}
}

func (h *ArtistDetailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	artistID := chi.URLParam(r, "artistId")
	if artistID == "" {
		http.Error(w, "Artist ID required", http.StatusBadRequest)
		return
	}

	logger.Info(r.Context(), "artist detail requested", logger.F("artistId", artistID))

	// TODO: Implement artist detail view
	// For now, just return a simple message
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<div class='text-white p-4'>Artist details for ID: %s (Coming soon...)</div>", artistID)
}
