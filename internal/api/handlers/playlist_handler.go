package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/api/middleware"
	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
	"github.com/FerNunez/NameThatSong/internal/services/playlist"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
	"github.com/FerNunez/NameThatSong/web/templates"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PlaylistHandler struct {
	playlistService playlist.PlaylistService
	spotifyService  spotify.SpotifyService
}

func NewPlaylistHandler(playlistService playlist.PlaylistService) *PlaylistHandler {
	return &PlaylistHandler{
		playlistService: playlistService,
	}
}

func NewPlaylistHandlerWithSpotify(playlistService playlist.PlaylistService, spotifyService spotify.SpotifyService) *PlaylistHandler {
	return &PlaylistHandler{
		playlistService: playlistService,
		spotifyService:  spotifyService,
	}
}

// GET /playlists - Get user playlists
func (h *PlaylistHandler) GetUserPlaylists(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	playlists, err := h.playlistService.GetUserPlaylists(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "Failed to get playlists", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(playlists)
}

// POST /playlists - Create playlist
func (h *PlaylistHandler) CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.CreatePlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	playlist, err := h.playlistService.CreatePlaylist(r.Context(), user.ID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(playlist)
}

// GET /playlists/{id} - Get specific playlist
func (h *PlaylistHandler) GetPlaylist(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	playlistIDStr := chi.URLParam(r, "id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		http.Error(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	// Check if we should include songs
	includeSongs := r.URL.Query().Get("include_songs") == "true"

	var playlist *models.Playlist
	if includeSongs {
		playlist, err = h.playlistService.GetPlaylistWithSongs(r.Context(), playlistID, user.ID)
	} else {
		playlist, err = h.playlistService.GetPlaylist(r.Context(), playlistID, user.ID)
	}

	if err != nil {
		http.Error(w, "Playlist not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(playlist)
}

// PUT /playlists/{id} - Update playlist
func (h *PlaylistHandler) UpdatePlaylist(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	playlistIDStr := chi.URLParam(r, "id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		http.Error(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	var req models.UpdatePlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	playlist, err := h.playlistService.UpdatePlaylist(r.Context(), playlistID, user.ID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(playlist)
}

// DELETE /playlists/{id} - Delete playlist
func (h *PlaylistHandler) DeletePlaylist(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	playlistIDStr := chi.URLParam(r, "id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		http.Error(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	if err := h.playlistService.DeletePlaylist(r.Context(), playlistID, user.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /playlists/{id}/songs - Add song to playlist
func (h *PlaylistHandler) AddSongToPlaylist(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	playlistIDStr := chi.URLParam(r, "id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		http.Error(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	var req models.AddSongRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.playlistService.AddSongToPlaylist(r.Context(), playlistID, user.ID, req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// DELETE /playlists/{id}/songs/{songId} - Remove song
func (h *PlaylistHandler) RemoveSongFromPlaylist(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	playlistIDStr := chi.URLParam(r, "id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		http.Error(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	songIDStr := chi.URLParam(r, "songId")
	songID, err := uuid.Parse(songIDStr)
	if err != nil {
		http.Error(w, "Invalid song ID", http.StatusBadRequest)
		return
	}

	if err := h.playlistService.RemoveSongFromPlaylist(r.Context(), playlistID, user.ID, songID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PUT /playlists/{id}/songs/reorder - Reorder playlist songs
func (h *PlaylistHandler) ReorderPlaylistSongs(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	playlistIDStr := chi.URLParam(r, "id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		http.Error(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	var req models.ReorderSongsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.playlistService.ReorderPlaylistSongs(r.Context(), playlistID, user.ID, req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// POST /playlists/import - Import from Spotify
func (h *PlaylistHandler) ImportFromSpotify(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.ImportPlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	playlist, err := h.playlistService.ImportFromSpotify(r.Context(), user.ID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(playlist)
}

// GET /api/local-playlists - Get user's local playlists only
func (h *PlaylistHandler) GetLocalPlaylists(w http.ResponseWriter, r *http.Request) {
	logger.Info(r.Context(), "fetching local playlists")

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		templates.LocalPlaylistsList([]templates.UserPlaylist{}).Render(r.Context(), w)
		return
	}

	// Get user's local playlists (where spotify_playlist_id IS NULL for pure local, or IS NOT NULL for imported)
	playlists, err := h.playlistService.GetUserPlaylists(r.Context(), user.ID)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		templates.LocalPlaylistsList([]templates.UserPlaylist{}).Render(r.Context(), w)
		return
	}

	// Convert to template format - these are local playlists
	var templatePlaylists []templates.UserPlaylist
	for _, playlist := range playlists {
		spotifyID := ""
		if playlist.SpotifyPlaylistID != nil {
			spotifyID = *playlist.SpotifyPlaylistID
		}
		templatePlaylists = append(templatePlaylists, templates.UserPlaylist{
			ID:         playlist.ID.String(),
			Name:       playlist.Name,
			TrackCount: len(playlist.Songs), // This might need to be fetched separately
			IsSpotify:  playlist.SpotifyPlaylistID != nil,
			SpotifyID:  spotifyID,
		})
	}

	w.Header().Set("Content-Type", "text/html")
	templates.LocalPlaylistsList(templatePlaylists).Render(r.Context(), w)
}

// GET /api/spotify-playlists - Get user's Spotify playlists for display
func (h *PlaylistHandler) GetSpotifyPlaylists(w http.ResponseWriter, r *http.Request) {
	logger.Info(r.Context(), "fetching spotify playlists for display")

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		templates.SpotifyPlaylistsList([]templates.UserPlaylist{}).Render(r.Context(), w)
		return
	}

	// Get user's Spotify playlists
	spotifyPlaylists, err := h.spotifyService.GetUserPlaylists(r.Context(), user.ID.String())
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		templates.SpotifyPlaylistsList([]templates.UserPlaylist{}).Render(r.Context(), w)
		return
	}

	// Convert to template format - these are Spotify playlists for viewing
	var templatePlaylists []templates.UserPlaylist
	for _, playlist := range spotifyPlaylists {
		templatePlaylists = append(templatePlaylists, templates.UserPlaylist{
			ID:         playlist.ID,
			Name:       playlist.Name,
			TrackCount: playlist.TotalTracks,
			IsSpotify:  true,
			SpotifyID:  playlist.ID,
			ImageURL:   playlist.ImageURL,
		})
	}

	w.Header().Set("Content-Type", "text/html")
	templates.SpotifyPlaylistsList(templatePlaylists).Render(r.Context(), w)
}

// GET /api/import-spotify-playlists - Get available Spotify playlists for import
func (h *PlaylistHandler) GetSpotifyPlaylistsForImport(w http.ResponseWriter, r *http.Request) {
	logger.Info(r.Context(), "fetching spotify playlists")

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		templates.UserPlaylistsList([]templates.UserPlaylist{}).Render(r.Context(), w)
		return
	}

	// Get user's Spotify playlists
	spotifyPlaylists, err := h.spotifyService.GetUserPlaylists(r.Context(), user.ID.String())
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		templates.SpotifyImportError(err.Error()).Render(r.Context(), w)
		return
	}

	// Convert to template format - these are available for import
	var templatePlaylists []templates.UserPlaylist
	for _, playlist := range spotifyPlaylists {
		templatePlaylists = append(templatePlaylists, templates.UserPlaylist{
			ID:         playlist.ID,
			Name:       playlist.Name,
			TrackCount: playlist.TotalTracks,
			IsSpotify:  true,
			SpotifyID:  playlist.ID,
		})
	}

	w.Header().Set("Content-Type", "text/html")
	templates.SpotifyImportList(templatePlaylists).Render(r.Context(), w)
}

// POST /playlists/{id}/export - Export to Spotify
func (h *PlaylistHandler) ExportToSpotify(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	playlistIDStr := chi.URLParam(r, "id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		http.Error(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	var req models.ExportPlaylistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set the playlist ID from URL
	req.PlaylistID = playlistID

	spotifyPlaylistID, err := h.playlistService.ExportToSpotify(r.Context(), user.ID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]string{
		"spotify_playlist_id": spotifyPlaylistID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// POST /playlists/{id}/sync - Sync with Spotify
func (h *PlaylistHandler) SyncWithSpotify(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	playlistIDStr := chi.URLParam(r, "id")
	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		http.Error(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	if err := h.playlistService.SyncWithSpotify(r.Context(), playlistID, user.ID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PUT /api/spotify-playlists/{id}/update - Update single Spotify playlist data
func (h *PlaylistHandler) UpdateSpotifyPlaylist(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		templates.SpotifyPlaylistsList([]templates.UserPlaylist{}).Render(r.Context(), w)
		return
	}

	spotifyPlaylistID := chi.URLParam(r, "id")
	if spotifyPlaylistID == "" {
		http.Error(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	// Refresh this specific playlist from Spotify
	spotifyPlaylists, err := h.spotifyService.GetUserPlaylists(r.Context(), user.ID.String())
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		templates.SpotifyPlaylistsList([]templates.UserPlaylist{}).Render(r.Context(), w)
		return
	}

	// Convert to template format - return all playlists but with updated data
	var templatePlaylists []templates.UserPlaylist
	for _, playlist := range spotifyPlaylists {
		templatePlaylists = append(templatePlaylists, templates.UserPlaylist{
			ID:         playlist.ID,
			Name:       playlist.Name,
			TrackCount: playlist.TotalTracks,
			IsSpotify:  true,
			SpotifyID:  playlist.ID,
		})
	}

	w.Header().Set("Content-Type", "text/html")
	templates.SpotifyPlaylistsList(templatePlaylists).Render(r.Context(), w)
}

// PUT /api/spotify-playlists/refresh - Refresh all Spotify playlists
func (h *PlaylistHandler) RefreshSpotifyPlaylists(w http.ResponseWriter, r *http.Request) {
	// This is essentially the same as GetSpotifyPlaylists but as a PUT endpoint
	h.GetSpotifyPlaylists(w, r)
}

// GET /api/playlist/create-form - Show playlist creation form
func (h *PlaylistHandler) ShowCreatePlaylistForm(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	templates.CreatePlaylistForm().Render(r.Context(), w)
}

// POST /api/playlist/create - Create new playlist and show it
func (h *PlaylistHandler) CreateAndShowPlaylist(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	description := r.FormValue("description")
	isPublic := r.FormValue("is_public") == "on"

	// Create playlist request
	req := models.CreatePlaylistRequest{
		Name:            name,
		Description:     description,
		IsPublic:        isPublic,
		SyncWithSpotify: false, // Local playlist
	}

	// Create playlist
	playlist, err := h.playlistService.CreatePlaylist(r.Context(), user.ID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Convert to template format and show new playlist view
	templatePlaylist := templates.UserPlaylist{
		ID:         playlist.ID.String(),
		Name:       playlist.Name,
		TrackCount: 0,
		IsSpotify:  false,
		SpotifyID:  "",
	}

	// Refresh local playlists sidebar by adding HX-Trigger header
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("HX-Trigger", "refreshLocalPlaylists")
	templates.NewPlaylistView(templatePlaylist).Render(r.Context(), w)
}

// GET /api/playlist/cancel-create - Cancel playlist creation
func (h *PlaylistHandler) CancelCreatePlaylist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	templates.PlaylistSongsEmpty().Render(r.Context(), w)
}
