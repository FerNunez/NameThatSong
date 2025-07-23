package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/api/middleware"
	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/services/playlist"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PlaylistHandler struct {
	playlistService playlist.PlaylistService
}

func NewPlaylistHandler(playlistService playlist.PlaylistService) *PlaylistHandler {
	return &PlaylistHandler{
		playlistService: playlistService,
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
