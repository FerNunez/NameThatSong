package cache

import (
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
)

type SpotifyCache interface {

	// kind of internal
	GetTrack(accessToken, trackId string) (spotify.TrackData, error)
	GetAlbum(accessToken, albumId string) (spotify.AlbumData, error)
	GetArtist(accessToken, artistId string) (spotify.ArtistData, error)
	GetPlaylist(accessToken, playlistId string) (spotify.PlaylistData, error)

	// Get realtions
	GetAlbumsFromArtist(accessToken, artistId string) ([]string, error)
	GetTracksFromAlbum(accessToken, albumId string) ([]string, error)
	GetTracksFromPlaylist(accessToken, playlistId string) ([]string, error)

	// Search
	SearchTracks(accessToken, query string) ([]spotify.TrackSearch, error)
	SearchAlbums(accessToken, query string) ([]spotify.AlbumSearch, error)
	SearchArtists(accessToken, query string) ([]spotify.ArtistSearch, error)
	SearchPlaylists(accessToken, query string) ([]spotify.PlaylistSearch, error)
}
