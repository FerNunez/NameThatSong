package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/FerNunez/NameThatSong/internal/spotify_api"
	"github.com/redis/go-redis/v9"
)

type RedisSpotifyCache struct {
	client *redis.Client
	ctx    context.Context
	ttl    time.Duration
}

func NewRedisSpotifyCache(redisAddr, password string, db int, ttl time.Duration) *RedisSpotifyCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr, //"localhost:6379",
		Password: password,  // no password set
		DB:       db,        // use default DB
	})

	return &RedisSpotifyCache{
		client: rdb,
		ctx:    context.Background(),
		ttl:    ttl,
	}
}

func (r *RedisSpotifyCache) generateKey(entityType, entityId string) string {
	return fmt.Sprintf("spotify:%s:%s", entityType, entityId)
}

func (r *RedisSpotifyCache) generateRelationalKey(parentType, parentId, childType string) string {
	return fmt.Sprintf("spotify:rel:%s:%s:%s", parentType, parentId, childType)
}

func (r *RedisSpotifyCache) GetTrack(accessToken, trackId string) (spotify_api.TrackData, error) {
	var track spotify_api.TrackData
	key := r.generateKey("track", trackId)

	val, err := r.client.Get(r.ctx, key).Result()
	if err == nil {
		// cache hit
		if err := json.Unmarshal([]byte(val), &track); err != nil {
			return spotify_api.TrackData{}, fmt.Errorf("failed to unmarshal cached track: %v", err)
		}
		return track, nil
	}

	// Something went wrong: network issues, Redis down, auth problems, etc.
	if err != redis.Nil {
		return spotify_api.TrackData{}, fmt.Errorf("redis error: %v", err)
	}

	fmt.Println("[RedisSpotifyCache] GetTrack miss chache trackId:", trackId)
	// Cache miss
	track, err = spotify_api.FetchTrack(accessToken, trackId)
	if err != nil {
		return spotify_api.TrackData{}, fmt.Errorf("failed to fetch track data from SpotifyApi: %v", err)
	}

	// Store in cache
	trackJson, err := json.Marshal(track)
	if err != nil {
		return track, fmt.Errorf("failed to marshal track for caching: %v", err)
	}
	if err := r.client.Set(r.ctx, key, trackJson, r.ttl).Err(); err != nil {
		fmt.Printf("Failed to cache track %v: %v", trackId, err)
	}
	return track, nil
}
func (r *RedisSpotifyCache) GetAlbum(accessToken, albumId string) (spotify_api.AlbumData, error) {
	var album spotify_api.AlbumData
	key := r.generateKey("album", albumId)

	val, err := r.client.Get(r.ctx, key).Result()
	if err == nil {
		// Cache hit
		if err := json.Unmarshal([]byte(val), &album); err != nil {
			return spotify_api.AlbumData{}, fmt.Errorf("failed to unmarshal cached album: %v", err)
		}
		return album, nil
	}

	if err != redis.Nil {
		return spotify_api.AlbumData{}, fmt.Errorf("redis error: %v", err)
	}

	fmt.Println("[RedisSpotifyCache] GetAlbum miss chache albumId:", albumId)
	// Cache miss
	album, _, err = spotify_api.FetchAlbum(accessToken, albumId)
	if err != nil {
		return spotify_api.AlbumData{}, fmt.Errorf("failed to fetch album data from SpotifyApi: %v", err)
	}

	// Store in cache
	albumJson, err := json.Marshal(album)
	if err != nil {
		return album, fmt.Errorf("failed to marshal album for caching: %v", err)
	}
	if err := r.client.Set(r.ctx, key, albumJson, r.ttl).Err(); err != nil {
		fmt.Printf("Failed to cache album %v: %v", albumId, err)
	}
	// TODO: Add tracks to check
	return album, nil
}
func (r *RedisSpotifyCache) GetArtist(accessToken, artistId string) (spotify_api.ArtistData, error) {
	var artist spotify_api.ArtistData
	key := r.generateKey("artist", artistId)

	val, err := r.client.Get(r.ctx, key).Result()
	if err == nil {
		// Cache hit
		if err := json.Unmarshal([]byte(val), &artist); err != nil {
			return spotify_api.ArtistData{}, fmt.Errorf("failed to unmarshal cached artist: %v", err)
		}
		return artist, nil
	}

	if err != redis.Nil {
		return spotify_api.ArtistData{}, fmt.Errorf("redis error: %v", err)
	}

	fmt.Println("[RedisSpotifyCache] GetArtist miss chache artistId:", artistId)
	// Cache miss
	artist, err = spotify_api.FetchArtist(accessToken, artistId)
	if err != nil {
		return spotify_api.ArtistData{}, fmt.Errorf("failed to fetch artist data from SpotifyApi: %v", err)
	}

	// Store in cache
	artistJson, err := json.Marshal(artist)
	if err != nil {
		return artist, fmt.Errorf("failed to marshal artist for caching: %v", err)
	}
	if err := r.client.Set(r.ctx, key, artistJson, r.ttl).Err(); err != nil {
		fmt.Printf("Failed to cache artist %v: %v", artistId, err)
	}
	return artist, nil
}
func (r *RedisSpotifyCache) GetPlaylist(accessToken, playlistId string) (spotify_api.PlaylistData, error) {
	key := r.generateKey("playlist", playlistId)

	val, err := r.client.Get(r.ctx, key).Result()
	if err == nil {
		// Cache hit
		var playlist spotify_api.PlaylistData
		if err := json.Unmarshal([]byte(val), &playlist); err != nil {
			return spotify_api.PlaylistData{}, fmt.Errorf("failed to unmarshal cached playlist: %v", err)
		}
		return playlist, nil
	}

	if err != redis.Nil {
		return spotify_api.PlaylistData{}, fmt.Errorf("redis error: %v", err)
	}
	// Cache miss
	playlist, _, err := spotify_api.FetchPlaylist(accessToken, playlistId)
	if err != nil {
		return spotify_api.PlaylistData{}, fmt.Errorf("failed to fetch playlist data from SpotifyApi: %v", err)
	}

	// Store in cache
	playlistJson, err := json.Marshal(playlist)
	if err != nil {
		return playlist, fmt.Errorf("failed to marshal playlist for caching: %v", err)
	}
	if err := r.client.Set(r.ctx, key, playlistJson, r.ttl).Err(); err != nil {
		fmt.Printf("Failed to cache playlist %v: %v", playlistId, err)
	}
	return playlist, nil
}
func (r *RedisSpotifyCache) GetAlbumsFromArtist(accessToken, artistId string) ([]string, error) {
	key := r.generateRelationalKey("artist", artistId, "albums")

	val, err := r.client.Get(r.ctx, key).Result()
	if err == nil {
		var albums []string
		// Cache hit
		if err := json.Unmarshal([]byte(val), &albums); err != nil {
			return []string{}, fmt.Errorf("failed to unmarshal cached Albums from Artist: %v", err)
		}
		return albums, nil
	}

	if err != redis.Nil {
		return []string{}, fmt.Errorf("redis error: %v", err)
	}
	// Cache miss
	albumList, err := spotify_api.FetchAlbumsFromArtist(accessToken, artistId)
	if err != nil {
		return []string{}, fmt.Errorf("failed to cache albums for artist %v: %v", artistId, err)
	}

	// Store cache
	albumListJson, err := json.Marshal(albumList)
	if err != nil {
		return []string{}, fmt.Errorf("failed to marshal albums: %v", err)
	}
	if err := r.client.Set(r.ctx, key, albumListJson, r.ttl).Err(); err != nil {
		fmt.Printf("Failed to cache rel albums for artistId %v: %v", artistId, err)
	}
	return albumList, nil
}
func (r *RedisSpotifyCache) GetTracksFromAlbum(accessToken, albumId string) ([]string, error) {
	key := r.generateRelationalKey("album", albumId, "tracks")

	val, err := r.client.Get(r.ctx, key).Result()
	if err == nil {
		var tracks []string
		// Cache hit
		if err := json.Unmarshal([]byte(val), &tracks); err != nil {
			return []string{}, fmt.Errorf("failed to unmarshal cached tracks from AlbumId %v, err: %v", albumId, err)
		}
		return tracks, nil
	}

	if err != redis.Nil {
		return []string{}, fmt.Errorf("redis error: %v", err)
	}
	// Cache miss
	trackList, err := spotify_api.FetchTracksFromAlbum(accessToken, albumId)
	if err != nil {
		return []string{}, fmt.Errorf("failed to cache tracks for album %v: %v", albumId, err)
	}

	// Store cache
	trackListJson, err := json.Marshal(trackList)
	if err != nil {
		return []string{}, fmt.Errorf("failed to marshal tracks: %v", err)
	}
	if err := r.client.Set(r.ctx, key, trackListJson, r.ttl).Err(); err != nil {
		fmt.Printf("Failed to cache rel tracks for albumId %v: %v", albumId, err)
	}
	return trackList, nil
}
func (r *RedisSpotifyCache) GetTracksFromPlaylist(accessToken, playlistId string) ([]string, error) {
	key := r.generateRelationalKey("playlist", playlistId, "tracks")

	val, err := r.client.Get(r.ctx, key).Result()
	if err == nil {
		var tracks []string
		// Cache hit
		if err := json.Unmarshal([]byte(val), &tracks); err != nil {
			return []string{}, fmt.Errorf("failed to unmarshal cached tracks from PlaylistId %v, err: %v", playlistId, err)
		}
		return tracks, nil
	}

	if err != redis.Nil {
		return []string{}, fmt.Errorf("redis error: %v", err)
	}
	// Cache miss
	trackList, err := spotify_api.FetchTracksFromPlaylist(accessToken, playlistId)
	if err != nil {
		return []string{}, fmt.Errorf("failed to cache tracks for playlist %v: %v", playlistId, err)
	}

	// Store cache
	trackListJson, err := json.Marshal(trackList)
	if err != nil {
		return []string{}, fmt.Errorf("failed to marshal tracks: %v", err)
	}
	if err := r.client.Set(r.ctx, key, trackListJson, r.ttl).Err(); err != nil {
		fmt.Printf("Failed to cache rel tracks for playlistId %v: %v", playlistId, err)
	}
	return trackList, nil
}
