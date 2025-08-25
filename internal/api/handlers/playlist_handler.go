package handlers

import (
	"fmt"
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
		
		// Use GetPlaylistWithSongs to get accurate song count
		playlistWithSongs, err := h.playlistService.GetPlaylistWithSongs(r.Context(), playlist.ID, user.ID)
		songCount := 0
		if err == nil {
			songCount = len(playlistWithSongs.Songs)
		}
		
		templatePlaylists = append(templatePlaylists, templates.UserPlaylist{
			ID:         playlist.ID.String(),
			Name:       playlist.Name,
			TrackCount: songCount,
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

// GET /api/game/setup - Show game setup component
func (h *PlaylistHandler) ShowGameSetup(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	templates.GameSetupView().Render(r.Context(), w)
}

// GET /api/playlist-songs-empty - Show empty playlist state
func (h *PlaylistHandler) ShowPlaylistSongsEmpty(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	templates.PlaylistSongsEmpty().Render(r.Context(), w)
}

// GET /api/game/playlists - Get user playlists for game setup
func (h *PlaylistHandler) GetGamePlaylists(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's playlists from the service
	playlists, err := h.playlistService.GetUserPlaylists(r.Context(), user.ID)
	if err != nil {
		logger.Error(r.Context(), "failed to get user playlists",
			logger.F("user_id", user.ID.String()),
			logger.F("error", err))
		http.Error(w, "failed to get playlists", http.StatusInternalServerError)
		return
	}

	// Convert to template format
	templatePlaylists := make([]templates.UserPlaylist, len(playlists))
	for i, p := range playlists {
		templatePlaylists[i] = templates.UserPlaylist{
			ID:         p.ID.String(),
			Name:       p.Name,
			TrackCount: len(p.Songs),
			IsSpotify:  p.SpotifyPlaylistID != nil,
			SpotifyID: func() string {
				if p.SpotifyPlaylistID != nil {
					return *p.SpotifyPlaylistID
				}
				return ""
			}(),
		}
	}

	w.Header().Set("Content-Type", "text/html")
	templates.GamePlaylistSelection(templatePlaylists).Render(r.Context(), w)
}

// GET /api/playlist/{id}/songs - Get playlist details with songs for display
func (h *PlaylistHandler) GetPlaylistSongsView(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	playlistIDStr := chi.URLParam(r, "id")

	// Check if this is a Spotify playlist ID or local playlist UUID
	var playlist *models.Playlist
	var err error

	if playlistID, uuidErr := uuid.Parse(playlistIDStr); uuidErr == nil {
		// It's a UUID, get local playlist with songs
		playlist, err = h.playlistService.GetPlaylistWithSongs(r.Context(), playlistID, user.ID)
	} else {
		// It's likely a Spotify playlist ID, get it from Spotify service
		spotifyPlaylist, trackIDs, albumIDs, spotifyErr := h.spotifyService.FetchPlaylist(r.Context(), user.ID.String(), playlistIDStr)
		if spotifyErr != nil {
			logger.Error(r.Context(), "failed to get spotify playlist",
				logger.F("playlist_id", playlistIDStr),
				logger.F("user_id", user.ID.String()),
				logger.F("error", spotifyErr))
			http.Error(w, "Playlist not found", http.StatusNotFound)
			return
		}
		// Convert Spotify playlist to our playlist model
		playlist = &models.Playlist{
			ID:                uuid.New(), // Temporary ID for display
			Name:              spotifyPlaylist.Name,
			Description:       spotifyPlaylist.Description,
			IsPublic:          spotifyPlaylist.Public,
			SpotifyPlaylistID: &spotifyPlaylist.ID,
			Songs:             []models.PlaylistSong{}, // Will be populated below
		}

		for idx, trackID := range trackIDs {
			trackData, err := h.spotifyService.FetchTrack(r.Context(), user.ID.String(), trackID)
			if err != nil {
				logger.Error(r.Context(), "failed to get spotify track",
					logger.F("trackID", trackID),
					logger.F("error", err))
				http.Error(w, "Track not found", http.StatusNotFound)
				return
			}

			albumData, err := h.spotifyService.FetchAlbum(r.Context(), user.ID.String(), albumIDs[idx])
			playlist.Songs = append(playlist.Songs, models.PlaylistSong{
				ID:             uuid.New(), // Temporary ID
				PlaylistID:     playlist.ID,
				SpotifyTrackID: trackData.ID,
				Position:       idx + 1,
				TrackName:      trackData.Name,
				ArtistName:     trackData.GetPrimaryArtistName(),
				AlbumName:      albumData.Name,
				AlbumURL:       albumData.ImageURL,
				DurationMs:     trackData.DurationMs,
			})
		}

	}

	if err != nil {
		logger.Error(r.Context(), "failed to get playlist",
			logger.F("playlist_id", playlistIDStr),
			logger.F("user_id", user.ID.String()),
			logger.F("error", err))
		http.Error(w, "Playlist not found", http.StatusNotFound)
		return
	}

	// Convert playlist to template format
	playlistInfo := templates.PlaylistInfo{
		ID:          playlistIDStr,
		Name:        playlist.Name,
		Description: playlist.Description,
		IsSpotify:   playlist.SpotifyPlaylistID != nil,
	}

	// Set image URL if available
	if playlist.SpotifyPlaylistID != nil && h.spotifyService != nil {
		if spotifyPlaylist, _, _, err := h.spotifyService.FetchPlaylist(r.Context(), user.ID.String(), *playlist.SpotifyPlaylistID); err == nil {
			playlistInfo.ImageURL = spotifyPlaylist.ImageURL
			playlistInfo.Owner = spotifyPlaylist.OwnerDisplayName
		}
	}

	// Convert songs to template format
	var templateSongs []templates.PlaylistSong
	for _, song := range playlist.Songs {
		templateSong := templates.PlaylistSong{
			ID:       song.SpotifyTrackID,
			Title:    song.TrackName,
			Artist:   song.ArtistName,
			AlbumArt: song.AlbumURL, // Will need to fetch this from Spotify if needed
		}

		// Format duration
		if song.DurationMs > 0 {
			duration := song.DurationMs / 1000
			minutes := duration / 60
			seconds := duration % 60
			templateSong.Duration = fmt.Sprintf("%d:%02d", minutes, seconds)
		}

		templateSongs = append(templateSongs, templateSong)
	}

	// Calculate total duration
	var totalMs int
	for _, song := range playlist.Songs {
		totalMs += song.DurationMs
	}
	if totalMs > 0 {
		totalSeconds := totalMs / 1000
		totalMinutes := totalSeconds / 60
		hours := totalMinutes / 60
		minutes := totalMinutes % 60
		if hours > 0 {
			playlistInfo.TotalDuration = fmt.Sprintf("%d hr %d min", hours, minutes)
		} else {
			playlistInfo.TotalDuration = fmt.Sprintf("%d min", minutes)
		}
	}

	w.Header().Set("Content-Type", "text/html")
	templates.PlaylistDetailsView(playlistInfo, templateSongs).Render(r.Context(), w)
}

// GET /api/set-playlist-context - Set current playlist context for search
func (h *PlaylistHandler) SetPlaylistContext(w http.ResponseWriter, r *http.Request) {
	playlistID := r.URL.Query().Get("playlistId")
	if playlistID == "" {
		http.Error(w, "Missing playlist ID", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<input type="hidden" name="currentPlaylistId" id="current-playlist-context" value="%s"/>`, playlistID)
}

// POST /api/add-to-current-playlist - Add track to specified playlist
func (h *PlaylistHandler) AddToCurrentPlaylist(w http.ResponseWriter, r *http.Request) {
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

	trackID := r.FormValue("trackId")
	playlistIDStr := r.FormValue("playlistId")

	if trackID == "" {
		http.Error(w, "Missing track ID", http.StatusBadRequest)
		return
	}

	if playlistIDStr == "" {
		logger.Info(r.Context(), "no playlist context provided for track addition",
			logger.F("track_id", trackID),
			logger.F("user_id", user.ID))
		w.WriteHeader(http.StatusOK)
		return
	}

	playlistID, err := uuid.Parse(playlistIDStr)
	if err != nil {
		http.Error(w, "Invalid playlist ID", http.StatusBadRequest)
		return
	}

	// Create add song request
	req := models.AddSongRequest{
		SpotifyTrackID: trackID,
	}

	// Add track to playlist
	err = h.playlistService.AddSongToPlaylist(r.Context(), playlistID, user.ID, req)
	if err != nil {
		logger.Error(r.Context(), "failed to add track to playlist",
			logger.F("error", err),
			logger.F("track_id", trackID),
			logger.F("playlist_id", playlistIDStr),
			logger.F("user_id", user.ID))
		http.Error(w, "Failed to add track to playlist", http.StatusInternalServerError)
		return
	}

	logger.Info(r.Context(), "successfully added track to playlist",
		logger.F("track_id", trackID),
		logger.F("playlist_id", playlistIDStr),
		logger.F("user_id", user.ID))

	// Return success response with optional playlist refresh trigger
	w.Header().Set("HX-Trigger", fmt.Sprintf(`{"trackAdded": {"trackId": "%s", "playlistId": "%s"}}`, trackID, playlistIDStr))
	w.WriteHeader(http.StatusOK)
}
