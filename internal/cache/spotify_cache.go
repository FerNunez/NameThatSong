package cache

import (
	"goth/internal/spotify_api"
)

type SpotifyCache struct {
	ArtistToAlbumsMap map[string][]string
	AlbumMap          map[string]spotify_api.AlbumData
	AlbumToTracksMap  map[string][]string
	TrackMap          map[string]spotify_api.TrackData
}

func (c *SpotifyCache) GetArtistData(id string) (spotify_api.ArtistData, error) {

	return spotify_api.ArtistData{}, nil
}
