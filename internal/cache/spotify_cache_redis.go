package cache

import (
	"github.com/FerNunez/NameThatSong/internal/spotify_api"
	"github.com/redis/go-redis/v9"
)

type SpotifyCacheRedis struct {
	client *redis.Client
}

func (scr *SpotifyCacheRedis) GetArtistData(s *spotify_api.SpotifySongProvider, id string) (spotify_api.ArtistData, error) {
	return spotify_api.ArtistData{}, nil
}

func (scr *SpotifyCacheRedis) GetArtistsAlbum(s *spotify_api.SpotifySongProvider, accessToken, artistId string) ([]spotify_api.AlbumData, error) {
	return []spotify_api.AlbumData{}, nil
}

func (scr *SpotifyCacheRedis) GetAlbumTracks(s *spotify_api.SpotifySongProvider, accessToken, albumId string) ([]spotify_api.TrackData, error) {
	return []spotify_api.TrackData{}, nil
}
