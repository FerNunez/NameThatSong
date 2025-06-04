package cache

import (
	"github.com/FerNunez/NameThatSong/internal/spotify_api"
)

type SpotifyCache interface {
	GetArtistsByName(s *spotify_api.SpotifySongProvider, artistName string) ([]spotify_api.ArtistData, error)
	GetAlbumsIdFromArtistId(s *spotify_api.SpotifySongProvider, artistId string) ([]string, error)

	GetArtistData(s *spotify_api.SpotifySongProvider, id string) (spotify_api.ArtistData, error)
	GetArtistsAlbum(s *spotify_api.SpotifySongProvider, accessToken, artistId string) ([]spotify_api.AlbumData, error)
	GetAlbumTracks(s *spotify_api.SpotifySongProvider, accessToken, albumId string) ([]spotify_api.TrackData, error)

	// kind of internal
	GetTrack(s *spotify_api.SpotifySongProvider, accessToken, trackId string) (spotify_api.TrackData, error)
	GetAlbum(s *spotify_api.SpotifySongProvider, accessToken, albumId string) (spotify_api.AlbumData, error)
	GetArtist(s *spotify_api.SpotifySongProvider, accessToken, artistId string) (spotify_api.ArtistData, error)
}
