package handlers

import (
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/api/middleware"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
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
	logger.Info(r.Context(), "debug search handler accessed")
	
	query := r.URL.Query().Get("q")
	
	var results interface{}
	var resultCount int
	var hasResults bool
	
	// If query provided, perform search
	if query != "" {
		// Get current user from context
		user, ok := middleware.GetUser(r.Context())
		if !ok {
			logger.Warn(r.Context(), "debug search attempted without authenticated user")
			results = "Error: User not authenticated. Please log in to perform searches."
			hasResults = true
		} else {
			logger.Info(r.Context(), "performing debug search",
				logger.F("user_id", user.ID.String()),
				logger.F("query", query))
			
			// Perform search with authenticated user
			tracks, err := h.spotifyService.SearchTracks(r.Context(), user.ID.String(), query)
			if err != nil {
				logger.Error(r.Context(), "debug search failed",
					logger.F("user_id", user.ID.String()),
					logger.F("query", query),
					logger.F("error", err))
				// Handle error but still show the form
				results = "Error: " + err.Error()
				hasResults = true
			} else {
				logger.Info(r.Context(), "debug search completed successfully",
					logger.F("user_id", user.ID.String()),
					logger.F("query", query),
					logger.F("results_count", len(tracks)))
				results = tracks
				resultCount = len(tracks)
				hasResults = true
			}
		}
	}
	
	// Render the debug search page
	component := templates.DebugSearch(query, results, resultCount, hasResults)
	component.Render(r.Context(), w)
}