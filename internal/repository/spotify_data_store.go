package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/repository/database"
)

// Batch operation configuration
const (
	// MaxBatchSize defines the maximum number of items to process in a single batch
	// This prevents PostgreSQL parameter limits and memory issues
	MaxBatchSize = 500

	// DefaultBatchSize is the recommended batch size for optimal performance
	DefaultBatchSize = 100
)

// SpotifyDataStore defines the interface for Spotify data persistence operations
type SpotifyDataStore interface {
	// Artist operations
	GetArtist(ctx context.Context, artistID string) (*models.ArtistData, error)
	StoreArtist(ctx context.Context, artist *models.ArtistData) error
	GetMultipleArtists(ctx context.Context, artistIDs []string) (map[string]*models.ArtistData, []string, error)
	StoreMultipleArtists(ctx context.Context, artists []*models.ArtistData) error

	// Album operations
	GetAlbum(ctx context.Context, albumID string) (*models.AlbumData, error)
	StoreAlbum(ctx context.Context, album *models.AlbumData) error
	GetMultipleAlbums(ctx context.Context, albumIDs []string) (map[string]*models.AlbumData, []string, error)
	StoreMultipleAlbums(ctx context.Context, albums []*models.AlbumData) error

	// Track operations
	GetTrack(ctx context.Context, trackID string) (*models.TrackData, error)
	StoreTrack(ctx context.Context, track *models.TrackData) error
	GetMultipleTracks(ctx context.Context, trackIDs []string) (map[string]*models.TrackData, []string, error)
	StoreMultipleTracks(ctx context.Context, tracks []*models.TrackData) error

	// Playlist cache operations
	GetPlaylist(ctx context.Context, playlistID string) (*models.PlaylistData, error)
	StorePlaylist(ctx context.Context, playlist *models.PlaylistData) error
	GetMultiplePlaylists(ctx context.Context, playlistIDs []string) (map[string]*models.PlaylistData, []string, error)
	StoreMultiplePlaylists(ctx context.Context, playlists []*models.PlaylistData) error

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
		FollowersTotal: int32(artist.FollowersTotal),
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
		ReleaseDatePrecision: album.ReleaseDatePrecision,
		TotalTracks:          int32(album.TotalTracks),
		ImageUrl:             nullStringFromString(album.ImageURL),
		Label:                nullStringFromString(album.Label),
		Popularity:           int32(album.Popularity),
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

// Batch track operations
// GetMultipleTracks efficiently fetches multiple tracks using chunked batch queries
func (s *SQLSpotifyDataStore) GetMultipleTracks(ctx context.Context, trackIDs []string) (map[string]*models.TrackData, []string, error) {
	if len(trackIDs) == 0 {
		return make(map[string]*models.TrackData), []string{}, nil
	}

	found := make(map[string]*models.TrackData)
	foundIDs := make(map[string]bool)

	// Process tracks in chunks to avoid PostgreSQL parameter limits
	for i := 0; i < len(trackIDs); i += MaxBatchSize {
		end := i + MaxBatchSize
		if end > len(trackIDs) {
			end = len(trackIDs)
		}

		batch := trackIDs[i:end]
		batchFound, err := s.getTrackBatch(ctx, batch)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch batch %d-%d: %w", i, end-1, err)
		}

		// Merge batch results
		for trackID, track := range batchFound {
			found[trackID] = track
			foundIDs[trackID] = true
		}
	}

	// Identify missing tracks
	var missing []string
	for _, trackID := range trackIDs {
		if !foundIDs[trackID] {
			missing = append(missing, trackID)
		}
	}

	return found, missing, nil
}

// getTrackBatch fetches a single batch of tracks using the ANY operator
func (s *SQLSpotifyDataStore) getTrackBatch(ctx context.Context, trackIDs []string) (map[string]*models.TrackData, error) {
	if len(trackIDs) == 0 {
		return make(map[string]*models.TrackData), nil
	}

	dbTracks, err := s.db.GetMultipleSpotifyTracks(ctx, trackIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to batch fetch tracks: %w", err)
	}

	found := make(map[string]*models.TrackData)
	for _, dbTrack := range dbTracks {
		// Convert database model to domain model

		track := convertDbTrackToModel(dbTrack)

		// TODO: consider batch all related data (album, artists) in single query
		// Currently loading album and artists separately - could be optimized
		if dbTrack.AlbumID.Valid {
			albumData, err := s.GetAlbum(ctx, dbTrack.AlbumID.String)
			if err == nil && albumData != nil {
				track.Album = albumData
			}
		}

		found[dbTrack.ID] = track
	}

	return found, nil
}

// StoreMultipleTracks efficiently stores multiple tracks using JSON-based batch operations
func (s *SQLSpotifyDataStore) StoreMultipleTracks(ctx context.Context, tracks []*models.TrackData) error {
	if len(tracks) == 0 {
		return nil
	}

	// Process tracks in chunks to avoid PostgreSQL memory limits and improve performance
	for i := 0; i < len(tracks); i += MaxBatchSize {
		end := min(i+MaxBatchSize, len(tracks))

		batch := tracks[i:end]
		// Convert tracks to JSON format
		jsonTracks := make([]map[string]any, len(batch))
		for j, track := range batch {
			albumID := ""
			if track.Album != nil {
				albumID = track.Album.ID
			}

			// TODO: ARTIST MISSING!
			jsonTracks[j] = map[string]any{
				"id":           track.ID,
				"name":         track.Name,
				"album_id":     albumID,
				"duration_ms":  track.DurationMs,
				"disc_number":  track.DiscNumber,
				"track_number": track.TrackNumber,
				"popularity":   track.Popularity,
				"explicit":     track.Explicit,
				"preview_url":  track.PreviewURL,
				"is_local":     track.IsLocal,
			}
		}

		// Marshal to JSON bytes
		jsonBytes, err := json.Marshal(jsonTracks)
		if err != nil {
			return fmt.Errorf("failed to marshal tracks to JSON: %w", err)
		}

		// Execute batch upsert
		if err := s.db.UpsertMultipleSpotifyTracksFromJSON(ctx, jsonBytes); err != nil {
			return fmt.Errorf("failed to execute batch upsert: %w", err)
		}

		// TODO: PRIORITY 1 - Add Relationship Storage for StoreMultipleTracks
		// STEP 1: Extract and deduplicate all albums from this batch:
		//   - Collect unique album IDs from tracks that have track.Album != nil
		//   - Call StoreMultipleAlbums() with the collected album data
		//
		// STEP 2: Extract and deduplicate all artists from this batch:
		//   - Collect all track.Artists[].ArtistData from all tracks
		//   - Collect all album.Artists from all track.Album.Artists (if album exists)
		//   - Deduplicate by artist.ID and call StoreMultipleArtists()
		//
		// STEP 3: Store track-artist relationships:
		//   - For each track in batch: Clear existing relationships: s.db.ClearTrackArtists(ctx, track.ID)
		//   - For each track.Artists[]: Call s.db.UpsertTrackArtist() with IsPrimary flag
		//
		// STEP 4: Store album-track relationships (if needed):
		//   - For each track with Album: Call s.db.UpsertAlbumTrack() with position
		//
		// Currently we only store track metadata - all relationships are MISSING!

	}

	return nil
}

// Albums
// Takes slice of albums and does multiple batch fetching
func (s *SQLSpotifyDataStore) GetMultipleAlbums(ctx context.Context, albumIDs []string) (map[string]*models.AlbumData, []string, error) {
	if len(albumIDs) == 0 {
		return make(map[string]*models.AlbumData), []string{}, nil
	}

	// Get albums by batches and mark what found
	found := make(map[string]*models.AlbumData, len(albumIDs))
	foundIDs := make(map[string]bool, len(albumIDs))
	for i := 0; i < len(albumIDs); i += MaxBatchSize {
		end := min(i+MaxBatchSize, len(albumIDs))

		batch := albumIDs[i:end]
		// Database batch get
		dbAlbums, err := s.db.GetMultipleSpotifyAlbums(ctx, batch) // this will not return an error if some ID doenst exist. It return fewer results
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch album batch %d-%d: %w", i, end-1, err)
		}
		for _, dbAlbum := range dbAlbums {
			// Convert database model to domain model
			found[dbAlbum.ID] = convertDbAlbumToModel(dbAlbum)
			foundIDs[dbAlbum.ID] = true
		}
	}

	// Check missing albums not found
	var missing []string
	for _, albumID := range albumIDs {
		if !foundIDs[albumID] {
			missing = append(missing, albumID)
		}
	}
	return found, missing, nil
}

// StoreMultipleAlbums efficiently stores multiple albums using JSON-based batch operations
func (s *SQLSpotifyDataStore) StoreMultipleAlbums(ctx context.Context, albums []*models.AlbumData) error {
	if len(albums) == 0 {
		return nil
	}

	for i := 0; i < len(albums); i += MaxBatchSize {
		end := min(i+MaxBatchSize, len(albums))
		batch := albums[i:end]

		// Create json and store using upsert
		jsonAlbums := make([]map[string]any, len(batch))
		for j, album := range batch {
			jsonAlbums[j] = map[string]any{
				"id":                     album.ID,
				"name":                   album.Name,
				"album_type":             album.AlbumType,
				"release_date":           album.ReleaseDate,
				"release_date_precision": album.ReleaseDatePrecision,
				"total_tracks":           album.TotalTracks,
				"image_url":              album.ImageURL,
				"label":                  album.Label,
				"popularity":             album.Popularity,
			}
			// TODO: PRIORITY 1 - Add Relationship Storage for StoreMultipleAlbums
			// STEP 1: Extract and deduplicate all artists from this batch:
			//   - Collect all album.Artists[].ArtistData from all albums in batch
			//   - Deduplicate by artist.ID and call StoreMultipleArtists()
			//
			// STEP 2: Store album-artist relationships:
			//   - For each album in batch: Clear existing relationships: s.db.ClearAlbumArtists(ctx, album.ID)
			//   - For each album.Artists[]: Call s.db.UpsertAlbumArtist()
			//
			// Currently we only store album metadata - artist relationships are MISSING!
		}

		// Marshal to JSON bytes
		jsonBytes, err := json.Marshal(jsonAlbums)
		if err != nil {
			return fmt.Errorf("failed to marshal albums to JSON: %w", err)
		}

		// Execute batch upsert
		if err := s.db.UpsertMultipleSpotifyAlbumsFromJSON(ctx, jsonBytes); err != nil {
			return fmt.Errorf("failed to execute albums batch upsert: %w", err)
		}
	}
	return nil
}

// Artists
// Takes slice of artists and does multiple batch fetching
func (s *SQLSpotifyDataStore) GetMultipleArtists(ctx context.Context, artistIDs []string) (map[string]*models.ArtistData, []string, error) {
	if len(artistIDs) == 0 {
		return make(map[string]*models.ArtistData), []string{}, nil
	}

	// Get artists by batches and mark what found
	found := make(map[string]*models.ArtistData, len(artistIDs))
	foundIDs := make(map[string]bool, len(artistIDs))
	for i := 0; i < len(artistIDs); i += MaxBatchSize {
		end := min(i+MaxBatchSize, len(artistIDs))

		batch := artistIDs[i:end]
		// Database batch get
		dbArtists, err := s.db.GetMultipleSpotifyArtists(ctx, batch) // this will not return an error if some ID doenst exist. It return fewer results
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch artist batch %d-%d: %w", i, end-1, err)
		}
		for _, dbArtist := range dbArtists {
			// Convert database model to domain model
			found[dbArtist.ID] = convertDbArtistToModel(dbArtist)
			foundIDs[dbArtist.ID] = true
		}
	}

	// Check missing artists not found
	var missing []string
	for _, artistID := range artistIDs {
		if !foundIDs[artistID] {
			missing = append(missing, artistID)
		}
	}
	return found, missing, nil
}

// StoreMultipleArtists efficiently stores multiple artists using JSON-based batch operations
func (s *SQLSpotifyDataStore) StoreMultipleArtists(ctx context.Context, artists []*models.ArtistData) error {
	if len(artists) == 0 {
		return nil
	}

	for i := 0; i < len(artists); i += MaxBatchSize {
		end := min(i+MaxBatchSize, len(artists))
		batch := artists[i:end]

		// Create json and store using upsert
		jsonArtists := make([]map[string]any, len(batch))
		for j, artist := range batch {
			jsonArtists[j] = map[string]any{
				"id":              artist.ID,
				"name":            artist.Name,
				"image_url":       artist.ImageURL,
				"popularity":      artist.Popularity,
				"followers_total": artist.FollowersTotal,
				"genres":          artist.Genres,
			}
		}

		// Marshal to JSON bytes
		jsonBytes, err := json.Marshal(jsonArtists)
		if err != nil {
			return fmt.Errorf("failed to marshal artists to JSON: %w", err)
		}

		// Execute batch upsert
		if err := s.db.UpsertMultipleSpotifyArtistsFromJSON(ctx, jsonBytes); err != nil {
			return fmt.Errorf("failed to execute batch upsert for artists: %w", err)
		}
	}
	return nil
}

// Playlists
// Takes slice of playlists and does multiple batch fetching
func (s *SQLSpotifyDataStore) GetMultiplePlaylists(ctx context.Context, playlistIDs []string) (map[string]*models.PlaylistData, []string, error) {
	if len(playlistIDs) == 0 {
		return make(map[string]*models.PlaylistData), []string{}, nil
	}

	// Get playlists by batches and mark what found
	found := make(map[string]*models.PlaylistData, len(playlistIDs))
	foundIDs := make(map[string]bool, len(playlistIDs))
	for i := 0; i < len(playlistIDs); i += MaxBatchSize {
		end := min(i+MaxBatchSize, len(playlistIDs))

		batch := playlistIDs[i:end]
		// Database batch get
		dbPlaylists, err := s.db.GetMultipleSpotifyPlaylists(ctx, batch) // this will not return an error if some ID doenst exist. It return fewer results
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch playlist batch %d-%d: %w", i, end-1, err)
		}
		for _, dbPlaylist := range dbPlaylists {
			// Convert database model to domain model
			found[dbPlaylist.ID] = convertDbPlaylistToModel(dbPlaylist)
			foundIDs[dbPlaylist.ID] = true
		}
	}

	// Check missing playlists not found
	var missing []string
	for _, playlistID := range playlistIDs {
		if !foundIDs[playlistID] {
			missing = append(missing, playlistID)
		}
	}
	return found, missing, nil
}

// StoreMultiplePlaylists efficiently stores multiple playlists using JSON-based batch operations
func (s *SQLSpotifyDataStore) StoreMultiplePlaylists(ctx context.Context, playlists []*models.PlaylistData) error {
	if len(playlists) == 0 {
		return nil
	}

	for i := 0; i < len(playlists); i += MaxBatchSize {
		end := min(i+MaxBatchSize, len(playlists))
		batch := playlists[i:end]

		// Create json and store using upsert
		jsonPlaylists := make([]map[string]any, len(batch))
		for i, playlist := range batch {
			jsonPlaylists[i] = map[string]any{
				"id":                 playlist.ID,
				"name":               playlist.Name,
				"description":        playlist.Description,
				"owner_id":           playlist.OwnerID,
				"owner_display_name": playlist.OwnerDisplayName,
				"public":             playlist.Public,
				"collaborative":      playlist.Collaborative,
				"followers_total":    playlist.FollowersTotal,
				"total_tracks":       playlist.TotalTracks,
				"image_url":          playlist.ImageURL,
			}
			// TODO: PRIORITY 1 - Add Relationship Storage for StoreMultiplePlaylists
			// STEP 1: Store playlist-track relationships (if track data available):
			//   - Check if playlist.Tracks data is available (may need to be added to PlaylistData model)
			//   - For each playlist: Clear existing tracks: s.db.ClearPlaylistTracks(ctx, playlist.ID)
			//   - For each track in playlist: Call s.db.UpsertPlaylistTracks() with position
			//
			// NOTE: Currently PlaylistData model doesn't include Tracks field
			// Consider if this should be added or handled separately via UpsertPlaylistTracks()
		}

		// Marshal to JSON bytes
		jsonBytes, err := json.Marshal(jsonPlaylists)
		if err != nil {
			return fmt.Errorf("failed to marshal playlists to JSON: %w", err)
		}

		// Execute batch upsert
		if err := s.db.UpsertMultipleSpotifyPlaylistsFromJSON(ctx, jsonBytes); err != nil {
			return fmt.Errorf("failed to execute batch upsert for playlists: %w", err)
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

	return convertDbPlaylistToModel(dbPlaylist), nil
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
			UpdatedAt:  time.Now(),
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
		FollowersTotal: int(dbArtist.FollowersTotal),
		Genres:         dbArtist.Genres,
		CachedAt:       dbArtist.CachedAt,
		UpdatedAt:      dbArtist.UpdatedAt,
	}
}

func convertDbAlbumToModel(dbAlbum database.SpotifyAlbum) *models.AlbumData {

	var releaseDate string
	if !dbAlbum.ReleaseDate.Valid {
		releaseDate = "Unknown date"
	} else {
		releaseDate = dbAlbum.ReleaseDate.Time.Format("2006-01-02")
	}

	return &models.AlbumData{
		ID:                   dbAlbum.ID,
		Name:                 dbAlbum.Name,
		AlbumType:            dbAlbum.AlbumType,
		ReleaseDate:          releaseDate,
		ReleaseDatePrecision: dbAlbum.ReleaseDatePrecision,
		TotalTracks:          int(dbAlbum.TotalTracks),
		ImageURL:             nullStringToString(dbAlbum.ImageUrl),
		Label:                nullStringToString(dbAlbum.Label),
		Popularity:           int(dbAlbum.Popularity),
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

func convertDbPlaylistToModel(dbPlaylist database.SpotifyPlaylist) *models.PlaylistData {
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
