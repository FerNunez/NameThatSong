package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	m "github.com/FerNunez/NameThatSong/internal/models"
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
	GetArtist(ctx context.Context, artistID m.SpotifyID) (*m.ArtistData, error)
	StoreArtist(ctx context.Context, artist *m.ArtistData) error
	GetMultipleArtists(ctx context.Context, artistIDs []m.SpotifyID) (map[m.SpotifyID]*m.ArtistData, []m.SpotifyID, error)
	StoreMultipleArtists(ctx context.Context, artists []*m.ArtistData) error

	// Album operations
	GetAlbum(ctx context.Context, albumID m.SpotifyID) (*m.AlbumData, error)
	StoreAlbum(ctx context.Context, album *m.AlbumData) error
	GetMultipleAlbums(ctx context.Context, albumIDs []m.SpotifyID) (map[m.SpotifyID]*m.AlbumData, []m.SpotifyID, error)
	StoreMultipleAlbums(ctx context.Context, albums []*m.AlbumData) error

	// Track operations
	GetTrack(ctx context.Context, trackID m.SpotifyID) (*m.TrackData, error)
	StoreTrack(ctx context.Context, track *m.TrackData) error
	GetMultipleTracks(ctx context.Context, trackIDs []m.SpotifyID) (map[m.SpotifyID]*m.TrackData, []m.SpotifyID, error)
	StoreMultipleTracks(ctx context.Context, tracks []*m.TrackData) error

	// Playlist cache operations
	GetPlaylist(ctx context.Context, playlistID m.SpotifyID) (*m.PlaylistData, error)
	StorePlaylist(ctx context.Context, playlist *m.PlaylistData) error
	GetMultiplePlaylists(ctx context.Context, playlistIDs []m.SpotifyID) (map[m.SpotifyID]*m.PlaylistData, []m.SpotifyID, error)
	StoreMultiplePlaylists(ctx context.Context, playlists []*m.PlaylistData) error

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

func (s *SQLSpotifyDataStore) GetArtist(ctx context.Context, artistID m.SpotifyID) (*m.ArtistData, error) {
	dbArtist, err := s.db.GetSpotifyArtist(ctx, string(artistID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found, not an error for cache layer
		}
		return nil, err
	}

	return convertDbArtistToModel(dbArtist), nil
}

func (s *SQLSpotifyDataStore) StoreArtist(ctx context.Context, artist *m.ArtistData) error {
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

func (s *SQLSpotifyDataStore) GetAlbum(ctx context.Context, albumID m.SpotifyID) (*m.AlbumData, error) {
	dbAlbum, err := s.db.GetSpotifyAlbum(ctx, string(albumID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return convertDbAlbumToModel(dbAlbum), nil
}

func (s *SQLSpotifyDataStore) StoreAlbum(ctx context.Context, album *m.AlbumData) error {
	// Convert artist IDs and track IDs to string slices for database storage
	artistIDStrings := make([]string, len(album.ArtistIDs))
	for i, id := range album.ArtistIDs {
		artistIDStrings[i] = string(id)
	}
	trackIDStrings := make([]string, len(album.TrackIDs))
	for i, id := range album.TrackIDs {
		trackIDStrings[i] = string(id)
	}

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
		ArtistIds:            artistIDStrings,
		TrackIds:             trackIDStrings,
		CachedAt:             album.CachedAt,
	})
	if err != nil {
		return err
	}
	return nil
}

// =============================================================================
// TRACK OPERATIONS
// =============================================================================

func (s *SQLSpotifyDataStore) GetTrack(ctx context.Context, trackID m.SpotifyID) (*m.TrackData, error) {
	dbTrack, err := s.db.GetSpotifyTrack(ctx, string(trackID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return convertDbTrackToModel(dbTrack), nil
}

func (s *SQLSpotifyDataStore) StoreTrack(ctx context.Context, track *m.TrackData) error {
	// Note: track.Album field doesn't exist in current model structure
	// Using track.AlbumID instead

	// Convert artist IDs to string slice for database storage
	artistIDStrings := make([]string, len(track.ArtistIDs))
	for i, id := range track.ArtistIDs {
		artistIDStrings[i] = string(id)
	}

	_, err := s.db.UpsertSpotifyTrack(ctx, database.UpsertSpotifyTrackParams{
		ID:          track.ID,
		Name:        track.Name,
		AlbumID:     string(track.AlbumID),
		DurationMs:  int32(track.DurationMs),
		DiscNumber:  nullInt32FromInt(track.DiscNumber),
		TrackNumber: nullInt32FromInt(track.TrackNumber),
		Popularity:  nullInt32FromInt(track.Popularity),
		Explicit:    nullBoolFromBool(track.Explicit),
		IsLocal:     nullBoolFromBool(track.IsLocal),
		ArtistIds:   artistIDStrings,
		CachedAt:    track.CachedAt,
	})
	if err != nil {
		return err
	}

	// Note: Track-artist relationship storage is handled via the artist_ids array field
	// in the new data structure, not via separate junction tables

	return nil
}

// Batch track operations
// GetMultipleTracks efficiently fetches multiple tracks using chunked batch queries
func (s *SQLSpotifyDataStore) GetMultipleTracks(ctx context.Context, trackIDs []m.SpotifyID) (map[m.SpotifyID]*m.TrackData, []m.SpotifyID, error) {
	if len(trackIDs) == 0 {
		return make(map[m.SpotifyID]*m.TrackData), []m.SpotifyID{}, nil
	}

	found := make(map[m.SpotifyID]*m.TrackData)
	foundIDs := make(map[m.SpotifyID]bool)

	// Process tracks in chunks to avoid PostgreSQL parameter limits
	for i := 0; i < len(trackIDs); i += MaxBatchSize {
		end := i + MaxBatchSize
		if end > len(trackIDs) {
			end = len(trackIDs)
		}

		batch := trackIDs[i:end]
		// Convert m.SpotifyID slice to string slice for database call
		strBatch := make([]string, len(batch))
		for j, id := range batch {
			strBatch[j] = string(id)
		}
		dbTracks, err := s.db.GetMultipleSpotifyTracks(ctx, strBatch)
		if err != nil {
			return nil, []m.SpotifyID{}, fmt.Errorf("failed to batch fetch tracks: %w", err)
		}

		for _, dbTrack := range dbTracks {
			// Convert database model to domain model
			track := convertDbTrackToModel(dbTrack)

			// TODO: consider batch all related data (album, artists) in single query
			// Currently loading album and artists separately - could be optimized
			if dbTrack.AlbumID != "" {
				albumData, err := s.GetAlbum(ctx, m.SpotifyID(dbTrack.AlbumID))
				if err == nil && albumData != nil {
					// Note: track.Album field doesn't exist in current model, skip this
				}
			}
			found[m.SpotifyID(dbTrack.ID)] = track
			foundIDs[m.SpotifyID(dbTrack.ID)] = true
		}
	}

	// Identify missing tracks
	var missing []m.SpotifyID
	for _, trackID := range trackIDs {
		if !foundIDs[trackID] {
			missing = append(missing, trackID)
		}
	}
	return found, missing, nil
}

// StoreMultipleTracks efficiently stores multiple tracks using JSON-based batch operations
func (s *SQLSpotifyDataStore) StoreMultipleTracks(ctx context.Context, tracks []*m.TrackData) error {
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
			albumID := string(track.AlbumID)
			// Note: Using track.AlbumID instead of track.Album.ID

			// Convert artist IDs to string slice for JSON
			artistIDStrings := make([]string, len(track.ArtistIDs))
			for k, id := range track.ArtistIDs {
				artistIDStrings[k] = string(id)
			}
			jsonTracks[j] = map[string]any{
				"id":           track.ID,
				"name":         track.Name,
				"album_id":     albumID,
				"duration_ms":  track.DurationMs,
				"disc_number":  track.DiscNumber,
				"track_number": track.TrackNumber,
				"popularity":   track.Popularity,
				"explicit":     track.Explicit,
				"is_local":     track.IsLocal,
				"artist_ids":   artistIDStrings,
				"cached_at":    track.CachedAt.Format(time.RFC3339),
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
	}

	return nil
}

// Albums
// Takes slice of albums and does multiple batch fetching
func (s *SQLSpotifyDataStore) GetMultipleAlbums(ctx context.Context, albumIDs []m.SpotifyID) (map[m.SpotifyID]*m.AlbumData, []m.SpotifyID, error) {
	if len(albumIDs) == 0 {
		return make(map[m.SpotifyID]*m.AlbumData), []m.SpotifyID{}, nil
	}

	// Get albums by batches and mark what found
	found := make(map[m.SpotifyID]*m.AlbumData, len(albumIDs))
	foundIDs := make(map[m.SpotifyID]bool, len(albumIDs))
	for i := 0; i < len(albumIDs); i += MaxBatchSize {
		end := min(i+MaxBatchSize, len(albumIDs))

		batch := albumIDs[i:end]
		// Convert m.SpotifyID slice to string slice for database call
		strBatch := make([]string, len(batch))
		for j, id := range batch {
			strBatch[j] = string(id)
		}
		// Database batch get
		dbAlbums, err := s.db.GetMultipleSpotifyAlbums(ctx, strBatch) // this will not return an error if some ID doenst exist. It return fewer results
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch album batch %d-%d: %w", i, end-1, err)
		}
		for _, dbAlbum := range dbAlbums {
			// Convert database model to domain model
			found[m.SpotifyID(dbAlbum.ID)] = convertDbAlbumToModel(dbAlbum)
			foundIDs[m.SpotifyID(dbAlbum.ID)] = true
		}
	}

	// Check missing albums not found
	var missing []m.SpotifyID
	for _, albumID := range albumIDs {
		if !foundIDs[albumID] {
			missing = append(missing, albumID)
		}
	}
	return found, missing, nil
}

// StoreMultipleAlbums efficiently stores multiple albums using JSON-based batch operations
func (s *SQLSpotifyDataStore) StoreMultipleAlbums(ctx context.Context, albums []*m.AlbumData) error {
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
func (s *SQLSpotifyDataStore) GetMultipleArtists(ctx context.Context, artistIDs []m.SpotifyID) (map[m.SpotifyID]*m.ArtistData, []m.SpotifyID, error) {
	if len(artistIDs) == 0 {
		return make(map[m.SpotifyID]*m.ArtistData), []m.SpotifyID{}, nil
	}

	// Get artists by batches and mark what found
	found := make(map[m.SpotifyID]*m.ArtistData, len(artistIDs))
	foundIDs := make(map[m.SpotifyID]bool, len(artistIDs))
	for i := 0; i < len(artistIDs); i += MaxBatchSize {
		end := min(i+MaxBatchSize, len(artistIDs))

		batch := artistIDs[i:end]
		// Convert m.SpotifyID slice to string slice for database call
		strBatch := make([]string, len(batch))
		for j, id := range batch {
			strBatch[j] = string(id)
		}
		// Database batch get
		dbArtists, err := s.db.GetMultipleSpotifyArtists(ctx, strBatch) // this will not return an error if some ID doenst exist. It return fewer results
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch artist batch %d-%d: %w", i, end-1, err)
		}
		for _, dbArtist := range dbArtists {
			// Convert database model to domain model
			found[m.SpotifyID(dbArtist.ID)] = convertDbArtistToModel(dbArtist)
			foundIDs[m.SpotifyID(dbArtist.ID)] = true
		}
	}

	// Check missing artists not found
	var missing []m.SpotifyID
	for _, artistID := range artistIDs {
		if !foundIDs[artistID] {
			missing = append(missing, artistID)
		}
	}
	return found, missing, nil
}

// StoreMultipleArtists efficiently stores multiple artists using JSON-based batch operations
func (s *SQLSpotifyDataStore) StoreMultipleArtists(ctx context.Context, artists []*m.ArtistData) error {
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
// Takes slice of playlists and does multiple batch fetching. Return found map and missing IDs, error
func (s *SQLSpotifyDataStore) GetMultiplePlaylists(ctx context.Context, playlistIDs []m.SpotifyID) (map[m.SpotifyID]*m.PlaylistData, []m.SpotifyID, error) {
	if len(playlistIDs) == 0 {
		return make(map[m.SpotifyID]*m.PlaylistData), []m.SpotifyID{}, nil
	}

	// Get playlists by batches and mark what found
	found := make(map[m.SpotifyID]*m.PlaylistData, len(playlistIDs))
	foundIDs := make(map[m.SpotifyID]bool, len(playlistIDs))
	for i := 0; i < len(playlistIDs); i += MaxBatchSize {
		end := min(i+MaxBatchSize, len(playlistIDs))

		batch := playlistIDs[i:end]
		// Convert m.SpotifyID slice to string slice for database call
		strBatch := make([]string, len(batch))
		for j, id := range batch {
			strBatch[j] = string(id)
		}
		// Database batch get
		dbPlaylists, err := s.db.GetMultipleSpotifyPlaylists(ctx, strBatch) // this will not return an error if some ID doenst exist. It return fewer results
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch playlist batch %d-%d: %w", i, end-1, err)
		}
		for _, dbPlaylist := range dbPlaylists {
			// Convert database model to domain model
			found[m.SpotifyID(dbPlaylist.ID)] = convertDbPlaylistToModel(dbPlaylist)
			foundIDs[m.SpotifyID(dbPlaylist.ID)] = true
		}
	}

	// Check missing playlists not found
	var missing []m.SpotifyID
	for _, playlistID := range playlistIDs {
		if !foundIDs[playlistID] {
			missing = append(missing, playlistID)
		}
	}
	return found, missing, nil
}

// StoreMultiplePlaylists efficiently stores multiple playlists using JSON-based batch operations
func (s *SQLSpotifyDataStore) StoreMultiplePlaylists(ctx context.Context, playlists []*m.PlaylistData) error {
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
				"track_ids":          playlist.TrackIDs,
				"cached_at":          playlist.CachedAt,
			}
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

func (s *SQLSpotifyDataStore) GetPlaylist(ctx context.Context, playlistID m.SpotifyID) (*m.PlaylistData, error) {
	dbPlaylist, err := s.db.GetSpotifyPlaylist(ctx, string(playlistID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return convertDbPlaylistToModel(dbPlaylist), nil
}

func (s *SQLSpotifyDataStore) StorePlaylist(ctx context.Context, playlist *m.PlaylistData) error {
	// Convert track IDs to string slice for database storage
	trackIDStrings := make([]string, len(playlist.TrackIDs))
	for i, id := range playlist.TrackIDs {
		trackIDStrings[i] = string(id)
	}

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
		TrackIds:         trackIDStrings,
		CachedAt:         playlist.CachedAt,
	})
	return err
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
func convertDbArtistToModel(dbArtist database.SpotifyArtist) *m.ArtistData {
	return &m.ArtistData{
		ID:             dbArtist.ID,
		Name:           dbArtist.Name,
		ImageURL:       nullStringToString(dbArtist.ImageUrl),
		Popularity:     nullInt32ToInt(dbArtist.Popularity),
		FollowersTotal: int(dbArtist.FollowersTotal),
		Genres:         dbArtist.Genres,
		CachedAt:       dbArtist.CachedAt,
	}
}

func convertDbAlbumToModel(dbAlbum database.SpotifyAlbum) *m.AlbumData {
	// Get releaseDate
	var releaseDate string
	if !dbAlbum.ReleaseDate.Valid {
		releaseDate = "Unknown date"
	} else {
		releaseDate = dbAlbum.ReleaseDate.Time.Format("2006-01-02")
	}

	// Convert artist IDs from string slice to m.SpotifyID slice
	artistIDs := make([]m.SpotifyID, len(dbAlbum.ArtistIds))
	for i, id := range dbAlbum.ArtistIds {
		artistIDs[i] = m.SpotifyID(id)
	}

	// Convert track IDs from string slice to m.SpotifyID slice
	trackIDs := make([]m.SpotifyID, len(dbAlbum.TrackIds))
	for i, id := range dbAlbum.TrackIds {
		trackIDs[i] = m.SpotifyID(id)
	}

	return &m.AlbumData{
		ID:                   dbAlbum.ID,
		Name:                 dbAlbum.Name,
		AlbumType:            dbAlbum.AlbumType,
		ReleaseDate:          releaseDate,
		ReleaseDatePrecision: dbAlbum.ReleaseDatePrecision,
		TotalTracks:          int(dbAlbum.TotalTracks),
		ImageURL:             nullStringToString(dbAlbum.ImageUrl),
		Label:                nullStringToString(dbAlbum.Label),
		Popularity:           int(dbAlbum.Popularity),
		ArtistIDs:            artistIDs,
		TrackIDs:             trackIDs,
		CachedAt:             dbAlbum.CachedAt,
	}
}

func convertDbTrackToModel(dbTrack database.SpotifyTrack) *m.TrackData {
	// Convert artist IDs from string slice to m.SpotifyID slice
	artistIDs := make([]m.SpotifyID, len(dbTrack.ArtistIds))
	for i, id := range dbTrack.ArtistIds {
		artistIDs[i] = m.SpotifyID(id)
	}

	return &m.TrackData{
		ID:          dbTrack.ID,
		Name:        dbTrack.Name,
		DurationMs:  int(dbTrack.DurationMs),
		DiscNumber:  nullInt32ToInt(dbTrack.DiscNumber),
		TrackNumber: nullInt32ToInt(dbTrack.TrackNumber),
		Popularity:  nullInt32ToInt(dbTrack.Popularity),
		Explicit:    nullBoolToBool(dbTrack.Explicit),
		IsLocal:     nullBoolToBool(dbTrack.IsLocal),
		AlbumID:     m.SpotifyID(dbTrack.AlbumID),
		ArtistIDs:   artistIDs,
		CachedAt:    dbTrack.CachedAt,
	}
}

func convertDbPlaylistToModel(dbPlaylist database.SpotifyPlaylist) *m.PlaylistData {
	// Convert track IDs from string slice to m.SpotifyID slice
	trackIDs := make([]m.SpotifyID, len(dbPlaylist.TrackIds))
	for i, id := range dbPlaylist.TrackIds {
		trackIDs[i] = m.SpotifyID(id)
	}

	return &m.PlaylistData{
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
		TrackIDs:         trackIDs,
		CachedAt:         dbPlaylist.CachedAt,
	}
}

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
