package cache

import (
	"github.com/FerNunez/NameThatSong/internal/spotify_api"
)

type SpotifyCache interface {

	// kind of internal
	GetTrack(accessToken, trackId string) (spotify_api.TrackData, error)
	GetAlbum(accessToken, albumId string) (spotify_api.AlbumData, error)
	GetArtist(accessToken, artistId string) (spotify_api.ArtistData, error)
	GetPlaylist(accessToken, playlistId string) (spotify_api.PlaylistData, error)

	// Get realtions
	GetAlbumsFromArtist(accessToken, artistId string) ([]string, error)
	GetTracksFromAlbum(accessToken, albumId string) ([]string, error)
	GetTracksFromPlaylist(accessToken, playlistId string) ([]string, error)
}
