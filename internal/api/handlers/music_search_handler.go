package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/FerNunez/NameThatSong/internal/api/middleware"
	m "github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
	"github.com/FerNunez/NameThatSong/web/templates"
	"github.com/go-chi/chi/v5"
)

// MusicSearchHandler handles specific music item searches (artist/album detail searches)
type MusicSearchHandler struct {
	spotifyService spotify.SpotifyService
}

func NewMusicSearchHandler(ss spotify.SpotifyService) *MusicSearchHandler {
	return &MusicSearchHandler{
		spotifyService: ss,
	}
}

// ServeHTTP handles the search API endpoint
func (h *MusicSearchHandler) SearchAll(w http.ResponseWriter, r *http.Request) {
	// Get the search query and playlist context
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	currentPlaylistID := r.URL.Query().Get("currentPlaylistId")
	logger.Debug(r.Context(), "search request received", logger.F("query", query), logger.F("currentPlaylistId", currentPlaylistID))

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

	// Use SearchAll for efficient single API call
	logger.Debug(r.Context(), "calling spotify search all", logger.F("userID", user.ID.String()), logger.F("query", query))
	results, err := h.spotifyService.SearchAll(r.Context(), user.ID.String(), query)
	if err != nil {
		logger.Error(r.Context(), "search all failed", logger.F("error", err), logger.F("query", query))
		component := templates.NoResultsFound(query)
		component.Render(r.Context(), w)
		return
	}

	logger.Debug(r.Context(), "search all completed", 
		logger.F("tracks_found", len(results.Tracks)),
		logger.F("albums_found", len(results.Albums)),
		logger.F("artists_found", len(results.Artists)),
		logger.F("playlists_found", len(results.Playlists)))

	// Apply result limiting to match previous behavior
	const maxResults = 5
	tracks := results.Tracks
	if len(tracks) > maxResults {
		tracks = tracks[:maxResults]
	}

	albums := results.Albums
	if len(albums) > maxResults {
		albums = albums[:maxResults]
	}

	artists := results.Artists
	if len(artists) > maxResults {
		artists = artists[:maxResults]
	}

	playlists := results.Playlists
	if len(playlists) > maxResults {
		playlists = playlists[:maxResults]
	}

	logger.Info(r.Context(), "search completed",
		logger.F("query", query),
		logger.F("#tracks", len(tracks)),
		logger.F("#albums", len(albums)),
		logger.F("#artists", len(artists)),
		logger.F("#playlists", len(playlists)))

	// Render the search results
	logger.Debug(r.Context(), "rendering search results", 
		logger.F("final_tracks", len(tracks)),
		logger.F("final_albums", len(albums)),
		logger.F("final_artists", len(artists)),
		logger.F("final_playlists", len(playlists)))
	component := templates.SearchResults(tracks, albums, artists, playlists, query, currentPlaylistID)
	component.Render(r.Context(), w)
}

// GET /api/music-search/artist/{id}
func (h *MusicSearchHandler) SearchArtistItems(w http.ResponseWriter, r *http.Request) {
	logger.Info(r.Context(), "get artist items")

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	searchArtistIDStr := chi.URLParam(r, "id")
	logger.Debug(r.Context(), "artist search request", logger.F("artistID", searchArtistIDStr))
	if searchArtistIDStr == "" {
		http.Error(w, "Invalid artist ID", http.StatusBadRequest)
		return
	}

	// TODO: Maybe add AlbumIDs into the ArtistData. But for this another call to api must be done :(
	// fetch Artist data
	logger.Debug(r.Context(), "fetching artist data", logger.F("artistID", searchArtistIDStr), logger.F("userID", user.ID.String()))
	_, err := h.spotifyService.FetchArtist(r.Context(), user.ID.String(), m.SpotifyID(searchArtistIDStr))
	if err != nil {
		logger.Error(r.Context(), "couldn't get artist", logger.F("artistID", searchArtistIDStr), logger.F("err", err))
		http.Error(w, "Failed to fetch artist", http.StatusInternalServerError)
		return
	}

	tracks := make([]m.TrackSearch, 0, 5)
	tracks = append(tracks, m.TrackSearch{
		ID:   "asd",
		Name: "TEST",
	})
	albums := make([]m.AlbumSearch, 0, 5)
	albums = append(albums, m.AlbumSearch{
		ID:   "asd",
		Name: "TEST",
	})

	component := templates.SearchResults(tracks, albums, []m.ArtistSearch{}, []m.PlaylistSearch{}, searchArtistIDStr, "")
	component.Render(r.Context(), w)
}

// GET /api/search/album/{id}
func (h *MusicSearchHandler) SearchAlbumItems(w http.ResponseWriter, r *http.Request) {
	logger.Info(r.Context(), "get album items")

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		logger.Error(r.Context(), "no user")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	searchAlbumIDStr := chi.URLParam(r, "id")
	logger.Debug(r.Context(), "album search request", logger.F("albumID", searchAlbumIDStr))
	if searchAlbumIDStr == "" {
		logger.Error(r.Context(), "empty id search")
		http.Error(w, "Invalid album ID", http.StatusBadRequest)
		return
	}

	// fetch albumData data
	logger.Debug(r.Context(), "fetching album data", logger.F("albumID", searchAlbumIDStr), logger.F("userID", user.ID.String()))
	albumData, err := h.spotifyService.FetchAlbum(r.Context(), user.ID.String(), m.SpotifyID(searchAlbumIDStr))
	if err != nil {
		logger.Error(r.Context(), "couldn't get album", logger.F("albumID", searchAlbumIDStr), logger.F("err", err))
		http.Error(w, "Failed to fetch album", http.StatusInternalServerError)
		return
	}

	// fetch batch tracks
	logger.Debug(r.Context(), "fetching batch tracks", logger.F("trackCount", len(albumData.TrackIDs)), logger.F("albumID", searchAlbumIDStr))
	tracksData, err := h.spotifyService.FetchMultipleTracks(r.Context(), user.ID.String(), albumData.TrackIDs)
	if err != nil {
		logger.Error(r.Context(), "couldn't get tracks batch for album", logger.F("albumID", searchAlbumIDStr), logger.F("err", err))
		http.Error(w, "Failed to fetch album tracks", http.StatusInternalServerError)
		return
	}

	// batch artists
	artistIDs := make([]m.SpotifyID, 0)
	for _, track := range tracksData {
		artistIDs = append(artistIDs, track.ArtistIDs...)
	}

	// get all batch artist
	logger.Debug(r.Context(), "fetching batch artists", logger.F("artistCount", len(artistIDs)), logger.F("albumID", searchAlbumIDStr))
	artistsData, err := h.spotifyService.FetchMultipleArtists(r.Context(), user.ID.String(), artistIDs)
	if err != nil {
		logger.Error(r.Context(), "couldn't get artists batch for album", logger.F("albumID", searchAlbumIDStr), logger.F("err", err))
		http.Error(w, "Failed to fetch album artists", http.StatusInternalServerError)
		return
	}

	// Map SpotifyID into ArtistID
	artistIDtoDataMap := make(map[m.SpotifyID]*m.ArtistData)
	for _, artistData := range artistsData {
		artistIDtoDataMap[artistData.ID] = &artistData
	}

	tracks := make([]m.TrackSearch, 0, len(tracksData))
	logger.Debug(r.Context(), "processing tracks for album", logger.F("trackCount", len(tracksData)), logger.F("albumID", searchAlbumIDStr))
	for _, track := range tracksData {

		// Find the list of artists
		artistIDList := make([]string, 0, len(track.ArtistIDs))
		for _, artistID := range track.ArtistIDs {
			if artistData, ok := artistIDtoDataMap[artistID]; ok {
				artistIDList = append(artistIDList, artistData.Name)
			}
		}

		tracks = append(tracks, m.TrackSearch{
			ID:            string(track.ID),
			Name:          track.Name,
			Popularity:    track.Popularity,
			DurationMs:    track.DurationMs,
			Explicit:      track.Explicit,
			ArtistNames:   artistIDList,
			AlbumName:     albumData.Name,
			AlbumImageURL: albumData.ImageURL,
		})
	}

	logger.Debug(r.Context(), "rendering album tracks results", logger.F("finalTrackCount", len(tracks)), logger.F("albumID", searchAlbumIDStr))
	component := templates.SearchResults(tracks, []m.AlbumSearch{}, []m.ArtistSearch{}, []m.PlaylistSearch{}, searchAlbumIDStr, "")
	component.Render(r.Context(), w)
}

