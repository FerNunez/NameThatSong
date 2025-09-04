package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	m "github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
	"github.com/FerNunez/NameThatSong/internal/pkg/utils"
)

type SpotifyCache interface {
	// Basic entity cache operations (cache-only, no API calls)
	GetTrack(trackId string) (m.TrackData, bool)
	SetTrack(trackId string, track m.TrackData)
	GetMultipleTracks(trackIDs []string) (map[string]m.TrackData, []string)
	SetMultipleTracks(tracks map[string]m.TrackData)
	GetAlbum(albumId string) (m.AlbumData, bool)
	SetAlbum(albumId string, album m.AlbumData)
	GetMultipleAlbums(albumIDs []string) (map[string]m.AlbumData, []string)
	SetMultipleAlbums(albums map[string]m.AlbumData)
	GetArtist(artistId string) (m.ArtistData, bool)
	SetArtist(artistId string, artist m.ArtistData)
	GetMultipleArtists(artistIDs []string) (map[string]m.ArtistData, []string)
	SetMultipleArtists(artists map[string]m.ArtistData)
	GetPlaylist(playlistId string) (m.PlaylistData, bool)
	SetPlaylist(playlistId string, playlist m.PlaylistData)
	GetMultiplePlaylists(artistIDs []string) (map[string]m.PlaylistData, []string)
	SetMultiplePlaylists(artists map[string]m.PlaylistData)
	GetPlaylistTracks(playlistId string) ([]string, bool)
	SetPlaylistTracks(playlistId string, tracks []string)
	GetPlaylistAlbums(playlistId string) ([]string, bool)
	SetPlaylistAlbums(playlistId string, tracks []string)

	// Search cache operations (cache-only, no API calls)
	GetSearchTracks(query string) ([]m.TrackSearch, bool)
	SetSearchTracks(query string, tracks []m.TrackSearch)
	GetSearchAlbums(query string) ([]m.AlbumSearch, bool)
	SetSearchAlbums(query string, albums []m.AlbumSearch)
	GetSearchArtists(query string) ([]m.ArtistSearch, bool)
	SetSearchArtists(query string, artists []m.ArtistSearch)
	GetSearchPlaylists(query string) ([]m.PlaylistSearch, bool)
	SetSearchPlaylists(query string, playlists []m.PlaylistSearch)

	// OAuth state management
	GetOAuthState(userID string) (string, bool)
	SetOAuthState(userID, state string)
}

type RedisSpotifyCache struct {
	client *redis.Client
	ctx    context.Context
	ttl    time.Duration
}

func NewRedisSpotifyCache(redisClient *redis.Client) *RedisSpotifyCache {
	return &RedisSpotifyCache{
		client: redisClient,
		ctx:    context.Background(),
		ttl:    time.Hour, // Default 1 hour TTL
	}
}

func (r *RedisSpotifyCache) generateKey(entityType, entityId string) string {
	return fmt.Sprintf("spotify:%s:%s", entityType, entityId)
}

func (r *RedisSpotifyCache) generateSearchKey(searchType, query string) string {
	// Normalize and sanitize the query for consistent cache keys
	normalizedQuery := utils.NormalizeAndSanitizeQuery(query)
	return fmt.Sprintf("spotify:search:%s:%s", searchType, normalizedQuery)
}

// Track cache operations
func (r *RedisSpotifyCache) GetTrack(trackId string) (m.TrackData, bool) {
	var track m.TrackData
	key := r.generateKey("track", trackId)

	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return m.TrackData{}, false
	}

	if err := json.Unmarshal([]byte(val), &track); err != nil {
		return m.TrackData{}, false
	}

	return track, true
}

func (r *RedisSpotifyCache) SetTrack(trackId string, track m.TrackData) {
	key := r.generateKey("track", trackId)
	trackJson, err := json.Marshal(track)
	if err != nil {
		return
	}
	r.client.Set(r.ctx, key, trackJson, r.ttl)
}

// Batch track operations
func (r *RedisSpotifyCache) GetMultipleTracks(trackIDs []string) (map[string]m.TrackData, []string) {
	if len(trackIDs) == 0 {
		return make(map[string]m.TrackData), []string{}
	}

	// Use Redis pipeline for batch GET
	pipeline := r.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(trackIDs))
	for i, trackID := range trackIDs {
		key := r.generateKey("track", trackID)
		cmds[i] = pipeline.Get(r.ctx, key)
	}

	// build pipeline
	if _, err := pipeline.Exec(r.ctx); err != nil {
		logger.Error(r.ctx, "could not fetch batch all trackIDs")
	}

	// Check what is missing, and what worked
	found := make(map[string]m.TrackData)
	var missing []string
	for i, cmd := range cmds {
		trackID := trackIDs[i]

		// single command error:
		val, err := cmd.Result()
		if err != nil {
			missing = append(missing, trackID)
			continue
		}

		// unmarshal into TrackData
		var track m.TrackData
		if err := json.Unmarshal([]byte(val), &track); err != nil {
			missing = append(missing, trackID)
			continue
		}

		found[trackID] = track
	}

	return found, missing
}

func (r *RedisSpotifyCache) SetMultipleTracks(tracks map[string]m.TrackData) {
	if len(tracks) == 0 {
		return
	}

	// Use Redis pipeline for batch SET
	pipeline := r.client.Pipeline()
	skippedCount := 0
	for trackID, track := range tracks {
		key := r.generateKey("track", trackID)
		trackJson, err := json.Marshal(track)
		if err != nil {
			logger.Warn(r.ctx, "Failed to marshal track for cache", logger.F("trackID", trackID), logger.F("error", err))
			skippedCount++
			continue
		}
		pipeline.Set(r.ctx, key, trackJson, r.ttl)
	}

	if _, err := pipeline.Exec(r.ctx); err != nil {
		logger.Warn(r.ctx, "Redis batch cache operation failed for tracks", logger.F("attempted", len(tracks)-skippedCount), logger.F("skipped", skippedCount), logger.F("error", err))
	}
}

// Album cache operations
func (r *RedisSpotifyCache) GetAlbum(albumId string) (m.AlbumData, bool) {
	var album m.AlbumData
	key := r.generateKey("album", albumId)

	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return m.AlbumData{}, false
	}

	if err := json.Unmarshal([]byte(val), &album); err != nil {
		return m.AlbumData{}, false
	}

	return album, true
}

func (r *RedisSpotifyCache) SetAlbum(albumId string, album m.AlbumData) {
	key := r.generateKey("album", albumId)
	albumJson, err := json.Marshal(album)
	if err != nil {
		return
	}
	r.client.Set(r.ctx, key, albumJson, r.ttl)
}

// Batch album operations
func (r *RedisSpotifyCache) GetMultipleAlbums(albumIDs []string) (map[string]m.AlbumData, []string) {
	if len(albumIDs) == 0 {
		return make(map[string]m.AlbumData), []string{}
	}

	// Build pipeline Redis
	pipeline := r.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(albumIDs))
	for i, albumID := range albumIDs {
		key := r.generateKey("album", albumID)
		cmds[i] = pipeline.Get(r.ctx, key)
	}

	if _, err := pipeline.Exec(r.ctx); err != nil {
		logger.Error(r.ctx, "could not fetch batch all albumIDs")
	}

	// Check what is missing, and what worked
	found := make(map[string]m.AlbumData)
	var missing []string
	for i, cmd := range cmds {
		albumID := albumIDs[i]

		// single command error:
		val, err := cmd.Result()
		if err != nil {
			missing = append(missing, albumID)
			continue
		}

		// unmarshal into AlbumData
		var album m.AlbumData
		if err := json.Unmarshal([]byte(val), &album); err != nil {
			missing = append(missing, albumID)
			continue
		}

		found[albumID] = album
	}

	return found, missing
}

func (r *RedisSpotifyCache) SetMultipleAlbums(albums map[string]m.AlbumData) {
	if len(albums) == 0 {
		return
	}

	// Use Redis pipeline for batch SET
	pipeline := r.client.Pipeline()
	skippedCount := 0
	for albumID, album := range albums {
		key := r.generateKey("album", albumID)
		albumJson, err := json.Marshal(album)
		if err != nil {
			logger.Warn(r.ctx, "Failed to marshal album for cache", logger.F("albumID", albumID), logger.F("error", err))
			skippedCount++
			continue
		}
		pipeline.Set(r.ctx, key, albumJson, r.ttl)
	}

	if _, err := pipeline.Exec(r.ctx); err != nil {
		logger.Warn(r.ctx, "Redis batch cache operation failed for albums", logger.F("attempted", len(albums)-skippedCount), logger.F("skipped", skippedCount), logger.F("error", err))
	}
}

// Artist cache operations
func (r *RedisSpotifyCache) GetArtist(artistId string) (m.ArtistData, bool) {
	var artist m.ArtistData
	key := r.generateKey("artist", artistId)

	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return m.ArtistData{}, false
	}

	if err := json.Unmarshal([]byte(val), &artist); err != nil {
		return m.ArtistData{}, false
	}

	return artist, true
}

func (r *RedisSpotifyCache) SetArtist(artistId string, artist m.ArtistData) {
	key := r.generateKey("artist", artistId)
	artistJson, err := json.Marshal(artist)
	if err != nil {
		return
	}
	r.client.Set(r.ctx, key, artistJson, r.ttl)
}

// Batch artist operations
func (r *RedisSpotifyCache) GetMultipleArtists(artistIDs []string) (map[string]m.ArtistData, []string) {
	if len(artistIDs) == 0 {
		return make(map[string]m.ArtistData), []string{}
	}
	// Build pipeline Redis
	pipeline := r.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(artistIDs))
	for i, artistID := range artistIDs {
		key := r.generateKey("artist", artistID)
		cmds[i] = pipeline.Get(r.ctx, key)
	}

	// Execute
	if _, err := pipeline.Exec(r.ctx); err != nil {
		logger.Error(r.ctx, "could not fetch batch all artistIDs")
	}

	// Check what is missing, and what worked
	found := make(map[string]m.ArtistData)
	var missing []string
	for i, cmd := range cmds {
		artistID := artistIDs[i]

		// single command error:
		val, err := cmd.Result()
		if err != nil {
			missing = append(missing, artistID)
			continue
		}

		// unmarshal into ArtistData
		var artist m.ArtistData
		if err := json.Unmarshal([]byte(val), &artist); err != nil {
			missing = append(missing, artistID)
			continue
		}

		found[artistID] = artist
	}

	return found, missing
}

func (r *RedisSpotifyCache) SetMultipleArtists(artists map[string]m.ArtistData) {
	if len(artists) == 0 {
		return
	}
	// Use Redis pipeline for batch SET
	pipeline := r.client.Pipeline()
	skippedCount := 0
	for artistID, artist := range artists {
		key := r.generateKey("artist", artistID)
		artistJson, err := json.Marshal(artist)
		if err != nil {
			logger.Warn(r.ctx, "Failed to marshal artist for cache", logger.F("artistID", artistID), logger.F("error", err))
			skippedCount++
			continue
		}
		pipeline.Set(r.ctx, key, artistJson, r.ttl)
	}
	if _, err := pipeline.Exec(r.ctx); err != nil {
		logger.Warn(r.ctx, "Redis batch cache operation failed for artists", logger.F("attempted", len(artists)-skippedCount), logger.F("skipped", skippedCount), logger.F("error", err))
	}
}

// Playlist cache operations
func (r *RedisSpotifyCache) GetPlaylist(playlistId string) (m.PlaylistData, bool) {
	var playlist m.PlaylistData
	key := r.generateKey("playlist", playlistId)

	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return m.PlaylistData{}, false
	}

	if err := json.Unmarshal([]byte(val), &playlist); err != nil {
		return m.PlaylistData{}, false
	}
	return playlist, true
}

func (r *RedisSpotifyCache) SetPlaylist(playlistId string, playlist m.PlaylistData) {
	key := r.generateKey("playlist", playlistId)
	playlistJson, err := json.Marshal(playlist)
	if err != nil {
		return
	}
	r.client.Set(r.ctx, key, playlistJson, r.ttl)
}

// Batch playlist operations
func (r *RedisSpotifyCache) GetMultiplePlaylists(playlistIDs []string) (map[string]m.PlaylistData, []string) {
	if len(playlistIDs) == 0 {
		return make(map[string]m.PlaylistData), []string{}
	}
	// Build pipeline Redis
	pipeline := r.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(playlistIDs))
	for i, playlistID := range playlistIDs {
		key := r.generateKey("playlist", playlistID)
		cmds[i] = pipeline.Get(r.ctx, key)
	}

	// Execute pipeline
	if _, err := pipeline.Exec(r.ctx); err != nil {
		logger.Error(r.ctx, "could not fetch batch all playlistIDs")
	}

	// Check what is missing, and what worked
	found := make(map[string]m.PlaylistData)
	var missing []string
	for i, cmd := range cmds {
		playlistID := playlistIDs[i]
		// single command error:
		val, err := cmd.Result()
		if err != nil {
			missing = append(missing, playlistID)
			continue
		}
		// unmarshal into PlaylistData
		var playlist m.PlaylistData
		if err := json.Unmarshal([]byte(val), &playlist); err != nil {
			missing = append(missing, playlistID)
			continue
		}
		found[playlistID] = playlist
	}
	return found, missing
}

func (r *RedisSpotifyCache) SetMultiplePlaylists(playlists map[string]m.PlaylistData) {
	if len(playlists) == 0 {
		return
	}
	// Use Redis pipeline for batch SET
	pipeline := r.client.Pipeline()
	skippedCount := 0
	for playlistID, playlist := range playlists {
		key := r.generateKey("playlist", playlistID)
		playlistJson, err := json.Marshal(playlist)
		if err != nil {
			logger.Warn(r.ctx, "Failed to marshal playlist for cache", logger.F("playlistID", playlistID), logger.F("error", err))
			skippedCount++
			continue
		}
		pipeline.Set(r.ctx, key, playlistJson, r.ttl)
	}
	if _, err := pipeline.Exec(r.ctx); err != nil {
		logger.Warn(r.ctx, "Redis batch cache operation failed for playlists", logger.F("attempted", len(playlists)-skippedCount), logger.F("skipped", skippedCount), logger.F("error", err))
	}
}

func (r *RedisSpotifyCache) GetPlaylistTracks(playlistId string) ([]string, bool) {
	key := r.generateKey("playlist:trackIDs", playlistId)

	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return []string{}, false
	}

	var trackIDs []string
	if err := json.Unmarshal([]byte(val), &trackIDs); err != nil {
		return trackIDs, false
	}

	return trackIDs, true
}
func (r *RedisSpotifyCache) SetPlaylistTracks(playlistId string, tracks []string) {
	key := r.generateKey("playlist:trackIDs", playlistId)

	playlistJson, err := json.Marshal(tracks)
	if err != nil {
		return
	}
	r.client.Set(r.ctx, key, playlistJson, r.ttl)
}

func (r *RedisSpotifyCache) GetPlaylistAlbums(playlistId string) ([]string, bool) {
	key := r.generateKey("playlist:albumIDs", playlistId)

	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return []string{}, false
	}

	var albumIDs []string
	if err := json.Unmarshal([]byte(val), &albumIDs); err != nil {
		return albumIDs, false
	}

	return albumIDs, true
}
func (r *RedisSpotifyCache) SetPlaylistAlbums(playlistId string, albums []string) {
	key := r.generateKey("playlist:albumIDs", playlistId)

	playlistJson, err := json.Marshal(albums)
	if err != nil {
		return
	}
	r.client.Set(r.ctx, key, playlistJson, r.ttl)
}

// Search cache operations
func (r *RedisSpotifyCache) GetSearchTracks(query string) ([]m.TrackSearch, bool) {
	var tracks []m.TrackSearch
	key := r.generateSearchKey("tracks", query)

	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return nil, false
	}

	if err := json.Unmarshal([]byte(val), &tracks); err != nil {
		return nil, false
	}

	return tracks, true
}

func (r *RedisSpotifyCache) SetSearchTracks(query string, tracks []m.TrackSearch) {
	key := r.generateSearchKey("tracks", query)
	tracksJson, err := json.Marshal(tracks)
	if err != nil {
		return
	}
	r.client.Set(r.ctx, key, tracksJson, r.ttl)
}

func (r *RedisSpotifyCache) GetSearchAlbums(query string) ([]m.AlbumSearch, bool) {
	var albums []m.AlbumSearch
	key := r.generateSearchKey("albums", query)

	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return nil, false
	}

	if err := json.Unmarshal([]byte(val), &albums); err != nil {
		return nil, false
	}

	return albums, true
}

func (r *RedisSpotifyCache) SetSearchAlbums(query string, albums []m.AlbumSearch) {
	key := r.generateSearchKey("albums", query)
	albumsJson, err := json.Marshal(albums)
	if err != nil {
		return
	}
	r.client.Set(r.ctx, key, albumsJson, r.ttl)
}

func (r *RedisSpotifyCache) GetSearchArtists(query string) ([]m.ArtistSearch, bool) {
	var artists []m.ArtistSearch
	key := r.generateSearchKey("artists", query)

	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return nil, false
	}

	if err := json.Unmarshal([]byte(val), &artists); err != nil {
		return nil, false
	}

	return artists, true
}

func (r *RedisSpotifyCache) SetSearchArtists(query string, artists []m.ArtistSearch) {
	key := r.generateSearchKey("artists", query)
	artistsJson, err := json.Marshal(artists)
	if err != nil {
		return
	}
	r.client.Set(r.ctx, key, artistsJson, r.ttl)
}

func (r *RedisSpotifyCache) GetSearchPlaylists(query string) ([]m.PlaylistSearch, bool) {
	var playlists []m.PlaylistSearch
	key := r.generateSearchKey("playlists", query)

	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return nil, false
	}

	if err := json.Unmarshal([]byte(val), &playlists); err != nil {
		return nil, false
	}

	return playlists, true
}

func (r *RedisSpotifyCache) SetSearchPlaylists(query string, playlists []m.PlaylistSearch) {
	key := r.generateSearchKey("playlists", query)
	playlistsJson, err := json.Marshal(playlists)
	if err != nil {
		return
	}
	r.client.Set(r.ctx, key, playlistsJson, r.ttl)
}

// OAuth state management operations
func (r *RedisSpotifyCache) GetOAuthState(userID string) (string, bool) {
	key := r.generateStateKey(userID)

	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		return "", false
	}

	return val, true
}

func (r *RedisSpotifyCache) SetOAuthState(userID, state string) {
	key := r.generateStateKey(userID)
	// OAuth state should have short TTL (5 minutes) for security
	r.client.Set(r.ctx, key, state, 5*time.Minute)
}

func (r *RedisSpotifyCache) generateStateKey(userID string) string {
	return fmt.Sprintf("spotify:oauth:state:%s", userID)
}
