package playlist

import (
	"context"
	"fmt"
	"time"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
	"github.com/FerNunez/NameThatSong/internal/pkg/validation"
	"github.com/FerNunez/NameThatSong/internal/repository"
	"github.com/FerNunez/NameThatSong/internal/services/events"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
	"github.com/google/uuid"
)

type PlaylistProvider struct {
	playlistStore  repository.PlaylistStore
	spotifyService spotify.SpotifyService
	eventBus       *events.EventBus
}

func NewPlaylistService(
	playlistStore repository.PlaylistStore,
	spotifyService spotify.SpotifyService,
) Service {
	return &PlaylistProvider{
		playlistStore:  playlistStore,
		spotifyService: spotifyService,
		eventBus:       events.GetGlobalEventBus(),
	}
}

// Local playlist operations
func (p *PlaylistProvider) CreatePlaylist(ctx context.Context, userID uuid.UUID, req models.CreatePlaylistRequest) (*models.LocalPlaylist, error) {
	logger.Info(ctx, "creating new playlist",
		logger.F("user_id", userID),
		logger.F("name", req.Name),
		logger.F("SnapshotID", req.SnapshotID))

	// Validate request
	if err := p.validateCreatePlaylistRequest(req); err != nil {
		logger.Warn(ctx, "playlist creation validation failed",
			logger.F("user_id", userID),
			logger.F("name", req.Name),
			logger.F("error", err))
		return nil, fmt.Errorf("validation error: %w", err)
	}

	playlist := &models.LocalPlaylist{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		IsPublic:    req.IsPublic,
		SnapshotID:  &req.SnapshotID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := p.playlistStore.UpsertPlaylistWithTracks(ctx, playlist); err != nil {
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

	// Emit playlist created event
	p.eventBus.Publish(events.Event{
		Type:   events.PlaylistCreated,
		UserID: userID,
		Data: map[string]any{
			"playlist_id": playlist.ID.String(),
		},
	})

	return playlist, nil
}

func (p *PlaylistProvider) GetUserPlaylists(ctx context.Context, userID uuid.UUID) ([]*models.LocalPlaylist, error) {
	localPlaylists, err := p.playlistStore.GetPlaylistsByUserIDWithTracks(ctx, userID)
	if err != nil {
		return nil, err
	}
	logger.Debug(ctx, "retreived localPlaylists for user", logger.F("user", userID), logger.F("# playlist", len(localPlaylists)))

	localPlaylistsPtr := make([]*models.LocalPlaylist, 0, len(localPlaylists))
	for _, p := range localPlaylists {
		localPlaylistsPtr = append(localPlaylistsPtr, &p)
	}

	return localPlaylistsPtr, nil
	//localPlaylistsPtr/return p.playlistStore.GetPlaylistsByUserID(ctx, userID)
}

func (p *PlaylistProvider) GetPlaylist(ctx context.Context, playlistID, userID uuid.UUID) (*models.LocalPlaylist, error) {
	return p.playlistStore.GetPlaylistByUserIDAndID(ctx, playlistID, userID)
}

func (p *PlaylistProvider) UpdatePlaylist(ctx context.Context, playlistID, userID uuid.UUID, req models.UpdatePlaylistRequest) (*models.LocalPlaylist, error) {
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
	playlist.SnapshotID = &req.SnapshotID
	playlist.UpdatedAt = time.Now()

	if err := p.playlistStore.UpsertPlaylistWithTracks(ctx, playlist); err != nil {
		return nil, fmt.Errorf("failed to update playlist: %w", err)
	}

	// Emit playlist updated event
	p.eventBus.Publish(events.Event{
		Type:   events.PlaylistUpdated,
		UserID: userID,
		Data: map[string]any{
			"playlist_id": playlist.ID.String(),
		},
	})

	return playlist, nil
}

func (p *PlaylistProvider) DeletePlaylist(ctx context.Context, playlistID, userID uuid.UUID) error {
	return p.playlistStore.DeletePlaylist(ctx, playlistID, userID)
}

// Song management
func (p *PlaylistProvider) AddSongToPlaylist(ctx context.Context, userID string, playlistID uuid.UUID, req models.AddSongRequest) error {
	// Validate request
	if err := p.validateAddSongRequest(req); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Ensure track exists in spotify_tracks table (3-tier caching will handle this)
	track, err := p.spotifyService.FetchTrack(ctx, userID, models.SpotifyID(req.SpotifyTrackID))
	if err != nil {
		return fmt.Errorf("couldn't fetch track from Spotify: %w", err)
	}

	songIDs, _, err := p.playlistStore.GetPlaylistSongs(ctx, playlistID)
	if err != nil {
		return err
	}

	err = p.playlistStore.AddSongToPlaylist(ctx, playlistID, req.SpotifyTrackID, len(songIDs))
	if err != nil {
		return err
	}

	// Convert userID string to UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		logger.Error(ctx, "failed to parse user ID", logger.F("user_id", userID), logger.F("error", err))
		return nil // Don't fail the operation, just skip the event
	}

	// Emit song added event
	p.eventBus.Publish(events.Event{
		Type:   events.PlaylistSongAdded,
		UserID: userUUID,
		Data: map[string]any{
			"playlist_id": playlistID.String(),
			"track_id":    track.Name,
			"track_name":  track.ID,
		},
	})

	return nil
}

func (p *PlaylistProvider) GetPlaylistSongs(ctx context.Context, userID string, playlistID uuid.UUID) ([]*models.PlaylistTrack, error) {
	// Use the new query that joins with spotify_tracks directly
	trackRows, err := p.playlistStore.GetPlaylistSongsWithTrackData(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	tracks := make([]*models.PlaylistTrack, len(trackRows))
	for i, row := range trackRows {
		tracks[i] = &models.PlaylistTrack{
			SpotifyTrackID: row.SpotifyTrackID,
			Position:       row.Position,
			UpdatedAt:      row.UpdatedAt,
			TrackName:      row.TrackName,
			DurationMs:     row.TrackDurationMs,
			AlbumID:        row.TrackAlbumID,
			ArtistIds:      row.TrackArtistIds,
		}
	}
	return tracks, nil
}

func (p *PlaylistProvider) GetPlaylistSongsWithDetails(ctx context.Context, userID string, playlistID uuid.UUID) ([]*models.PlaylistTrackWithDetails, error) {
	// 1. Get basic tracks
	basicTracks, err := p.GetPlaylistSongs(ctx, userID, playlistID)
	if err != nil {
		return nil, err
	}

	if len(basicTracks) == 0 {
		return []*models.PlaylistTrackWithDetails{}, nil
	}

	// 2. Collect ALL album IDs and artist IDs (duplicates OK - SpotifyService handles deduplication)
	albumIDs := make([]models.SpotifyID, len(basicTracks))
	var allArtistIDs []models.SpotifyID

	for i, track := range basicTracks {
		albumIDs[i] = models.SpotifyID(track.AlbumID)
		for _, artistID := range track.ArtistIds {
			allArtistIDs = append(allArtistIDs, models.SpotifyID(artistID))
		}
	}

	// 3. Batch fetch albums and artists via 3-tier caching (SpotifyService handles deduplication)
	albums, err := p.spotifyService.FetchMultipleAlbums(ctx, userID, albumIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch album data: %w", err)
	}

	artists, err := p.spotifyService.FetchMultipleArtists(ctx, userID, allArtistIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch artist data: %w", err)
	}

	// 4. Build enriched tracks
	return p.enrichTracks(basicTracks, albums, artists), nil
}

// enrichTracks combines basic track data with album/artist details
func (p *PlaylistProvider) enrichTracks(
	basicTracks []*models.PlaylistTrack,
	albums []models.AlbumData,
	artists []models.ArtistData,
) []*models.PlaylistTrackWithDetails {

	// Create lookup maps from batch results
	albumMap := make(map[string]models.AlbumData)
	for _, album := range albums {
		albumMap[string(album.ID)] = album
	}

	artistMap := make(map[string]models.ArtistData)
	for _, artist := range artists {
		artistMap[string(artist.ID)] = artist
	}

	// Enrich tracks
	enriched := make([]*models.PlaylistTrackWithDetails, len(basicTracks))

	for i, track := range basicTracks {
		enriched[i] = &models.PlaylistTrackWithDetails{
			PlaylistTrack: *track,
		}

		// Add album data if available
		if album, exists := albumMap[track.AlbumID]; exists {
			enriched[i].AlbumName = album.Name
			enriched[i].AlbumImageUrl = album.ImageURL
		}

		// Add artist names if available
		artistNames := make([]string, 0, len(track.ArtistIds))
		for j, artistID := range track.ArtistIds {
			if artist, exists := artistMap[artistID]; exists {
				artistNames = append(artistNames, artist.Name)
				if j == 0 { // Use first artist image as primary
					enriched[i].PrimaryArtistImageUrl = artist.ImageURL
				}
			}
		}
		enriched[i].ArtistNames = artistNames
	}

	return enriched
}

func (p *PlaylistProvider) RemoveSongFromPlaylist(ctx context.Context, playlistID, userID uuid.UUID, spotifyTrackID string) error {
	// Verify playlist exists and user owns it
	_, err := p.playlistStore.GetPlaylistByUserIDAndID(ctx, playlistID, userID)
	if err != nil {
		return fmt.Errorf("playlist not found: %w", err)
	}

	err = p.playlistStore.RemoveSongFromPlaylist(ctx, playlistID, spotifyTrackID)
	if err != nil {
		return err
	}

	// Emit song removed event
	p.eventBus.Publish(events.Event{
		Type:   events.PlaylistSongRemoved,
		UserID: userID,
		Data: map[string]any{
			"playlist_id":      playlistID.String(),
			"spotify_track_id": spotifyTrackID,
		},
	})

	return nil
}

// func (p *Playlist) ReorderPlaylistSongs(ctx context.Context, playlistID, userID uuid.UUID, req models.ReorderSongsRequest) error {
// 	// Validate request
// 	if err := p.validateReorderSongsRequest(req); err != nil {
// 		return fmt.Errorf("validation error: %w", err)
// 	}
//
// 	// Verify playlist exists and user owns it
// 	_, err := p.playlistStore.GetPlaylistByUserIDAndID(ctx, playlistID, userID)
// 	if err != nil {
// 		return fmt.Errorf("playlist not found: %w", err)
// 	}
//
// 	// Update positions for each song
// 	for i, songID := range req.SongOrder {
// 		if err := p.playlistStore.UpdateSongPosition(ctx, playlistID, req., i+1); err != nil {
// 			return fmt.Errorf("failed to update song position: %w", err)
// 		}
// 	}
//
// 	return nil
// }

// Spotify integration
func (p *PlaylistProvider) ImportFromSpotify(ctx context.Context, userID uuid.UUID, req models.ImportPlaylistRequest) (*models.LocalPlaylist, error) {
	logger.Info(ctx, "importing playlist from Spotify",
		logger.F("user_id", userID),
		logger.F("spotify_playlist_id", req.SpotifyPlaylistID),
		logger.F("snapshot_id", req.SnapshotID))

	// Validate request
	if err := p.validateImportPlaylistRequest(req); err != nil {
		logger.Warn(ctx, "playlist import validation failed",
			logger.F("user_id", userID),
			logger.F("spotify_playlist_id", req.SpotifyPlaylistID),
			logger.F("error", err))
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Fetch playlist data from Spotify
	spotifyPlaylist, err := p.spotifyService.FetchPlaylist(ctx, userID.String(), models.SpotifyID(req.SpotifyPlaylistID))
	if err != nil {
		logger.Error(ctx, "failed to fetch playlist from Spotify API",
			logger.F("user_id", userID),
			logger.F("spotify_playlist_id", req.SpotifyPlaylistID),
			logger.F("error", err))
		return nil, fmt.Errorf("failed to fetch playlist from Spotify: %w", err)
	}

	// Create local playlist
	playlist := &models.LocalPlaylist{
		ID:                uuid.New(),
		UserID:            userID,
		Name:              spotifyPlaylist.Name,
		Description:       spotifyPlaylist.Description,
		SpotifyPlaylistID: &req.SpotifyPlaylistID,
		IsPublic:          spotifyPlaylist.Public,
		SnapshotID:        &req.SnapshotID,
		LastSyncedAt:      &time.Time{},
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := p.playlistStore.UpsertPlaylistWithTracks(ctx, playlist); err != nil {
		return nil, fmt.Errorf("failed to create playlist: %w", err)
	}

	for idx, trackID := range spotifyPlaylist.TrackIDs {
		err := p.playlistStore.AddSongToPlaylist(ctx, playlist.ID, string(trackID), idx+1)
		if err != nil {
			return nil, err
		}
	}

	// Emit playlist sync completed event (since import includes all tracks)
	p.eventBus.Publish(events.Event{
		Type:   events.PlaylistSyncCompleted,
		UserID: userID,
		Data: map[string]any{
			"playlist_id":   playlist.ID.String(),
			"playlist_name": playlist.Name,
		},
	})

	return playlist, nil
}

func (p *PlaylistProvider) ExportToSpotify(ctx context.Context, userID uuid.UUID, req models.ExportPlaylistRequest) (string, error) {
	// Validate request
	if err := p.validateExportPlaylistRequest(req); err != nil {
		return "", fmt.Errorf("validation error: %w", err)
	}

	// Get playlist and verify ownership
	playlist, err := p.playlistStore.GetPlaylistByUserIDAndID(ctx, req.PlaylistID, userID)
	if err != nil {
		return "", fmt.Errorf("playlist not found: %w", err)
	}

	trackIDs, _, err := p.playlistStore.GetPlaylistSongs(ctx, playlist.ID)
	if err != nil {
		return "", err
	}

	// Create playlist on Spotify
	spotifyPlaylist, err := p.spotifyService.CreatePlaylist(ctx, userID.String(), req.Name, playlist.Description, req.IsPublic)
	if err != nil {
		return "", fmt.Errorf("failed to create Spotify playlist: %w", err)
	}

	// Add tracks to Spotify playlist
	if len(trackIDs) > 0 {
		if err := p.spotifyService.AddTracksToPlaylist(ctx, userID.String(), string(spotifyPlaylist.ID), trackIDs); err != nil {
			return "", fmt.Errorf("failed to add tracks to Spotify playlist: %w", err)
		}
	}

	return string(spotifyPlaylist.ID), nil
}

// func (p *Playlist) SyncWithSpotify(ctx context.Context, playlistID, userID uuid.UUID) error {
// 	// Get playlist and verify it has Spotify ID
// 	playlist, err := p.playlistStore.GetPlaylistByUserIDAndID(ctx, playlistID, userID)
// 	if err != nil {
// 		return fmt.Errorf("playlist not found: %w", err)
// 	}
//
// 	if playlist.SpotifyPlaylistID == nil {
// 		return fmt.Errorf("playlist is not linked to Spotify")
// 	}
//
// 	// Fetch current tracks from Spotify
// 	trackIDs, err := p.spotifyService.FetchTracksFromPlaylist(ctx, userID.String(), *playlist.SpotifyPlaylistID)
// 	if err != nil {
// 		return fmt.Errorf("failed to fetch tracks from Spotify: %w", err)
// 	}
//
// 	// Clear existing songs
// 	if err := p.playlistStore.ClearPlaylistSongs(ctx, playlistID); err != nil {
// 		return fmt.Errorf("failed to clear playlist songs: %w", err)
// 	}
//
// 	// Update sync time
// 	return p.playlistStore.UpdatePlaylistSyncTime(ctx, playlistID, userID)
// }

// Validation helpers
func (p *PlaylistProvider) validateCreatePlaylistRequest(req models.CreatePlaylistRequest) error {
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

func (p *PlaylistProvider) validateUpdatePlaylistRequest(req models.UpdatePlaylistRequest) error {
	// Same validation rules as create
	if err := validation.ValidatePlaylistName(req.Name); err != nil {
		return fmt.Errorf("invalid name: %w", err)
	}

	if err := validation.ValidatePlaylistDescription(req.Description); err != nil {
		return fmt.Errorf("invalid description: %w", err)
	}

	return nil
}

func (p *PlaylistProvider) validateImportPlaylistRequest(req models.ImportPlaylistRequest) error {
	// Validate Spotify playlist ID
	if err := validation.ValidateSpotifyID(req.SpotifyPlaylistID); err != nil {
		return fmt.Errorf("invalid spotify playlist ID: %w", err)
	}

	return nil
}

func (p *PlaylistProvider) validateExportPlaylistRequest(req models.ExportPlaylistRequest) error {
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

func (p *PlaylistProvider) validateAddSongRequest(req models.AddSongRequest) error {
	// Validate Spotify track ID
	if err := validation.ValidateSpotifyID(req.SpotifyTrackID); err != nil {
		return fmt.Errorf("invalid spotify track ID: %w", err)
	}

	return nil
}

func (p *PlaylistProvider) validateReorderSongsRequest(req models.ReorderSongsRequest) error {
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
func (p *PlaylistProvider) ImportUsersPlaylistsFromSpotify(ctx context.Context, userID uuid.UUID) ([]models.LocalPlaylist, error) {
	// fetch users playlists versions
	spotifyPlaylistsVersion, err := p.spotifyService.FetchUserSpotifyPlaylistsVersion(ctx, userID.String())
	if err != nil {
		return nil, err
	}
	logger.Debug(ctx, "fetched playlist versions", logger.F("nb playlists", len(spotifyPlaylistsVersion)))

	localPlaylists := make([]models.LocalPlaylist, 0, len(spotifyPlaylistsVersion))
	// loop playlist and check if they in local or sync
	for _, playlist := range spotifyPlaylistsVersion {

		localPlaylist, err := p.playlistStore.GetPlaylistBySpotifyIDAndUserID(ctx, playlist.ID, userID)

		if err == nil && *localPlaylist.SnapshotID == playlist.SnapshotID {
			logger.Debug(ctx, "local playlist found & same snapshot")
			continue
		} else if err == nil {
			logger.Error(ctx, "local playlist found with different snapshot. Needs update ")
		} else if err != nil {
			logger.Debug(ctx, "local playlist not found")

			// Fetch playlist data and tracks
			playlistData, err := p.spotifyService.FetchPlaylist(ctx, userID.String(), playlist.ID)
			if err != nil {
				logger.Error(ctx, "couldnt playlist")
				continue
			}
			tracks, err := p.spotifyService.FetchMultipleTracks(ctx, userID.String(), playlistData.TrackIDs)
			if err != nil {
				logger.Error(ctx, "couldnt fetch playlist song")
			}

			// Create new playlist
			timeNow := time.Now()
			spotifyPlaylistID := string(playlist.ID)
			newLocalPlaylist := &models.LocalPlaylist{
				ID:                uuid.New(),
				UserID:            userID,
				Name:              playlistData.Name,
				Description:       playlistData.Description,
				SpotifyPlaylistID: &spotifyPlaylistID,
				IsPublic:          playlistData.Public,
				SnapshotID:        &playlistData.SnapshotID,
				LastSyncedAt:      &timeNow,
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
				Tracks:            tracks,
				ImageURL:          &playlistData.ImageURL,
			}

			logger.Debug(ctx, "local playlist ot add", logger.F("localplaylist", *newLocalPlaylist))

			if err := p.playlistStore.UpsertPlaylistWithTracks(ctx, newLocalPlaylist); err != nil {
				logger.Error(ctx, "couldnt store new playlist", logger.F("err", err))
			}
			logger.Debug(ctx, "added new local playlist", logger.F("name", newLocalPlaylist.Name))
			localPlaylists = append(localPlaylists, *newLocalPlaylist)

			// publish event that song
			p.eventBus.Publish(events.Event{
				Type:   events.PlaylistImportCompleted,
				UserID: userID,
				Data: map[string]any{
					"playlist_id": newLocalPlaylist.ID.String(),
				},
			})
		}
	}

	return localPlaylists, nil
}
