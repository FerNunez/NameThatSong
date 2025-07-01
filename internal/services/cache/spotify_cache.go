package cache

import (
	m "github.com/FerNunez/NameThatSong/internal/models"
)

type SpotifyCache interface {

	// kind of internal
	GetTrack(accessToken, trackId string) (m.TrackData, error)
	GetAlbum(accessToken, albumId string) (m.AlbumData, error)
	GetArtist(accessToken, artistId string) (m.ArtistData, error)
	GetPlaylist(accessToken, playlistId string) (m.PlaylistData, error)

	// Get realtions
	GetAlbumsFromArtist(accessToken, artistId string) ([]string, error)
	GetTracksFromAlbum(accessToken, albumId string) ([]string, error)
	GetTracksFromPlaylist(accessToken, playlistId string) ([]string, error)

	// Search
	SearchTracks(accessToken, query string) ([]m.TrackSearch, error)
	SearchAlbums(accessToken, query string) ([]m.AlbumSearch, error)
	SearchArtists(accessToken, query string) ([]m.ArtistSearch, error)
	SearchPlaylists(accessToken, query string) ([]m.PlaylistSearch, error)
}
