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
	CreatePlaylist(ctx context.Context, playlist *models.Playlist) error
	GetPlaylistByID(ctx context.Context, id uuid.UUID) (*models.Playlist, error)
	GetPlaylistByUserIDAndID(ctx context.Context, id, userID uuid.UUID) (*models.Playlist, error)
	GetPlaylistsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Playlist, error)
	UpdatePlaylist(ctx context.Context, playlist *models.Playlist) error
	UpdatePlaylistSyncTime(ctx context.Context, id, userID uuid.UUID) error
	DeletePlaylist(ctx context.Context, id, userID uuid.UUID) error

	AddSongToPlaylist(ctx context.Context, song *models.PlaylistSong) error
	GetPlaylistSongs(ctx context.Context, playlistID uuid.UUID) ([]*models.PlaylistSong, error)
	GetPlaylistSongByID(ctx context.Context, songID, playlistID uuid.UUID) (*models.PlaylistSong, error)
	RemoveSongFromPlaylist(ctx context.Context, songID, playlistID uuid.UUID) error
	UpdateSongPosition(ctx context.Context, songID uuid.UUID, position int) error
	GetMaxSongPosition(ctx context.Context, playlistID uuid.UUID) (int, error)
	ClearPlaylistSongs(ctx context.Context, playlistID uuid.UUID) error
}

type SQLPlaylistStore struct {
	db *database.Queries
}

func NewSQLPlaylistStore(db *database.Queries) PlaylistStore {
	return &SQLPlaylistStore{
		db: db,
	}
}

// Playlist operations
func (s *SQLPlaylistStore) CreatePlaylist(ctx context.Context, playlist *models.Playlist) error {
	dbPlaylist, err := s.db.CreatePlaylist(ctx, database.CreatePlaylistParams{
		ID:                playlist.ID,
		UserID:            playlist.UserID,
		Name:              playlist.Name,
		Description:       nullStringFromStringPtr(&playlist.Description),
		SpotifyPlaylistID: nullStringFromStringPtr(playlist.SpotifyPlaylistID),
		IsPublic:          playlist.IsPublic,
		SyncWithSpotify:   playlist.SyncWithSpotify,
	})
	if err != nil {
		return err
	}

	// Update the playlist with the returned values
	playlist.CreatedAt = dbPlaylist.CreatedAt
	playlist.UpdatedAt = dbPlaylist.UpdatedAt
	return nil
}

func (s *SQLPlaylistStore) GetPlaylistByID(ctx context.Context, id uuid.UUID) (*models.Playlist, error) {
	dbPlaylist, err := s.db.GetPlaylistByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertDbPlaylistToModel(dbPlaylist), nil
}

func (s *SQLPlaylistStore) GetPlaylistByUserIDAndID(ctx context.Context, id, userID uuid.UUID) (*models.Playlist, error) {
	dbPlaylist, err := s.db.GetPlaylistByUserIDAndID(ctx, database.GetPlaylistByUserIDAndIDParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return convertDbPlaylistToModel(dbPlaylist), nil
}

func (s *SQLPlaylistStore) GetPlaylistsByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Playlist, error) {
	dbPlaylists, err := s.db.GetPlaylistsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	playlists := make([]*models.Playlist, len(dbPlaylists))
	for i, dbPlaylist := range dbPlaylists {
		playlists[i] = convertDbPlaylistToModel(dbPlaylist)
	}
	return playlists, nil
}

func (s *SQLPlaylistStore) UpdatePlaylist(ctx context.Context, playlist *models.Playlist) error {
	return s.db.UpdatePlaylist(ctx, database.UpdatePlaylistParams{
		ID:              playlist.ID,
		Name:            playlist.Name,
		Description:     nullStringFromStringPtr(&playlist.Description),
		IsPublic:        playlist.IsPublic,
		SyncWithSpotify: playlist.SyncWithSpotify,
		UserID:          playlist.UserID,
	})
}

func (s *SQLPlaylistStore) UpdatePlaylistSyncTime(ctx context.Context, id, userID uuid.UUID) error {
	return s.db.UpdatePlaylistSyncTime(ctx, database.UpdatePlaylistSyncTimeParams{
		ID:     id,
		UserID: userID,
	})
}

func (s *SQLPlaylistStore) DeletePlaylist(ctx context.Context, id, userID uuid.UUID) error {
	return s.db.DeletePlaylist(ctx, database.DeletePlaylistParams{
		ID:     id,
		UserID: userID,
	})
}

// Playlist song operations
func (s *SQLPlaylistStore) AddSongToPlaylist(ctx context.Context, song *models.PlaylistSong) error {
	dbSong, err := s.db.AddSongToPlaylist(ctx, database.AddSongToPlaylistParams{
		ID:             song.ID,
		PlaylistID:     song.PlaylistID,
		SpotifyTrackID: song.SpotifyTrackID,
		Position:       int32(song.Position),
		TrackName:      song.TrackName,
		ArtistName:     song.ArtistName,
		AlbumName:      nullStringFromStringPtr(&song.AlbumName),
		DurationMs:     nullInt32FromIntPtr(&song.DurationMs),
	})
	if err != nil {
		return err
	}

	// Update the song with the returned values
	song.AddedAt = dbSong.AddedAt
	return nil
}

func (s *SQLPlaylistStore) GetPlaylistSongs(ctx context.Context, playlistID uuid.UUID) ([]*models.PlaylistSong, error) {
	dbSongs, err := s.db.GetPlaylistSongs(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	songs := make([]*models.PlaylistSong, len(dbSongs))
	for i, dbSong := range dbSongs {
		songs[i] = convertDbPlaylistSongToModel(dbSong)
	}
	return songs, nil
}

func (s *SQLPlaylistStore) GetPlaylistSongByID(ctx context.Context, songID, playlistID uuid.UUID) (*models.PlaylistSong, error) {
	dbSong, err := s.db.GetPlaylistSongByID(ctx, database.GetPlaylistSongByIDParams{
		ID:         songID,
		PlaylistID: playlistID,
	})
	if err != nil {
		return nil, err
	}
	return convertDbPlaylistSongToModel(dbSong), nil
}

func (s *SQLPlaylistStore) RemoveSongFromPlaylist(ctx context.Context, songID, playlistID uuid.UUID) error {
	return s.db.RemoveSongFromPlaylist(ctx, database.RemoveSongFromPlaylistParams{
		ID:         songID,
		PlaylistID: playlistID,
	})
}

func (s *SQLPlaylistStore) UpdateSongPosition(ctx context.Context, songID uuid.UUID, position int) error {
	return s.db.UpdateSongPosition(ctx, database.UpdateSongPositionParams{
		ID:       songID,
		Position: int32(position),
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

// Helper functions for conversion
func convertDbPlaylistToModel(dbPlaylist database.Playlist) *models.Playlist {
	return &models.Playlist{
		ID:                dbPlaylist.ID,
		UserID:            dbPlaylist.UserID,
		Name:              dbPlaylist.Name,
		Description:       stringFromNullString(dbPlaylist.Description),
		SpotifyPlaylistID: stringPtrFromNullString(dbPlaylist.SpotifyPlaylistID),
		IsPublic:          dbPlaylist.IsPublic,
		SyncWithSpotify:   dbPlaylist.SyncWithSpotify,
		LastSyncedAt:      timePtrFromNullTime(dbPlaylist.LastSyncedAt),
		CreatedAt:         dbPlaylist.CreatedAt,
		UpdatedAt:         dbPlaylist.UpdatedAt,
	}
}

func convertDbPlaylistSongToModel(dbSong database.PlaylistSong) *models.PlaylistSong {
	return &models.PlaylistSong{
		ID:             dbSong.ID,
		PlaylistID:     dbSong.PlaylistID,
		SpotifyTrackID: dbSong.SpotifyTrackID,
		Position:       int(dbSong.Position),
		TrackName:      dbSong.TrackName,
		ArtistName:     dbSong.ArtistName,
		AlbumName:      stringFromNullString(dbSong.AlbumName),
		DurationMs:     intFromNullInt32(dbSong.DurationMs),
		AddedAt:        dbSong.AddedAt,
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

func nullInt32FromIntPtr(i *int) sql.NullInt32 {
	if i == nil {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{Int32: int32(*i), Valid: true}
}

func intFromNullInt32(ni sql.NullInt32) int {
	if !ni.Valid {
		return 0
	}
	return int(ni.Int32)
}

func timePtrFromNullTime(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}
