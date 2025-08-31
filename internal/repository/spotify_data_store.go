package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/repository/database"
)

// SpotifyDataStore defines the interface for Spotify data persistence operations
type SpotifyDataStore interface {
	// Artist operations
	GetArtist(ctx context.Context, artistID string) (*models.ArtistData, error)
	StoreArtist(ctx context.Context, artist *models.ArtistData) error

	// Album operations
	GetAlbum(ctx context.Context, albumID string) (*models.AlbumData, error)
	StoreAlbum(ctx context.Context, album *models.AlbumData) error

	// Track operations
	GetTrack(ctx context.Context, trackID string) (*models.TrackData, error)
	StoreTrack(ctx context.Context, track *models.TrackData) error

	// Playlist cache operations
	GetPlaylist(ctx context.Context, playlistID string) (*models.PlaylistData, error)
	StorePlaylist(ctx context.Context, playlist *models.PlaylistData) error

	// Relations
	UpsertPlaylistTracks(ctx context.Context, playlistID string, trackIDs []string) error
	GetPlaylistTracks(ctx context.Context, playlistID string) ([]string, error)
	DeletePaylistTrack(ctx context.Context, playlistID string) error

	UpsertAlbumTracks(ctx context.Context, albumID string, trackID []string) error
	GetAlbumTracks(ctx context.Context, albumID string) ([]string, error)
	GetAlbumByTrackID(ctx context.Context, trackID string) (string, error)
	ClearAlbumTracks(ctx context.Context, albumID string) error

	// Cache management
	CleanupOldCacheData(ctx context.Context, olderThan time.Duration) error
	GetCacheStats(ctx context.Context) (map[string]int64, error)
}

// SQLSpotifyDataStore implements SpotifyDataStore using PostgreSQL
type SQLSpotifyDataStore struct {
	db *database.Queries
}

// NewSQLSpotifyDataStore creates a new SQL-based Spotify data store
func NewSQLSpotifyDataStore(db *database.Queries) SpotifyDataStore {
	return &SQLSpotifyDataStore{
		db: db,
	}
}

// =============================================================================
// ARTIST OPERATIONS
// =============================================================================

func (s *SQLSpotifyDataStore) GetArtist(ctx context.Context, artistID string) (*models.ArtistData, error) {
	dbArtist, err := s.db.GetSpotifyArtist(ctx, artistID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found, not an error for cache layer
		}
		return nil, err
	}

	return convertDbArtistToModel(dbArtist), nil
}

func (s *SQLSpotifyDataStore) StoreArtist(ctx context.Context, artist *models.ArtistData) error {
	_, err := s.db.UpsertSpotifyArtist(ctx, database.UpsertSpotifyArtistParams{
		ID:             artist.ID,
		Name:           artist.Name,
		ImageUrl:       nullStringFromString(artist.ImageURL),
		Popularity:     nullInt32FromInt(artist.Popularity),
		FollowersTotal: nullInt32FromInt(artist.FollowersTotal),
		Genres:         artist.Genres,
	})
	return err
}

// =============================================================================
// ALBUM OPERATIONS
// =============================================================================

func (s *SQLSpotifyDataStore) GetAlbum(ctx context.Context, albumID string) (*models.AlbumData, error) {
	dbAlbum, err := s.db.GetSpotifyAlbum(ctx, albumID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return convertDbAlbumToModel(dbAlbum), nil
}

func (s *SQLSpotifyDataStore) StoreAlbum(ctx context.Context, album *models.AlbumData) error {
	// Store album
	_, err := s.db.UpsertSpotifyAlbum(ctx, database.UpsertSpotifyAlbumParams{
		ID:                   album.ID,
		Name:                 album.Name,
		AlbumType:            album.AlbumType,
		ReleaseDate:          nullDateFromString(album.ReleaseDate),
		ReleaseDatePrecision: nullStringFromString(album.ReleaseDatePrecision),
		TotalTracks:          nullInt32FromInt(album.TotalTracks),
		ImageUrl:             nullStringFromString(album.ImageURL),
		Label:                nullStringFromString(album.Label),
		Popularity:           nullInt32FromInt(album.Popularity),
	})
	if err != nil {
		return err
	}

	// Clear existing album-artist relationships
	if err := s.db.ClearAlbumArtists(ctx, album.ID); err != nil {
		return err
	}

	// Store album artists and relationships
	for _, artist := range album.Artists {
		// Store artist
		if err := s.StoreArtist(ctx, &artist); err != nil {
			continue // Skip failed artists
		}

		// Store relationship
		if err := s.db.UpsertAlbumArtist(ctx, database.UpsertAlbumArtistParams{
			AlbumID:  album.ID,
			ArtistID: artist.ID,
		}); err != nil {
			continue // Skip failed relationships
		}
	}

	return nil
}

// =============================================================================
// TRACK OPERATIONS
// =============================================================================

func (s *SQLSpotifyDataStore) GetTrack(ctx context.Context, trackID string) (*models.TrackData, error) {
	dbTrack, err := s.db.GetSpotifyTrack(ctx, trackID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return convertDbTrackToModel(dbTrack), nil
}

func (s *SQLSpotifyDataStore) StoreTrack(ctx context.Context, track *models.TrackData) error {
	// Store album if present
	if track.Album != nil {
		if err := s.StoreAlbum(ctx, track.Album); err != nil {
			return fmt.Errorf("failed to store album: %w", err)
		}
	}

	// Store track
	albumID := sql.NullString{}
	if track.Album != nil {
		albumID = sql.NullString{String: track.Album.ID, Valid: true}
	}

	_, err := s.db.UpsertSpotifyTrack(ctx, database.UpsertSpotifyTrackParams{
		ID:          track.ID,
		Name:        track.Name,
		AlbumID:     albumID,
		DurationMs:  int32(track.DurationMs),
		DiscNumber:  nullInt32FromInt(track.DiscNumber),
		TrackNumber: nullInt32FromInt(track.TrackNumber),
		Popularity:  nullInt32FromInt(track.Popularity),
		Explicit:    nullBoolFromBool(track.Explicit),
		PreviewUrl:  nullStringFromString(track.PreviewURL),
		IsLocal:     nullBoolFromBool(track.IsLocal),
	})
	if err != nil {
		return err
	}

	// Clear existing track-artist relationships
	if err := s.db.ClearTrackArtists(ctx, track.ID); err != nil {
		return err
	}

	// Store track artists and relationships
	for _, trackArtist := range track.Artists {
		// Store artist
		if err := s.StoreArtist(ctx, &trackArtist.ArtistData); err != nil {
			continue // Skip failed artists
		}

		// Store relationship
		if err := s.db.UpsertTrackArtist(ctx, database.UpsertTrackArtistParams{
			TrackID:   track.ID,
			ArtistID:  trackArtist.ID,
			IsPrimary: nullBoolFromBool(trackArtist.IsPrimary),
		}); err != nil {
			continue // Skip failed relationships
		}
	}

	return nil
}

// =============================================================================
// PLAYLIST CACHE OPERATIONS
// =============================================================================

func (s *SQLSpotifyDataStore) GetPlaylist(ctx context.Context, playlistID string) (*models.PlaylistData, error) {
	dbPlaylist, err := s.db.GetSpotifyPlaylist(ctx, playlistID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return convertDbPlaylistCacheToModel(dbPlaylist), nil
}

func (s *SQLSpotifyDataStore) StorePlaylist(ctx context.Context, playlist *models.PlaylistData) error {
	_, err := s.db.UpsertSpotifyPlaylist(ctx, database.UpsertSpotifyPlaylistParams{
		ID:               playlist.ID,
		Name:             playlist.Name,
		Description:      nullStringFromString(playlist.Description),
		OwnerID:          playlist.OwnerID,
		OwnerDisplayName: nullStringFromString(playlist.OwnerDisplayName),
		Public:           nullBoolFromBool(playlist.Public),
		Collaborative:    nullBoolFromBool(playlist.Collaborative),
		FollowersTotal:   nullInt32FromInt(playlist.FollowersTotal),
		TotalTracks:      nullInt32FromInt(playlist.TotalTracks),
		ImageUrl:         nullStringFromString(playlist.ImageURL),
	})
	return err
}

// =============================================================================
// RELATIONS
// =============================================================================

func (s *SQLSpotifyDataStore) UpsertPlaylistTracks(ctx context.Context, playlistID string, trackIDs []string) error {

	for idx, trackID := range trackIDs {
		err := s.db.UpsertPlaylistTracks(ctx, database.UpsertPlaylistTracksParams{
			PlaylistID: playlistID,
			TrackID:    trackID,
			Position:   int32(idx),
		})

		if err != nil {
			return err
		}
	}
	return nil
}
func (s *SQLSpotifyDataStore) GetPlaylistTracks(ctx context.Context, playlistID string) ([]string, error) {
	trackIDDBs, err := s.db.GetPlaylistTracks(ctx, playlistID)
	if err != nil {
		return []string{}, err
	}

	trackIDs := make([]string, 0, len(trackIDDBs))
	for _, trackIDDB := range trackIDDBs {
		trackIDs = append(trackIDs, trackIDDB.TrackID)
	}
	return trackIDs, nil
}
func (s *SQLSpotifyDataStore) DeletePaylistTrack(ctx context.Context, trackID string) error {
	err := s.db.DeletePlaylistTrack(ctx, trackID)
	return err
}
func (s *SQLSpotifyDataStore) UpsertAlbumTracks(ctx context.Context, albumID string, trackIDs []string) error {
	for idx, trackID := range trackIDs {
		err := s.db.UpsertAlbumTrack(ctx, database.UpsertAlbumTrackParams{
			AlbumID:  albumID,
			TrackID:  trackID,
			Position: int32(idx + 1),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *SQLSpotifyDataStore) GetAlbumTracks(ctx context.Context, albumID string) ([]string, error) {
	trackDBs, err := s.db.GetAlbumTracks(ctx, albumID)
	if err != nil {
		return []string{}, err
	}

	trackIDs := make([]string, 0, len(trackDBs))
	for _, trackDB := range trackDBs {
		trackIDs = append(trackIDs, trackDB.TrackID)
	}
	return trackIDs, nil
}
func (s *SQLSpotifyDataStore) GetAlbumByTrackID(ctx context.Context, trackID string) (string, error) {
	albumID, err := s.db.GetAlbumByTrackID(ctx, trackID)
	if err != nil {
		return "", err
	}
	return albumID, nil
}
func (s *SQLSpotifyDataStore) ClearAlbumTracks(ctx context.Context, albumID string) error {
	return s.db.ClearAlbumTracks(ctx, albumID)
}

// =============================================================================
// CACHE MANAGEMENT
// =============================================================================

func (s *SQLSpotifyDataStore) CleanupOldCacheData(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	// Cleanup in parallel (these are independent operations)
	if err := s.db.CleanupOldSpotifyArtists(ctx, cutoff); err != nil {
		return fmt.Errorf("failed to cleanup artists: %w", err)
	}

	if err := s.db.CleanupOldSpotifyAlbums(ctx, cutoff); err != nil {
		return fmt.Errorf("failed to cleanup albums: %w", err)
	}

	if err := s.db.CleanupOldSpotifyTracks(ctx, cutoff); err != nil {
		return fmt.Errorf("failed to cleanup tracks: %w", err)
	}

	if err := s.db.CleanupOldSpotifyPlaylists(ctx, cutoff); err != nil {
		return fmt.Errorf("failed to cleanup playlists: %w", err)
	}

	return nil
}

func (s *SQLSpotifyDataStore) GetCacheStats(ctx context.Context) (map[string]int64, error) {
	stats, err := s.db.GetCacheStats(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]int64{
		"artists":   stats.ArtistsCount,
		"albums":    stats.AlbumsCount,
		"tracks":    stats.TracksCount,
		"playlists": stats.PlaylistsCount,
	}, nil
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// Convert database models to domain models
func convertDbArtistToModel(dbArtist database.SpotifyArtist) *models.ArtistData {
	return &models.ArtistData{
		ID:             dbArtist.ID,
		Name:           dbArtist.Name,
		ImageURL:       nullStringToString(dbArtist.ImageUrl),
		Popularity:     nullInt32ToInt(dbArtist.Popularity),
		FollowersTotal: nullInt32ToInt(dbArtist.FollowersTotal),
		Genres:         dbArtist.Genres,
		CachedAt:       dbArtist.CachedAt,
		UpdatedAt:      dbArtist.UpdatedAt,
	}
}

func convertDbAlbumToModel(dbAlbum database.SpotifyAlbum) *models.AlbumData {
	return &models.AlbumData{
		ID:                   dbAlbum.ID,
		Name:                 dbAlbum.Name,
		AlbumType:            dbAlbum.AlbumType,
		ReleaseDate:          dbAlbum.ReleaseDate.Time.Format("2006-01-02"),
		ReleaseDatePrecision: nullStringToString(dbAlbum.ReleaseDatePrecision),
		TotalTracks:          nullInt32ToInt(dbAlbum.TotalTracks),
		ImageURL:             nullStringToString(dbAlbum.ImageUrl),
		Label:                nullStringToString(dbAlbum.Label),
		Popularity:           nullInt32ToInt(dbAlbum.Popularity),
		CachedAt:             dbAlbum.CachedAt,
		UpdatedAt:            dbAlbum.UpdatedAt,
	}
}

func convertDbTrackToModel(dbTrack database.SpotifyTrack) *models.TrackData {
	return &models.TrackData{
		ID:          dbTrack.ID,
		Name:        dbTrack.Name,
		DurationMs:  int(dbTrack.DurationMs),
		DiscNumber:  nullInt32ToInt(dbTrack.DiscNumber),
		TrackNumber: nullInt32ToInt(dbTrack.TrackNumber),
		Popularity:  nullInt32ToInt(dbTrack.Popularity),
		Explicit:    nullBoolToBool(dbTrack.Explicit),
		PreviewURL:  nullStringToString(dbTrack.PreviewUrl),
		IsLocal:     nullBoolToBool(dbTrack.IsLocal),
		CachedAt:    dbTrack.CachedAt,
		UpdatedAt:   dbTrack.UpdatedAt,
	}
}

func convertDbPlaylistCacheToModel(dbPlaylist database.SpotifyPlaylist) *models.PlaylistData {
	return &models.PlaylistData{
		ID:               dbPlaylist.ID,
		Name:             dbPlaylist.Name,
		Description:      nullStringToString(dbPlaylist.Description),
		OwnerID:          dbPlaylist.OwnerID,
		OwnerDisplayName: nullStringToString(dbPlaylist.OwnerDisplayName),
		Public:           nullBoolToBool(dbPlaylist.Public),
		Collaborative:    nullBoolToBool(dbPlaylist.Collaborative),
		FollowersTotal:   nullInt32ToInt(dbPlaylist.FollowersTotal),
		TotalTracks:      nullInt32ToInt(dbPlaylist.TotalTracks),
		ImageURL:         nullStringToString(dbPlaylist.ImageUrl),
		CachedAt:         dbPlaylist.CachedAt,
		UpdatedAt:        dbPlaylist.UpdatedAt,
	}
}

// TODO: Implement buildTrackFromRows when complex relationship queries are needed
// This method would be used for fetching tracks with full album and artist relationships
// in a single query, but for now we use the simpler approach of separate queries

// Null handling helpers
func nullStringFromString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullStringToString(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}

func nullDateFromString(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{Valid: false}
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return sql.NullTime{Time: t, Valid: true}
	}
	return sql.NullTime{Valid: false}
}

func nullInt32FromInt(i int) sql.NullInt32 {
	return sql.NullInt32{Int32: int32(i), Valid: true}
}

func nullInt32ToInt(ni sql.NullInt32) int {
	if !ni.Valid {
		return 0
	}
	return int(ni.Int32)
}

func nullBoolFromBool(b bool) sql.NullBool {
	return sql.NullBool{Bool: b, Valid: true}
}

func nullBoolToBool(nb sql.NullBool) bool {
	if !nb.Valid {
		return false
	}
	return nb.Bool
}
