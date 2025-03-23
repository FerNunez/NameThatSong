package base

import (
	"fmt"
	"time"
)

type ApiCache struct {
	TracksCache  map[string]TrackCache
	AlbumsCache  map[string]AlbumCache
	ArtistsCache map[string]ArtistCache
}

func NewApiCache() ApiCache {
	return ApiCache{
		TracksCache:  map[string]TrackCache{},
		AlbumsCache:  map[string]AlbumCache{},
		ArtistsCache: map[string]ArtistCache{},
	}
}

type TrackCache struct {
	Track      Track
	LastUpdate time.Time
}
type AlbumCache struct {
	Album      Album
	LastUpdate time.Time
}
type ArtistCache struct {
	Artist     Artist
	LastUpdate time.Time
}

func (ac *AlbumCache) ToggleSelect() {
	ac.Album.Selected = !ac.Album.Selected
}

func (c *ApiCache) InsertAlbum(key string, album Album) {

	result, exists := c.AlbumsCache[key]
	if exists {
		fmt.Println("Album key already exists", key)
		result.LastUpdate = time.Now()
	} else {
		result.Album = album
	}
	c.AlbumsCache[key] = result

}

func (c *ApiCache) RetrieveAlbum(key string) (Album, error) {

	result, exists := c.AlbumsCache[key]
	if !exists {
		return Album{}, fmt.Errorf("album key doesnt exists")
	}

	return result.Album, nil
}

func (c *ApiCache) InsertTrack(key string, track Track) {

	result, exists := c.TracksCache[key]
	if exists {
		fmt.Println("track key already exists for key:", key)
		result.LastUpdate = time.Now()
	} else {
		result.Track = track
	}
	c.TracksCache[key] = result

}

func (c *ApiCache) RetrieveTrack(key string) (Track, error) {

	result, exists := c.TracksCache[key]
	if !exists {
		return Track{}, fmt.Errorf("track key %v doesnt exists", key)
	}

	return result.Track, nil
}

func (c *ApiCache) InsertArtist(key string, album Artist) {

	result, exists := c.ArtistsCache[key]
	if exists {
		fmt.Println("artist key already exists for key:", key)
		result.LastUpdate = time.Now()
	} else {
		result.Artist = album
	}
	c.ArtistsCache[key] = result

}

func (c *ApiCache) RetrieveArtist(key string) (Artist, error) {

	result, exists := c.ArtistsCache[key]
	if !exists {
		return Artist{}, fmt.Errorf("track key %v doesnt exists", key)
	}

	return result.Artist, nil
}
