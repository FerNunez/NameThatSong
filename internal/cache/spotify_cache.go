package cache

import (
	"github.com/FerNunez/NameThatSong/internal/spotify_api"
)

type SpotifyCache interface {
	GetArtistData(id string) (spotify_api.ArtistData, error)
	GetArtistsAlbum(accessToken, artistId string) ([]spotify_api.AlbumData, error)
	GetAlbumTracks(accessToken, albumId string) ([]spotify_api.TrackData, error)
}
