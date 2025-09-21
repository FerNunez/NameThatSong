package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/FerNunez/NameThatSong/internal/api/middleware"
	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
	"github.com/FerNunez/NameThatSong/internal/services/events"
	"github.com/FerNunez/NameThatSong/internal/services/playlist"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
	"github.com/FerNunez/NameThatSong/web/templates"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type PlaylistHandler struct {
	playlistService playlist.Service
	spotifyService  spotify.SpotifyService
	eventBus        *events.EventBus
}

func NewPlaylistHandler(playlistService playlist.Service) *PlaylistHandler {
	return &PlaylistHandler{
		playlistService: playlistService,
		eventBus:        events.NewEventBus(),
	}
}

func NewPlaylistHandlerWithSpotify(playlistService playlist.Service, spotifyService spotify.SpotifyService) *PlaylistHandler {
	return &PlaylistHandler{
		playlistService: playlistService,
		spotifyService:  spotifyService,
		eventBus:        events.NewEventBus(),
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

	localPlaylists, err := h.playlistService.GetUserPlaylists(r.Context(), user.ID)
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		templates.LocalPlaylistsList([]templates.UserPlaylist{}).Render(r.Context(), w)
		return
	}

	// Convert to template format - these are local playlists
	var templatePlaylists []templates.UserPlaylist
	for _, p := range localPlaylists {

		var imgUrl string
		if p.ImageURL == nil {
			imgUrl = ""
		} else {
			imgUrl = *p.ImageURL
		}
		var spotifyID string
		if p.SpotifyPlaylistID == nil {
			spotifyID = ""
		} else {
			spotifyID = *p.SpotifyPlaylistID
		}

		templatePlaylists = append(templatePlaylists, templates.UserPlaylist{
			ID:         p.ID.String(),
			Name:       p.Name,
			TrackCount: len(p.Tracks),
			IsSpotify:  p.SpotifyPlaylistID != nil,
			SpotifyID:  spotifyID,
			ImageURL:   imgUrl,
		})
	}

	w.Header().Set("Content-Type", "text/html")
	templates.LocalPlaylistsList(templatePlaylists).Render(r.Context(), w)
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
	spotifyPlaylists, err := h.spotifyService.FetchUserSpotifyPlaylistsVersion(r.Context(), user.ID.String())
	if err != nil {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		templates.SpotifyImportError(err.Error()).Render(r.Context(), w)
		return
	}

	// Import playlists in background to avoid blocking the response
	go func() {
		// Use background context to avoid cancellation when HTTP request completes
		ctx := context.Background()
		importedCount := 0
		for _, spotifyPlaylist := range spotifyPlaylists {
			_, err := h.playlistService.ImportFromSpotify(ctx, user.ID, models.ImportPlaylistRequest{
				SpotifyPlaylistID: string(spotifyPlaylist.ID),
				SnapshotID:        spotifyPlaylist.SnapshotID,
			})
			if err != nil {
				logger.Error(ctx, "Couldnt import playlist", logger.F("error", err))
			} else {
				importedCount++
			}
		}

		// Import completed - events will be emitted by service layer
	}()

	// Convert to template format - these are available for import
	var templatePlaylists []templates.UserPlaylist
	for _, playlist := range spotifyPlaylists {
		templatePlaylists = append(templatePlaylists, templates.UserPlaylist{
			ID:         string(playlist.ID),
			Name:       "Spotify Playlist", // TODO: Fetch full playlist data for name
			TrackCount: 0,                  // TODO: Fetch full playlist data for track count
			IsSpotify:  true,
			SpotifyID:  string(playlist.ID),
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
	spotifyPlaylists, err := h.spotifyService.FetchUserSpotifyPlaylistsVersion(r.Context(), user.ID.String())
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
			ID:         string(playlist.ID),
			Name:       "Spotify Playlist", // TODO: Fetch full playlist data for name
			TrackCount: 0,                  // TODO: Fetch full playlist data for track count
			IsSpotify:  true,
			SpotifyID:  string(playlist.ID),
		})
	}

	w.Header().Set("Content-Type", "text/html")
	templates.SpotifyPlaylistsList(templatePlaylists).Render(r.Context(), w)
}

// PUT /api/spotify-playlists/refresh - Refresh all Spotify playlists
func (h *PlaylistHandler) RefreshSpotifyPlaylists(w http.ResponseWriter, r *http.Request) {
	// This is essentially the same as GetSpotifyPlaylists but as a PUT endpoint
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

	logger.Debug(r.Context(), "CreateAndShowPlaylist")

	name := r.FormValue("name")
	description := r.FormValue("description")
	isPublic := r.FormValue("is_public") == "on"
	logger.Debug(r.Context(), "CreateAndShowPlaylist with", logger.F("name", name), logger.F("description", description), logger.F("is_public", isPublic))

	// Create playlist request
	req := models.CreatePlaylistRequest{
		Name:        name,
		Description: description,
		IsPublic:    isPublic,
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
			TrackCount: 0, // TODO: Fetch track count separately if needed
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
	var enrichedSongs []*models.PlaylistTrackWithDetails
	var err error

	if playlistID, uuidErr := uuid.Parse(playlistIDStr); uuidErr == nil {
		// It's a UUID, get local playlist with enriched song data (album/artist names via 3-tier caching)
		enrichedSongs, err = h.playlistService.GetPlaylistSongsWithDetails(r.Context(), user.ID.String(), playlistID)
	} else {
		fmt.Println("Spotify playlist not yet implemented")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
	// } else {
	// 	// It's likely a Spotify playlist ID, get it from Spotify service
	// 	spotifyPlaylist, trackIDs, albumIDs, spotifyErr := h.spotifyService.FetchPlaylist(r.Context(), user.ID.String(), playlistIDStr)
	// 	if spotifyErr != nil {
	// 		logger.Error(r.Context(), "failed to get spotify playlist",
	// 			logger.F("playlist_id", playlistIDStr),
	// 			logger.F("user_id", user.ID.String()),
	// 			logger.F("error", spotifyErr))
	// 		http.Error(w, "Playlist not found", http.StatusNotFound)
	// 		return
	// 	}
	// 	// Convert Spotify playlist to our playlist model
	// 	playlist = &models.Playlist{
	// 		ID:                uuid.New(), // Temporary ID for display
	// 		Name:              spotifyPlaylist.Name,
	// 		Description:       spotifyPlaylist.Description,
	// 		IsPublic:          spotifyPlaylist.Public,
	// 		SpotifyPlaylistID: &spotifyPlaylist.ID,
	// 		Songs:             []models.Song{}, // Will be populated below
	// 	}
	//
	// 	for idx, trackID := range trackIDs {
	// 		trackData, err := h.spotifyService.FetchTrack(r.Context(), user.ID.String(), trackID)
	// 		if err != nil {
	// 			logger.Error(r.Context(), "failed to get spotify track",
	// 				logger.F("trackID", trackID),
	// 				logger.F("error", err))
	// 			http.Error(w, "Track not found", http.StatusNotFound)
	// 			return
	// 		}
	//
	// 		albumData, err := h.spotifyService.FetchAlbum(r.Context(), user.ID.String(), albumIDs[idx])
	// 		playlist.Songs = append(playlist.Songs, models.Song{
	// 			ID:             uuid.New(), // Temporary ID
	// 			PlaylistID:     playlist.ID,
	// 			SpotifyTrackID: trackData.ID,
	// 			Position:       idx + 1,
	// 			TrackName:      trackData.Name,
	// 			ArtistName:     trackData.GetPrimaryArtistName(),
	// 			AlbumName:      albumData.Name,
	// 			AlbumURL:       albumData.ImageURL,
	// 			DurationMs:     trackData.DurationMs,
	// 		})
	// 	}
	//
	// }

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
		Name:        "Playlist Name",
		Description: "Playlist Descp",
		IsSpotify:   false,
	}

	// // Set image URL if available
	// if playlist.SpotifyPlaylistID != nil && h.spotifyService != nil {
	// 	if spotifyPlaylist, _, _, err := h.spotifyService.FetchPlaylist(r.Context(), user.ID.String(), *playlist.SpotifyPlaylistID); err == nil {
	// 		playlistInfo.ImageURL = spotifyPlaylist.ImageURL
	// 		playlistInfo.Owner = spotifyPlaylist.OwnerDisplayName
	// 	}
	// }
	//
	// Convert enriched songs to template format
	var templateSongs []templates.PlaylistSong
	for _, song := range enrichedSongs {
		templateSong := templates.PlaylistSong{
			ID:    song.SpotifyTrackID,
			Title: song.TrackName,
			Artist: func() string {
				if len(song.ArtistNames) > 0 {
					// Join multiple artist names with comma
					return strings.Join(song.ArtistNames, ", ")
				}
				return "Unknown Artist"
			}(),
			AlbumArt: song.AlbumImageUrl, // Rich album image URL from 3-tier caching
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
	logger.Info(r.Context(), "adding track to playlist",
		logger.F("track_id", trackID),
		logger.F("playlist_id", playlistIDStr),
		logger.F("user_id", user.ID))

	// Add track to playlist
	err = h.playlistService.AddSongToPlaylist(r.Context(), user.ID.String(), playlistID, req)
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

// GET /api/events/playlist-updates - Server-Sent Events for playlist updates
func (h *PlaylistHandler) HandlePlaylistEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Add panic recovery for production safety
	defer func() {
		if r := recover(); r != nil {
			logger.Error(ctx, "SSE handler panic recovered", logger.F("panic", r))
		}
	}()

	user, ok := middleware.GetUser(ctx)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	logger.Info(r.Context(), "SSE connection established", logger.F("user_id", user.ID))

	// Check if ResponseWriter supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Subscribe to all playlist events for this user
	eventChan := h.eventBus.SubscribeUser(user.ID)
	defer h.eventBus.UnsubscribeUser(user.ID, eventChan) // Clean up on exit

	// Send initial connection event
	fmt.Fprintf(w, "event: connected\n")
	fmt.Fprintf(w, "data: {\"status\": \"connected\"}\n\n")
	flusher.Flush()
	logger.Info(ctx, "SSE initial connection event sent", logger.F("user_id", user.ID))

	// Keep connection alive and send events
	for {
		select {
		case <-ctx.Done():
			logger.Info(ctx, "SSE connection closed by client", logger.F("user_id", user.ID))
			return
		case event, ok := <-eventChan:
			if !ok {
				logger.Info(ctx, "Event channel closed", logger.F("user_id", user.ID))
				return
			}

			// Send event to client
			switch event.Type {
			case events.PlaylistImportCompleted:
				logger.Info(ctx, "Sending playlist_created SSE event", logger.F("user_id", user.ID))
				if dataBytes, err := json.Marshal(event.Data); err == nil {
					fmt.Fprintf(w, "event: playlist_created\n")
					fmt.Fprintf(w, "data: %s\n\n", dataBytes)
				}
			case events.PlaylistCreated:
				logger.Info(ctx, "Sending playlist_created SSE event", logger.F("user_id", user.ID))
				if dataBytes, err := json.Marshal(event.Data); err == nil {
					fmt.Fprintf(w, "event: playlist_created\n")
					fmt.Fprintf(w, "data: %s\n\n", dataBytes)
				}
			case events.PlaylistUpdated:
				logger.Info(ctx, "Sending playlist_updated SSE event", logger.F("user_id", user.ID))
				if dataBytes, err := json.Marshal(event.Data); err == nil {
					fmt.Fprintf(w, "event: playlist_updated\n")
					fmt.Fprintf(w, "data: %s\n\n", dataBytes)
				}
			case events.PlaylistDeleted:
				logger.Info(ctx, "Sending playlist_deleted SSE event", logger.F("user_id", user.ID))
				if dataBytes, err := json.Marshal(event.Data); err == nil {
					fmt.Fprintf(w, "event: playlist_deleted\n")
					fmt.Fprintf(w, "data: %s\n\n", dataBytes)
				}
			case events.PlaylistSongAdded:
				logger.Info(ctx, "Sending playlist_song_added SSE event", logger.F("user_id", user.ID))
				if dataBytes, err := json.Marshal(event.Data); err == nil {
					fmt.Fprintf(w, "event: playlist_song_added\n")
					fmt.Fprintf(w, "data: %s\n\n", dataBytes)
				}
			case events.PlaylistSongRemoved:
				logger.Info(ctx, "Sending playlist_song_removed SSE event", logger.F("user_id", user.ID))
				if dataBytes, err := json.Marshal(event.Data); err == nil {
					fmt.Fprintf(w, "event: playlist_song_removed\n")
					fmt.Fprintf(w, "data: %s\n\n", dataBytes)
				}
			case events.PlaylistSyncCompleted:
				logger.Info(ctx, "Sending playlist_sync_completed SSE event", logger.F("user_id", user.ID))
				if dataBytes, err := json.Marshal(event.Data); err == nil {
					fmt.Fprintf(w, "event: playlist_sync_completed\n")
					fmt.Fprintf(w, "data: %s\n\n", dataBytes)
				}
			}
			flusher.Flush()

		case <-time.After(30 * time.Second):
			// Send heartbeat to keep connection alive
			fmt.Fprintf(w, "event: heartbeat\n")
			fmt.Fprintf(w, "data: {\"status\": \"alive\"}\n\n")
			flusher.Flush()
		}
	}
}
