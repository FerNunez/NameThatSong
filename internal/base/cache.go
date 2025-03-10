package base

import "time"

type ApiCache struct {
	TracksCache map[string]TrackCache
	AlbumsCache map[string]AlbumCache
	ArtistCache map[string]ArtistCache
}

func NewApiCache() ApiCache {
	return ApiCache{
		TracksCache: map[string]TrackCache{},
		AlbumsCache: map[string]AlbumCache{},
		ArtistCache: map[string]ArtistCache{},
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
