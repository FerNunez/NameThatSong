package handlers

import (
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/services/spotify"
	"github.com/FerNunez/NameThatSong/web/templates"
)

type DebugSearchHandler struct {
	spotifyService spotify.SpotifyService
}

func NewDebugSearchHandler(spotifyService spotify.SpotifyService) *DebugSearchHandler {
	return &DebugSearchHandler{
		spotifyService: spotifyService,
	}
}

func (h *DebugSearchHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	
	var results interface{}
	var resultCount int
	var hasResults bool
	
	// If query provided, perform search
	if query != "" {
		// Use a debug user ID for testing
		userID := "debug-user"
		
		// Perform search (this will log cache hit/miss to console)
		tracks, err := h.spotifyService.SearchTracks(r.Context(), userID, query)
		if err != nil {
			// Handle error but still show the form
			results = "Error: " + err.Error()
			hasResults = true
		} else {
			results = tracks
			resultCount = len(tracks)
			hasResults = true
		}
	}
	
	// Render the debug search page
	component := templates.DebugSearch(query, results, resultCount, hasResults)
	component.Render(r.Context(), w)
}