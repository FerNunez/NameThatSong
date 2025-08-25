package handlers

import (
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/api/middleware"
	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
	"github.com/FerNunez/NameThatSong/web/templates"
	"github.com/go-chi/chi/v5"
)

type SearchHandler struct {
	spotifyService spotify.SpotifyService
}

func NewSearchHandler(ss spotify.SpotifyService) *SearchHandler {
	return &SearchHandler{
		spotifyService: ss,
	}
}

// GET /api/search/artist/{id}
func (h *SearchHandler) GetArtistItems(w http.ResponseWriter, r *http.Request) {
	logger.Info(r.Context(), "get artist items")

	_, ok := middleware.GetUser(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	searchArtistIDStr := chi.URLParam(r, "id")
	if searchArtistIDStr == "" {
		http.Error(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	// TODO:
	//h.spotifyService.SearchArtistsSongs()

	tracks := make([]models.TrackSearch, 0, 5)
	tracks = append(tracks, models.TrackSearch{
		ID:   "asd",
		Name: "TEST",
	})
	albums := make([]models.AlbumSearch, 0, 5)
	albums = append(albums, models.AlbumSearch{
		ID:   "asd",
		Name: "TEST",
	})

	component := templates.SearchResults(tracks, albums, []models.ArtistSearch{}, []models.PlaylistSearch{}, searchArtistIDStr, "")
	component.Render(r.Context(), w)
}

// GET /api/search/album/{id}
func (h *SearchHandler) GetAlbumItems(w http.ResponseWriter, r *http.Request) {
	logger.Info(r.Context(), "get album items")

	_, ok := middleware.GetUser(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	searchAlbumIDStr := chi.URLParam(r, "id")
	if searchAlbumIDStr == "" {
		http.Error(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	// TODO:
	//h.spotifyService.SearchAlbumsSongs()

	tracks := make([]models.TrackSearch, 0, 5)
	tracks = append(tracks, models.TrackSearch{
		ID:   "asd",
		Name: "TEST",
	})

	component := templates.SearchResults(tracks, []models.AlbumSearch{}, []models.ArtistSearch{}, []models.PlaylistSearch{}, searchAlbumIDStr, "")
	component.Render(r.Context(), w)
}
