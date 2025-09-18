package handlers

import (
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/api/middleware"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
	"github.com/FerNunez/NameThatSong/internal/services/events"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
	"github.com/FerNunez/NameThatSong/web/templates"
	"github.com/go-chi/chi/v5"
)

type ActionHandler struct {
	spotifyService spotify.SpotifyService
	eventBus       *events.EventBus
}

func NewActionHandler(spotifyService spotify.SpotifyService) *ActionHandler {
	return &ActionHandler{
		spotifyService: spotifyService,
		eventBus:       events.NewEventBus(),
	}
}

// POST /api/track/{id}/play - Create resource to play track
func (h *ActionHandler) PlayTrackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.Debug(r.Context(), "Playing Track")

	trackID := chi.URLParam(r, "id")
	if trackID == "" {
		http.Error(w, "Invalid track ID", http.StatusBadRequest)
		return
	}

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		templates.LocalPlaylistsList([]templates.UserPlaylist{}).Render(r.Context(), w)
		return
	}

	if err := h.spotifyService.PlaySong(ctx, user.ID.String(), trackID); err != nil {

		http.Error(w, "couldnt play track", http.StatusInternalServerError)
	}

}
