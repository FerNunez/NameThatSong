package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	m "github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
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

func (r *RedisSpotifyCache) generateSearchKey(searchType, query string) string {
	return fmt.Sprintf("spotify:search:%s:%s", searchType, query)
}

func (r *RedisSpotifyCache) GetTrack(accessToken, trackId string) (m.TrackData, error) {
	var track m.TrackData
	key := r.generateKey("track", trackId)

	val, err := r.client.Get(r.ctx, key).Result()
	if err == nil {
		// cache hit
		if err := json.Unmarshal([]byte(val), &track); err != nil {
			return m.TrackData{}, fmt.Errorf("failed to unmarshal cached track: %v", err)
		}
		return track, nil
	}

	// Something went wrong: network issues, Redis down, auth problems, etc.
	if err != redis.Nil {
		return m.TrackData{}, fmt.Errorf("redis error: %v", err)
	}

	fmt.Println("[RedisSpotifyCache] GetTrack miss chache trackId:", trackId)
	// Cache miss
	track, err = spotify.FetchTrack(accessToken, trackId)
	if err != nil {
		return m.TrackData{}, fmt.Errorf("failed to fetch track data from SpotifyApi: %v", err)
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
func (r *RedisSpotifyCache) GetAlbum(accessToken, albumId string) (m.AlbumData, error) {
	var album m.AlbumData
	key := r.generateKey("album", albumId)

	val, err := r.client.Get(r.ctx, key).Result()
	if err == nil {
		// Cache hit
		if err := json.Unmarshal([]byte(val), &album); err != nil {
			return m.AlbumData{}, fmt.Errorf("failed to unmarshal cached album: %v", err)
		}
		return album, nil
	}

	if err != redis.Nil {
		return m.AlbumData{}, fmt.Errorf("redis error: %v", err)
	}

	fmt.Println("[RedisSpotifyCache] GetAlbum miss chache albumId:", albumId)
	// Cache miss
	album, _, err = spotify.FetchAlbum(accessToken, albumId)
	if err != nil {
		return m.AlbumData{}, fmt.Errorf("failed to fetch album data from SpotifyApi: %v", err)
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
func (r *RedisSpotifyCache) GetArtist(accessToken, artistId string) (m.ArtistData, error) {
	var artist m.ArtistData
	key := r.generateKey("artist", artistId)

	val, err := r.client.Get(r.ctx, key).Result()
	if err == nil {
		// Cache hit
		if err := json.Unmarshal([]byte(val), &artist); err != nil {
			return m.ArtistData{}, fmt.Errorf("failed to unmarshal cached artist: %v", err)
		}
		return artist, nil
	}

	if err != redis.Nil {
		return m.ArtistData{}, fmt.Errorf("redis error: %v", err)
	}

	fmt.Println("[RedisSpotifyCache] GetArtist miss chache artistId:", artistId)
	// Cache miss
	artist, err = spotify.FetchArtist(accessToken, artistId)
	if err != nil {
		return m.ArtistData{}, fmt.Errorf("failed to fetch artist data from SpotifyApi: %v", err)
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
func (r *RedisSpotifyCache) GetPlaylist(accessToken, playlistId string) (m.PlaylistData, error) {
	key := r.generateKey("playlist", playlistId)

	val, err := r.client.Get(r.ctx, key).Result()
	if err == nil {
		// Cache hit
		var playlist m.PlaylistData
		if err := json.Unmarshal([]byte(val), &playlist); err != nil {
			return m.PlaylistData{}, fmt.Errorf("failed to unmarshal cached playlist: %v", err)
		}
		return playlist, nil
	}

	if err != redis.Nil {
		return m.PlaylistData{}, fmt.Errorf("redis error: %v", err)
	}
	// Cache miss
	playlist, _, err := spotify.FetchPlaylist(accessToken, playlistId)
	if err != nil {
		return m.PlaylistData{}, fmt.Errorf("failed to fetch playlist data from SpotifyApi: %v", err)
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
	albumList, err := spotify.FetchAlbumsFromArtist(accessToken, artistId)
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
	trackList, err := spotify.FetchTracksFromAlbum(accessToken, albumId)
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
	trackList, err := spotify.FetchTracksFromPlaylist(accessToken, playlistId)
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

func (r *RedisSpotifyCache) SearchTracks(accessToken, query string) ([]m.TrackSearch, error) {
	key := r.generateSearchKey("tracks", query)

	val, err := r.client.Get(r.ctx, key).Result()
	if err == nil {
		// cache hit

		var tracks []m.TrackSearch
		if err := json.Unmarshal([]byte(val), &tracks); err != nil {
			return []m.TrackSearch{}, fmt.Errorf("could not unmarshal tracks for query: %v, err %v", query, err)
		}
		return tracks, nil
	}

	if err != redis.Nil {
		return []m.TrackSearch{}, fmt.Errorf("could not get tracks for query: %v, err %v", query, err)
	}

	// cache miss
	trackList, err := spotify.SearchTracksByName(accessToken, query)
	if err != nil {
		return []m.TrackSearch{}, fmt.Errorf("could not search tracks: %v, err %v", query, err)
	}
	// Store cache
	tracklistJson, err := json.Marshal(trackList)
	if err != nil {
		return []m.TrackSearch{}, fmt.Errorf("could not marshal searched tracks for: %v, err %v", query, err)
	}
	if err := r.client.Set(r.ctx, key, tracklistJson, r.ttl).Err(); err != nil {
		fmt.Printf("could not store searched tracks: %v, err %v", query, err)
	}
	return trackList, nil
}
func (r *RedisSpotifyCache) SearchAlbums(accessToken, query string) ([]m.AlbumSearch, error) {
	key := r.generateSearchKey("albums", query)

	val, err := r.client.Get(r.ctx, key).Result()
	if err == nil {
		// cache hit

		var albums []m.AlbumSearch
		if err := json.Unmarshal([]byte(val), &albums); err != nil {
			return []m.AlbumSearch{}, fmt.Errorf("could not unmarshal albums for query: %v, err %v", query, err)
		}
		return albums, nil
	}

	if err != redis.Nil {
		return []m.AlbumSearch{}, fmt.Errorf("could not get albums for query: %v, err %v", query, err)
	}

	// cache miss
	albumList, err := spotify.SearchAlbumsByName(accessToken, query)
	if err != nil {
		return []m.AlbumSearch{}, fmt.Errorf("could not search albums: %v, err %v", query, err)
	}
	// Store cache
	albumlistJson, err := json.Marshal(albumList)
	if err != nil {
		return []m.AlbumSearch{}, fmt.Errorf("could not marshal searched albums for: %v, err %v", query, err)
	}
	if err := r.client.Set(r.ctx, key, albumlistJson, r.ttl).Err(); err != nil {
		fmt.Printf("could not store searched albums: %v, err %v", query, err)
	}
	return albumList, nil
}
func (r *RedisSpotifyCache) SearchArtists(accessToken, query string) ([]m.ArtistSearch, error) {
	key := r.generateSearchKey("artists", query)

	val, err := r.client.Get(r.ctx, key).Result()
	if err == nil {
		// cache hit

		var artists []m.ArtistSearch
		if err := json.Unmarshal([]byte(val), &artists); err != nil {
			return []m.ArtistSearch{}, fmt.Errorf("could not unmarshal artists for query: %v, err %v", query, err)
		}
		return artists, nil
	}

	if err != redis.Nil {
		return []m.ArtistSearch{}, fmt.Errorf("could not get artists for query: %v, err %v", query, err)
	}

	// cache miss
	artistList, err := spotify.SearchArtistsByName(accessToken, query)
	if err != nil {
		return []m.ArtistSearch{}, fmt.Errorf("could not search artists: %v, err %v", query, err)
	}
	// Store cache
	artistlistJson, err := json.Marshal(artistList)
	if err != nil {
		return []m.ArtistSearch{}, fmt.Errorf("could not marshal searched artists for: %v, err %v", query, err)
	}
	if err := r.client.Set(r.ctx, key, artistlistJson, r.ttl).Err(); err != nil {
		fmt.Printf("could not store searched artists: %v, err %v", query, err)
	}
	return artistList, nil
}

func (r *RedisSpotifyCache) SearchPlaylists(accessToken, query string) ([]m.PlaylistSearch, error) {
	key := r.generateSearchKey("playlists", query)

	val, err := r.client.Get(r.ctx, key).Result()
	if err == nil {
		// cache hit

		var playlists []m.PlaylistSearch
		if err := json.Unmarshal([]byte(val), &playlists); err != nil {
			return []m.PlaylistSearch{}, fmt.Errorf("could not unmarshal playlists for query: %v, err %v", query, err)
		}
		return playlists, nil
	}

	if err != redis.Nil {
		return []m.PlaylistSearch{}, fmt.Errorf("could not get playlists for query: %v, err %v", query, err)
	}

	// cache miss
	playlistList, err := spotify.SearchPlaylistsByName(accessToken, query)
	if err != nil {
		return []m.PlaylistSearch{}, fmt.Errorf("could not search playlists: %v, err %v", query, err)
	}
	// Store cache
	playlistlistJson, err := json.Marshal(playlistList)
	if err != nil {
		return []m.PlaylistSearch{}, fmt.Errorf("could not marshal searched playlists for: %v, err %v", query, err)
	}
	if err := r.client.Set(r.ctx, key, playlistlistJson, r.ttl).Err(); err != nil {
		fmt.Printf("could not store searched playlists: %v, err %v", query, err)
	}
	return playlistList, nil
}
