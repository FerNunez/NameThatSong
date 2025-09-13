package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/repository/database"
	"github.com/google/uuid"
)

type PlaylistStore interface {
	CreatePlaylist(ctx context.Context, playlist *models.LocalPlaylist) error
	GetPlaylistByID(ctx context.Context, id uuid.UUID) (*models.LocalPlaylist, error)
	GetPlaylistByUserIDAndID(ctx context.Context, id, userID uuid.UUID) (*models.LocalPlaylist, error)
	GetPlaylistsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.LocalPlaylist, error)
	UpdatePlaylist(ctx context.Context, playlist *models.LocalPlaylist) error
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
	db *database.Queries
}

func NewSQLPlaylistStore(db *database.Queries) PlaylistStore {
	return &SQLPlaylistStore{
		db,
	}
}

// Playlist operations
func (s *SQLPlaylistStore) CreatePlaylist(ctx context.Context, playlist *models.LocalPlaylist) error {
	dbPlaylist, err := s.db.CreatePlaylist(ctx, database.CreatePlaylistParams{
		ID:                playlist.ID,
		UserID:            playlist.UserID,
		Name:              playlist.Name,
		Description:       nullStringFromStringPtr(&playlist.Description),
		SpotifyPlaylistID: nullStringFromStringPtr(playlist.SpotifyPlaylistID),
		SnapshotID:        nullStringFromStringPtr(playlist.SnapshotID),
		IsPublic:          playlist.IsPublic,
	})
	if err != nil {
		return err
	}

	// Update the playlist with the returned values
	playlist.CreatedAt = dbPlaylist.CreatedAt
	playlist.UpdatedAt = dbPlaylist.UpdatedAt
	return nil
}

func (s *SQLPlaylistStore) GetPlaylistByID(ctx context.Context, id uuid.UUID) (*models.LocalPlaylist, error) {
	dbPlaylist, err := s.db.GetPlaylistByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertDbLocalPlaylistToModel(dbPlaylist), nil
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
		//TODO: FILL tracks
	}

	return playlists, nil
}

func (s *SQLPlaylistStore) UpdatePlaylist(ctx context.Context, playlist *models.LocalPlaylist) error {
	return s.db.UpdatePlaylist(ctx, database.UpdatePlaylistParams{
		ID:          playlist.ID,
		Name:        playlist.Name,
		Description: nullStringFromStringPtr(&playlist.Description),
		IsPublic:    playlist.IsPublic,
		UserID:      playlist.UserID,
	})
}

func (s *SQLPlaylistStore) DeletePlaylist(ctx context.Context, id, userID uuid.UUID) error {
	return s.db.DeletePlaylist(ctx, database.DeletePlaylistParams{
		ID:     id,
		UserID: userID,
	})
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
