package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
	"github.com/FerNunez/NameThatSong/internal/repository/database"
	"github.com/google/uuid"
)

type PlaylistStore interface {
	UpsertPlaylistWithTracks(ctx context.Context, playlist *models.LocalPlaylist) error
	GetPlaylistByID(ctx context.Context, id uuid.UUID) (*models.LocalPlaylist, error)
	GetPlaylistByIDWithTracks(ctx context.Context, id, userID uuid.UUID) (*models.LocalPlaylist, error)
	GetPlaylistByUserIDAndID(ctx context.Context, id, userID uuid.UUID) (*models.LocalPlaylist, error)
	GetPlaylistsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.LocalPlaylist, error)
	GetPlaylistsByUserIDWithTracks(ctx context.Context, userID uuid.UUID) ([]models.LocalPlaylist, error)
	DeletePlaylist(ctx context.Context, id, userID uuid.UUID) error

	AddSongToPlaylist(ctx context.Context, playlistID uuid.UUID, trackID string, position int) error
	GetPlaylistSongs(ctx context.Context, playlistID uuid.UUID) ([]string, []int, error)
	GetPlaylistSongsWithTrackData(ctx context.Context, playlistID uuid.UUID) ([]database.GetPlaylistSongsWithTrackDataRow, error)
	RemoveSongFromPlaylist(ctx context.Context, playlistID uuid.UUID, songID string) error
	//UpdateSongPosition(ctx context.Context, playlistID uuid.UUID, songID string, position int) error
	ClearPlaylistSongs(ctx context.Context, playlistID uuid.UUID) error
	GetPlaylistBySpotifyIDAndUserID(ctx context.Context, playlistID models.SpotifyID, userID uuid.UUID) (*models.LocalPlaylist, error)
}

type SQLPlaylistStore struct {
	db   *database.Queries
	conn database.DBTX
}

func NewSQLPlaylistStore(queries *database.Queries, conn database.DBTX) PlaylistStore {
	return &SQLPlaylistStore{
		db:   queries,
		conn: conn,
	}
}

// Playlist operations

func (s *SQLPlaylistStore) GetPlaylistByID(ctx context.Context, id uuid.UUID) (*models.LocalPlaylist, error) {
	dbPlaylist, err := s.db.GetPlaylistByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertDbLocalPlaylistToModel(dbPlaylist), nil
}

func (s *SQLPlaylistStore) GetPlaylistByIDWithTracks(ctx context.Context, id, userID uuid.UUID) (*models.LocalPlaylist, error) {
	rows, err := s.db.GetPlaylistByIDWithTracks(ctx, database.GetPlaylistByIDWithTracksParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}

	// First row contains playlist metadata
	firstRow := rows[0]
	playlist := &models.LocalPlaylist{
		ID:                firstRow.PlaylistID,
		UserID:            firstRow.UserID,
		Name:              firstRow.PlaylistName,
		Description:       stringFromNullString(firstRow.Description),
		ImageURL:          stringPtrFromNullString(firstRow.ImageUrl),
		SpotifyPlaylistID: stringPtrFromNullString(firstRow.SpotifyPlaylistID),
		SnapshotID:        stringPtrFromNullString(firstRow.SnapshotID),
		IsPublic:          firstRow.IsPublic,
		LastSyncedAt:      timePtrFromNullTime(firstRow.LastSyncedAt),
		CreatedAt:         firstRow.CreatedAt,
		UpdatedAt:         firstRow.UpdatedAt,
		Tracks:            []models.TrackData{},
	}

	// Process tracks from all rows
	for _, row := range rows {
		if row.SpotifyTrackID.Valid {
			track := models.TrackData{
				ID:          models.SpotifyID(row.SpotifyTrackID.String),
				Name:        stringFromNullString(row.TrackName),
				DurationMs:  int(row.DurationMs.Int32),
				DiscNumber:  int(row.DiscNumber.Int32),
				TrackNumber: int(row.TrackNumber.Int32),
				Popularity:  int(row.Popularity.Int32),
				Explicit:    row.Explicit.Bool,
				IsLocal:     row.IsLocal.Bool,
				AlbumID:     models.SpotifyID(stringFromNullString(row.AlbumID)),
				ArtistIDs:   convertStringArrayToSpotifyIDs(row.ArtistIds),
				CachedAt:    row.CachedAt.Time,
			}
			playlist.Tracks = append(playlist.Tracks, track)
		}
	}

	return playlist, nil
}

func (s *SQLPlaylistStore) GetPlaylistByUserIDAndID(ctx context.Context, id, userID uuid.UUID) (*models.LocalPlaylist, error) {
	dbPlaylist, err := s.db.GetPlaylistByUserIDAndID(ctx, database.GetPlaylistByUserIDAndIDParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return convertDbLocalPlaylistToModel(dbPlaylist), nil
}

func (s *SQLPlaylistStore) GetPlaylistsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.LocalPlaylist, error) {
	dbPlaylists, err := s.db.GetPlaylistsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// convert into LocalPlaylist
	playlists := make([]*models.LocalPlaylist, len(dbPlaylists))
	for i, dbPlaylist := range dbPlaylists {
		playlists[i] = convertDbLocalPlaylistToModel(dbPlaylist)
		logger.Debug(ctx, "fetched a playlistByUserID", logger.F("", dbPlaylists))
		//TODO: FILL tracks
	}

	return playlists, nil
}

func (s *SQLPlaylistStore) DeletePlaylist(ctx context.Context, id, userID uuid.UUID) error {
	return s.db.DeletePlaylist(ctx, database.DeletePlaylistParams{
		ID:     id,
		UserID: userID,
	})
}

func (s *SQLPlaylistStore) UpsertPlaylistWithTracks(ctx context.Context, playlist *models.LocalPlaylist) error {
	// Validate playlist and track data
	if err := s.validatePlaylistWithTracks(playlist); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Start transaction
	tx, err := s.conn.(*sql.DB).BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := s.db.WithTx(tx)

	// 1. Upsert playlist metadata
	dbPlaylist, err := qtx.UpsertPlaylist(ctx, database.UpsertPlaylistParams{
		ID:                playlist.ID,
		UserID:            playlist.UserID,
		Name:              playlist.Name,
		Description:       nullStringFromStringPtr(&playlist.Description),
		ImageUrl:          nullStringFromStringPtr(playlist.ImageURL),
		SpotifyPlaylistID: nullStringFromStringPtr(playlist.SpotifyPlaylistID),
		SnapshotID:        nullStringFromStringPtr(playlist.SnapshotID),
		IsPublic:          playlist.IsPublic,
	})
	if err != nil {
		return err
	}

	// 2. Clear existing tracks
	err = qtx.ClearPlaylistSongs(ctx, playlist.ID)
	if err != nil {
		return err
	}

	// 3. Insert tracks if any exist
	if len(playlist.Tracks) > 0 {
		trackIDs := make([]string, len(playlist.Tracks))
		positions := make([]int32, len(playlist.Tracks))

		for i, track := range playlist.Tracks {
			trackIDs[i] = string(track.ID)
			positions[i] = int32(i)
		}

		err = qtx.BulkInsertPlaylistTracks(ctx, database.BulkInsertPlaylistTracksParams{
			PlaylistID: playlist.ID,
			Column2:    trackIDs,
			Column3:    positions,
		})
		if err != nil {
			return err
		}
	}

	// Update playlist with returned values
	playlist.CreatedAt = dbPlaylist.CreatedAt
	playlist.UpdatedAt = dbPlaylist.UpdatedAt

	return tx.Commit()
}

// Playlist song operations
func (s *SQLPlaylistStore) AddSongToPlaylist(ctx context.Context, playlistID uuid.UUID, track_id string, position int) error {
	_, err := s.db.AddSongToPlaylist(ctx, database.AddSongToPlaylistParams{
		PlaylistID:     playlistID,
		SpotifyTrackID: track_id,
		Position:       int32(position),
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *SQLPlaylistStore) GetPlaylistSongs(ctx context.Context, playlistID uuid.UUID) ([]string, []int, error) {
	dbSongs, err := s.db.GetPlaylistSongs(ctx, playlistID)
	if err != nil {
		return nil, nil, err
	}

	songIDs := make([]string, len(dbSongs))
	positions := make([]int, len(dbSongs))
	for i, dbSong := range dbSongs {
		songIDs[i] = dbSong.SpotifyTrackID
		positions[i] = int(dbSong.Position)
	}
	return songIDs, positions, nil
}

func (s *SQLPlaylistStore) GetPlaylistSongsWithTrackData(ctx context.Context, playlistID uuid.UUID) ([]database.GetPlaylistSongsWithTrackDataRow, error) {
	return s.db.GetPlaylistSongsWithTrackData(ctx, playlistID)
}

func (s *SQLPlaylistStore) RemoveSongFromPlaylist(ctx context.Context, playlistID uuid.UUID, spotify_track_id string) error {
	return s.db.RemoveSongFromPlaylist(ctx, database.RemoveSongFromPlaylistParams{
		PlaylistID:     playlistID,
		SpotifyTrackID: spotify_track_id,
	})
}

func (s *SQLPlaylistStore) UpdateSongPosition(ctx context.Context, playlistID uuid.UUID, spotify_track_id string, position int) error {
	return s.db.UpdateSongPosition(ctx, database.UpdateSongPositionParams{
		Position:       int32(position),
		SpotifyTrackID: spotify_track_id,
	})
}

func (s *SQLPlaylistStore) GetMaxSongPosition(ctx context.Context, playlistID uuid.UUID) (int, error) {
	maxPos, err := s.db.GetMaxSongPosition(ctx, playlistID)
	if err != nil {
		return 0, err
	}
	return int(maxPos), nil
}

func (s *SQLPlaylistStore) ClearPlaylistSongs(ctx context.Context, playlistID uuid.UUID) error {
	return s.db.ClearPlaylistSongs(ctx, playlistID)
}

func (s *SQLPlaylistStore) GetPlaylistBySpotifyIDAndUserID(ctx context.Context, playlistSpotifyID models.SpotifyID, userID uuid.UUID) (*models.LocalPlaylist, error) {
	localPlaylisDb, err := s.db.GetPlaylistBySpotifyIDAndUserID(ctx, database.GetPlaylistBySpotifyIDAndUserIDParams{
		SpotifyPlaylistID: sql.NullString{String: string(playlistSpotifyID), Valid: true},
		UserID:            userID,
	})
	if err != nil {
		return nil, err
	}
	return convertDbLocalPlaylistToModel(localPlaylisDb), nil
}

// Helper functions for conversion
func convertDbLocalPlaylistToModel(dbPlaylist database.LocalPlaylist) *models.LocalPlaylist {
	return &models.LocalPlaylist{
		ID:                dbPlaylist.ID,
		UserID:            dbPlaylist.UserID,
		Name:              dbPlaylist.Name,
		Description:       stringFromNullString(dbPlaylist.Description),
		SpotifyPlaylistID: stringPtrFromNullString(dbPlaylist.SpotifyPlaylistID),
		IsPublic:          dbPlaylist.IsPublic,
		SnapshotID:        stringPtrFromNullString(dbPlaylist.SnapshotID),
		LastSyncedAt:      timePtrFromNullTime(dbPlaylist.LastSyncedAt),
		CreatedAt:         dbPlaylist.CreatedAt,
		UpdatedAt:         dbPlaylist.UpdatedAt,
		ImageURL:          stringPtrFromNullString(dbPlaylist.ImageUrl),
		Tracks:            []models.TrackData{},
	}
}

// Helper functions for handling nullable types
func nullStringFromStringPtr(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func stringPtrFromNullString(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func stringFromNullString(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

func timePtrFromNullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

// GetPlaylistsByUserIDWithTracks returns complete LocalPlaylist objects with TrackData included
func (s *SQLPlaylistStore) GetPlaylistsByUserIDWithTracks(ctx context.Context, userID uuid.UUID) ([]models.LocalPlaylist, error) {
	rows, err := s.db.GetPlaylistsByUserIDWithTracks(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Group tracks by playlist
	playlistMap := make(map[uuid.UUID]*models.LocalPlaylist)

	for _, row := range rows {
		// Create playlist if not exists
		if _, exists := playlistMap[row.PlaylistID]; !exists {
			playlistMap[row.PlaylistID] = &models.LocalPlaylist{
				ID:                row.PlaylistID,
				UserID:            row.UserID,
				Name:              row.PlaylistName,
				Description:       stringFromNullString(row.Description),
				ImageURL:          stringPtrFromNullString(row.ImageUrl),
				SpotifyPlaylistID: stringPtrFromNullString(row.SpotifyPlaylistID),
				SnapshotID:        stringPtrFromNullString(row.SnapshotID),
				IsPublic:          row.IsPublic,
				LastSyncedAt:      timePtrFromNullTime(row.LastSyncedAt),
				CreatedAt:         row.CreatedAt,
				UpdatedAt:         row.UpdatedAt,
				Tracks:            []models.TrackData{},
			}
		}

		// Add track if exists (LEFT JOIN can return null tracks)
		if row.SpotifyTrackID.Valid {
			track := models.TrackData{
				ID:          models.SpotifyID(row.SpotifyTrackID.String),
				Name:        stringFromNullString(row.TrackName),
				DurationMs:  int(row.DurationMs.Int32),
				DiscNumber:  int(row.DiscNumber.Int32),
				TrackNumber: int(row.TrackNumber.Int32),
				Popularity:  int(row.Popularity.Int32),
				Explicit:    row.Explicit.Bool,
				IsLocal:     row.IsLocal.Bool,
				AlbumID:     models.SpotifyID(stringFromNullString(row.AlbumID)),
				ArtistIDs:   convertStringArrayToSpotifyIDs(row.ArtistIds),
				CachedAt:    row.CachedAt.Time,
			}
			playlistMap[row.PlaylistID].Tracks = append(playlistMap[row.PlaylistID].Tracks, track)
		}
	}

	// Convert map to slice and maintain order (created_at DESC from SQL query)
	result := make([]models.LocalPlaylist, 0, len(playlistMap))

	// We need to maintain the order from the original query
	seenPlaylists := make(map[uuid.UUID]bool)
	for _, row := range rows {
		if !seenPlaylists[row.PlaylistID] {
			if playlist, exists := playlistMap[row.PlaylistID]; exists {
				result = append(result, *playlist)
				seenPlaylists[row.PlaylistID] = true
			}
		}
	}

	return result, nil
}

// Helper function to convert string array to SpotifyID array
func convertStringArrayToSpotifyIDs(strArray []string) []models.SpotifyID {
	spotifyIDs := make([]models.SpotifyID, len(strArray))
	for i, str := range strArray {
		spotifyIDs[i] = models.SpotifyID(str)
	}
	return spotifyIDs
}

// validatePlaylistWithTracks validates playlist data and track relationships
func (s *SQLPlaylistStore) validatePlaylistWithTracks(playlist *models.LocalPlaylist) error {
	// Basic playlist validation
	if playlist == nil {
		return fmt.Errorf("playlist cannot be nil")
	}
	if playlist.Name == "" {
		return fmt.Errorf("playlist name cannot be empty")
	}
	if playlist.ID == uuid.Nil {
		return fmt.Errorf("playlist ID cannot be nil")
	}
	if playlist.UserID == uuid.Nil {
		return fmt.Errorf("user ID cannot be nil")
	}

	// Skip track validation if no tracks
	if len(playlist.Tracks) == 0 {
		return nil
	}

	// Track validation
	trackIDsSeen := make(map[string]bool)

	for i, track := range playlist.Tracks {
		// Check for empty track ID
		if string(track.ID) == "" {
			return fmt.Errorf("track %d has empty ID", i)
		}

		// Check for duplicate track IDs within playlist
		trackIDStr := string(track.ID)
		if trackIDsSeen[trackIDStr] {
			return fmt.Errorf("duplicate track ID %s at position %d", trackIDStr, i)
		}
		trackIDsSeen[trackIDStr] = true
	}

	// Validate that track IDs exist in spotify_tracks table
	if len(playlist.Tracks) > 0 {
		trackIDs := make([]string, len(playlist.Tracks))
		for i, track := range playlist.Tracks {
			trackIDs[i] = string(track.ID)
		}

		// Check if tracks exist in database (optional - depends on if tracks should be pre-cached)
		// This can be expensive for large playlists, so we might want to make it optional
		// For now, we'll assume tracks are valid if they have proper IDs
	}

	return nil
}
