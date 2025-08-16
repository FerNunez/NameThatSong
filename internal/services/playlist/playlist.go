package playlist

import (
	"context"
	"fmt"
	"time"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
	"github.com/FerNunez/NameThatSong/internal/pkg/validation"
	"github.com/FerNunez/NameThatSong/internal/repository"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
	"github.com/google/uuid"
)

type Playlist struct {
	playlistStore  repository.PlaylistStore
	spotifyService spotify.SpotifyService
}

func NewPlaylistService(
	playlistStore repository.PlaylistStore,
	spotifyService spotify.SpotifyService,
) PlaylistService {
	return &Playlist{
		playlistStore:  playlistStore,
		spotifyService: spotifyService,
	}
}

// Local playlist operations
func (p *Playlist) CreatePlaylist(ctx context.Context, userID uuid.UUID, req models.CreatePlaylistRequest) (*models.Playlist, error) {
	logger.Info(ctx, "creating new playlist",
		logger.F("user_id", userID),
		logger.F("name", req.Name),
		logger.F("sync_with_spotify", req.SyncWithSpotify))
	
	// Validate request
	if err := p.validateCreatePlaylistRequest(req); err != nil {
		logger.Warn(ctx, "playlist creation validation failed",
			logger.F("user_id", userID),
			logger.F("name", req.Name),
			logger.F("error", err))
		return nil, fmt.Errorf("validation error: %w", err)
	}

	playlist := &models.Playlist{
		ID:              uuid.New(),
		UserID:          userID,
		Name:            req.Name,
		Description:     req.Description,
		IsPublic:        req.IsPublic,
		SyncWithSpotify: req.SyncWithSpotify,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := p.playlistStore.CreatePlaylist(ctx, playlist); err != nil {
		logger.Error(ctx, "failed to store playlist in database",
			logger.F("user_id", userID),
			logger.F("playlist_id", playlist.ID),
			logger.F("name", req.Name),
			logger.F("error", err))
		return nil, fmt.Errorf("failed to create playlist: %w", err)
	}

	logger.Info(ctx, "playlist created successfully",
		logger.F("user_id", userID),
		logger.F("playlist_id", playlist.ID),
		logger.F("name", req.Name))
	return playlist, nil
}

func (p *Playlist) GetUserPlaylists(ctx context.Context, userID uuid.UUID) ([]*models.Playlist, error) {
	return p.playlistStore.GetPlaylistsByUserID(ctx, userID)
}

func (p *Playlist) GetPlaylist(ctx context.Context, playlistID, userID uuid.UUID) (*models.Playlist, error) {
	return p.playlistStore.GetPlaylistByUserIDAndID(ctx, playlistID, userID)
}

func (p *Playlist) UpdatePlaylist(ctx context.Context, playlistID, userID uuid.UUID, req models.UpdatePlaylistRequest) (*models.Playlist, error) {
	// Validate request
	if err := p.validateUpdatePlaylistRequest(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Get existing playlist to ensure it exists and user owns it
	playlist, err := p.playlistStore.GetPlaylistByUserIDAndID(ctx, playlistID, userID)
	if err != nil {
		return nil, fmt.Errorf("playlist not found: %w", err)
	}

	// Update playlist fields
	playlist.Name = req.Name
	playlist.Description = req.Description
	playlist.IsPublic = req.IsPublic
	playlist.SyncWithSpotify = req.SyncWithSpotify
	playlist.UpdatedAt = time.Now()

	if err := p.playlistStore.UpdatePlaylist(ctx, playlist); err != nil {
		return nil, fmt.Errorf("failed to update playlist: %w", err)
	}

	return playlist, nil
}

func (p *Playlist) DeletePlaylist(ctx context.Context, playlistID, userID uuid.UUID) error {
	return p.playlistStore.DeletePlaylist(ctx, playlistID, userID)
}

// Song management
func (p *Playlist) AddSongToPlaylist(ctx context.Context, playlistID, userID uuid.UUID, req models.AddSongRequest) error {
	// Validate request
	if err := p.validateAddSongRequest(req); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Verify playlist exists and user owns it
	_, err := p.playlistStore.GetPlaylistByUserIDAndID(ctx, playlistID, userID)
	if err != nil {
		return fmt.Errorf("playlist not found: %w", err)
	}

	// Get track data from Spotify
	trackData, err := p.spotifyService.FetchTrack(ctx, userID.String(), req.SpotifyTrackID)
	if err != nil {
		return fmt.Errorf("failed to fetch track data: %w", err)
	}

	// Get next position
	maxPos, err := p.playlistStore.GetMaxSongPosition(ctx, playlistID)
	if err != nil {
		return fmt.Errorf("failed to get max position: %w", err)
	}

	// Create playlist song with real artist and album data
	song := &models.PlaylistSong{
		ID:             uuid.New(),
		PlaylistID:     playlistID,
		SpotifyTrackID: req.SpotifyTrackID,
		Position:       maxPos + 1,
		TrackName:      trackData.Name,
		ArtistName:     trackData.GetPrimaryArtistName(), // Real artist name from normalized data
		AlbumName:      trackData.GetAlbumName(),         // Real album name from normalized data
		DurationMs:     trackData.DurationMs,
		AddedAt:        time.Now(),
	}

	return p.playlistStore.AddSongToPlaylist(ctx, song)
}

func (p *Playlist) RemoveSongFromPlaylist(ctx context.Context, playlistID, userID uuid.UUID, songID uuid.UUID) error {
	// Verify playlist exists and user owns it
	_, err := p.playlistStore.GetPlaylistByUserIDAndID(ctx, playlistID, userID)
	if err != nil {
		return fmt.Errorf("playlist not found: %w", err)
	}

	return p.playlistStore.RemoveSongFromPlaylist(ctx, songID, playlistID)
}

func (p *Playlist) ReorderPlaylistSongs(ctx context.Context, playlistID, userID uuid.UUID, req models.ReorderSongsRequest) error {
	// Validate request
	if err := p.validateReorderSongsRequest(req); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Verify playlist exists and user owns it
	_, err := p.playlistStore.GetPlaylistByUserIDAndID(ctx, playlistID, userID)
	if err != nil {
		return fmt.Errorf("playlist not found: %w", err)
	}

	// Update positions for each song
	for i, songID := range req.SongOrder {
		if err := p.playlistStore.UpdateSongPosition(ctx, songID, i+1); err != nil {
			return fmt.Errorf("failed to update song position: %w", err)
		}
	}

	return nil
}

func (p *Playlist) GetPlaylistWithSongs(ctx context.Context, playlistID, userID uuid.UUID) (*models.Playlist, error) {
	// Get playlist
	playlist, err := p.playlistStore.GetPlaylistByUserIDAndID(ctx, playlistID, userID)
	if err != nil {
		return nil, fmt.Errorf("playlist not found: %w", err)
	}

	// Get songs
	songs, err := p.playlistStore.GetPlaylistSongs(ctx, playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist songs: %w", err)
	}

	playlist.Songs = make([]models.PlaylistSong, len(songs))
	for i, song := range songs {
		playlist.Songs[i] = *song
	}

	return playlist, nil
}

// Spotify integration
func (p *Playlist) ImportFromSpotify(ctx context.Context, userID uuid.UUID, req models.ImportPlaylistRequest) (*models.Playlist, error) {
	logger.Info(ctx, "importing playlist from Spotify",
		logger.F("user_id", userID),
		logger.F("spotify_playlist_id", req.SpotifyPlaylistID),
		logger.F("sync_with_spotify", req.SyncWithSpotify))
	
	// Validate request
	if err := p.validateImportPlaylistRequest(req); err != nil {
		logger.Warn(ctx, "playlist import validation failed",
			logger.F("user_id", userID),
			logger.F("spotify_playlist_id", req.SpotifyPlaylistID),
			logger.F("error", err))
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Fetch playlist data from Spotify
	spotifyPlaylist, err := p.spotifyService.FetchPlaylist(ctx, userID.String(), req.SpotifyPlaylistID)
	if err != nil {
		logger.Error(ctx, "failed to fetch playlist from Spotify API",
			logger.F("user_id", userID),
			logger.F("spotify_playlist_id", req.SpotifyPlaylistID),
			logger.F("error", err))
		return nil, fmt.Errorf("failed to fetch playlist from Spotify: %w", err)
	}

	// Create local playlist
	playlist := &models.Playlist{
		ID:                uuid.New(),
		UserID:            userID,
		Name:              spotifyPlaylist.Name,
		Description:       spotifyPlaylist.Description,
		SpotifyPlaylistID: &req.SpotifyPlaylistID,
		IsPublic:          spotifyPlaylist.Public,
		SyncWithSpotify:   req.SyncWithSpotify,
		LastSyncedAt:      &time.Time{},
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := p.playlistStore.CreatePlaylist(ctx, playlist); err != nil {
		return nil, fmt.Errorf("failed to create playlist: %w", err)
	}

	logger.Debug(ctx, "fetched playlist metadata from Spotify",
		logger.F("playlist_name", spotifyPlaylist.Name),
		logger.F("spotify_playlist_id", req.SpotifyPlaylistID))
	
	// Fetch and import tracks
	trackIDs, err := p.spotifyService.FetchTracksFromPlaylist(ctx, userID.String(), req.SpotifyPlaylistID)
	if err != nil {
		logger.Error(ctx, "failed to fetch tracks from Spotify playlist",
			logger.F("user_id", userID),
			logger.F("spotify_playlist_id", req.SpotifyPlaylistID),
			logger.F("error", err))
		return nil, fmt.Errorf("failed to fetch tracks from playlist: %w", err)
	}
	
	logger.Info(ctx, "fetched tracks from Spotify playlist",
		logger.F("spotify_playlist_id", req.SpotifyPlaylistID),
		logger.F("track_count", len(trackIDs)))

	// Add songs to playlist (batch process for better performance)
	songs := make([]*models.PlaylistSong, 0, len(trackIDs))
	for i, trackID := range trackIDs {
		trackData, err := p.spotifyService.FetchTrack(ctx, userID.String(), trackID)
		if err != nil {
			continue // Skip tracks that can't be fetched
		}

		song := &models.PlaylistSong{
			ID:             uuid.New(),
			PlaylistID:     playlist.ID,
			SpotifyTrackID: trackID,
			Position:       i + 1,
			TrackName:      trackData.Name,
			ArtistName:     trackData.GetPrimaryArtistName(), // Real artist name from normalized data
			AlbumName:      trackData.GetAlbumName(),         // Real album name from normalized data
			DurationMs:     trackData.DurationMs,
			AddedAt:        time.Now(),
		}
		songs = append(songs, song)
	}

	// Batch insert all songs
	for _, song := range songs {
		if err := p.playlistStore.AddSongToPlaylist(ctx, song); err != nil {
			// Continue with other songs if one fails
			continue
		}
	}

	// Update sync time
	now := time.Now()
	playlist.LastSyncedAt = &now
	if err := p.playlistStore.UpdatePlaylistSyncTime(ctx, playlist.ID, userID); err != nil {
		// Non-fatal error
	}

	return playlist, nil
}

func (p *Playlist) ExportToSpotify(ctx context.Context, userID uuid.UUID, req models.ExportPlaylistRequest) (string, error) {
	// Validate request
	if err := p.validateExportPlaylistRequest(req); err != nil {
		return "", fmt.Errorf("validation error: %w", err)
	}

	// Get playlist and verify ownership
	playlist, err := p.playlistStore.GetPlaylistByUserIDAndID(ctx, req.PlaylistID, userID)
	if err != nil {
		return "", fmt.Errorf("playlist not found: %w", err)
	}

	// Create playlist on Spotify
	spotifyPlaylist, err := p.spotifyService.CreatePlaylist(ctx, userID.String(), req.Name, playlist.Description, req.IsPublic)
	if err != nil {
		return "", fmt.Errorf("failed to create Spotify playlist: %w", err)
	}

	// Get playlist songs
	songs, err := p.playlistStore.GetPlaylistSongs(ctx, req.PlaylistID)
	if err != nil {
		return "", fmt.Errorf("failed to get playlist songs: %w", err)
	}

	// Extract track IDs
	trackIDs := make([]string, len(songs))
	for i, song := range songs {
		trackIDs[i] = song.SpotifyTrackID
	}

	// Add tracks to Spotify playlist
	if len(trackIDs) > 0 {
		if err := p.spotifyService.AddTracksToPlaylist(ctx, userID.String(), spotifyPlaylist.ID, trackIDs); err != nil {
			return "", fmt.Errorf("failed to add tracks to Spotify playlist: %w", err)
		}
	}

	return spotifyPlaylist.ID, nil
}

func (p *Playlist) SyncWithSpotify(ctx context.Context, playlistID, userID uuid.UUID) error {
	// Get playlist and verify it has Spotify ID
	playlist, err := p.playlistStore.GetPlaylistByUserIDAndID(ctx, playlistID, userID)
	if err != nil {
		return fmt.Errorf("playlist not found: %w", err)
	}

	if playlist.SpotifyPlaylistID == nil {
		return fmt.Errorf("playlist is not linked to Spotify")
	}

	// Fetch current tracks from Spotify
	trackIDs, err := p.spotifyService.FetchTracksFromPlaylist(ctx, userID.String(), *playlist.SpotifyPlaylistID)
	if err != nil {
		return fmt.Errorf("failed to fetch tracks from Spotify: %w", err)
	}

	// Clear existing songs
	if err := p.playlistStore.ClearPlaylistSongs(ctx, playlistID); err != nil {
		return fmt.Errorf("failed to clear playlist songs: %w", err)
	}

	// Add updated songs (batch process for better performance)
	songs := make([]*models.PlaylistSong, 0, len(trackIDs))
	for i, trackID := range trackIDs {
		trackData, err := p.spotifyService.FetchTrack(ctx, userID.String(), trackID)
		if err != nil {
			continue // Skip tracks that can't be fetched
		}

		song := &models.PlaylistSong{
			ID:             uuid.New(),
			PlaylistID:     playlistID,
			SpotifyTrackID: trackID,
			Position:       i + 1,
			TrackName:      trackData.Name,
			ArtistName:     trackData.GetPrimaryArtistName(), // Real artist name from normalized data
			AlbumName:      trackData.GetAlbumName(),         // Real album name from normalized data
			DurationMs:     trackData.DurationMs,
			AddedAt:        time.Now(),
		}
		songs = append(songs, song)
	}

	// Batch insert all songs
	for _, song := range songs {
		if err := p.playlistStore.AddSongToPlaylist(ctx, song); err != nil {
			continue
		}
	}

	// Update sync time
	return p.playlistStore.UpdatePlaylistSyncTime(ctx, playlistID, userID)
}

// Validation helpers
func (p *Playlist) validateCreatePlaylistRequest(req models.CreatePlaylistRequest) error {
	// Validate playlist name
	if err := validation.ValidatePlaylistName(req.Name); err != nil {
		return fmt.Errorf("invalid name: %w", err)
	}

	// Validate playlist description (optional)
	if err := validation.ValidatePlaylistDescription(req.Description); err != nil {
		return fmt.Errorf("invalid description: %w", err)
	}

	// Boolean fields (IsPublic, SyncWithSpotify) are automatically validated by type system
	return nil
}

func (p *Playlist) validateUpdatePlaylistRequest(req models.UpdatePlaylistRequest) error {
	// Same validation rules as create
	if err := validation.ValidatePlaylistName(req.Name); err != nil {
		return fmt.Errorf("invalid name: %w", err)
	}

	if err := validation.ValidatePlaylistDescription(req.Description); err != nil {
		return fmt.Errorf("invalid description: %w", err)
	}

	return nil
}

func (p *Playlist) validateImportPlaylistRequest(req models.ImportPlaylistRequest) error {
	// Validate Spotify playlist ID
	if err := validation.ValidateSpotifyID(req.SpotifyPlaylistID); err != nil {
		return fmt.Errorf("invalid spotify playlist ID: %w", err)
	}

	return nil
}

func (p *Playlist) validateExportPlaylistRequest(req models.ExportPlaylistRequest) error {
	// Validate playlist ID is not nil/empty
	if req.PlaylistID == uuid.Nil {
		return fmt.Errorf("playlist ID is required")
	}

	// Validate export name
	if err := validation.ValidatePlaylistName(req.Name); err != nil {
		return fmt.Errorf("invalid export name: %w", err)
	}

	return nil
}

func (p *Playlist) validateAddSongRequest(req models.AddSongRequest) error {
	// Validate Spotify track ID
	if err := validation.ValidateSpotifyID(req.SpotifyTrackID); err != nil {
		return fmt.Errorf("invalid spotify track ID: %w", err)
	}

	return nil
}

func (p *Playlist) validateReorderSongsRequest(req models.ReorderSongsRequest) error {
	// Check that song order is not empty
	if len(req.SongOrder) == 0 {
		return fmt.Errorf("song order cannot be empty")
	}

	// Check that no UUIDs are nil
	for i, songID := range req.SongOrder {
		if songID == uuid.Nil {
			return fmt.Errorf("song ID at position %d is invalid", i)
		}
	}

	// Check for duplicates
	seen := make(map[uuid.UUID]bool)
	for i, songID := range req.SongOrder {
		if seen[songID] {
			return fmt.Errorf("duplicate song ID found at position %d", i)
		}
		seen[songID] = true
	}

	return nil
}

