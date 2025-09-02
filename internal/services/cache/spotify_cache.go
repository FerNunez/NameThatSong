package cache

import (
	m "github.com/FerNunez/NameThatSong/internal/models"
)

type SpotifyCache interface {
	// Basic entity cache operations (cache-only, no API calls)
	GetTrack(trackId string) (m.TrackData, bool)
	SetTrack(trackId string, track m.TrackData)
	GetMultipleTracks(trackIDs []string) (map[string]m.TrackData, []string)
	SetMultipleTracks(tracks map[string]m.TrackData)
	GetAlbum(albumId string) (m.AlbumData, bool)
	SetAlbum(albumId string, album m.AlbumData)
	GetArtist(artistId string) (m.ArtistData, bool)
	SetArtist(artistId string, artist m.ArtistData)
	GetPlaylist(playlistId string) (m.PlaylistData, bool)
	SetPlaylist(playlistId string, playlist m.PlaylistData)
	GetPlaylistTracks(playlistId string) ([]string, bool)
	SetPlaylistTracks(playlistId string, tracks []string)
	GetPlaylistAlbums(playlistId string) ([]string, bool)
	SetPlaylistAlbums(playlistId string, tracks []string)

	// Search cache operations (cache-only, no API calls)
	GetSearchTracks(query string) ([]m.TrackSearch, bool)
	SetSearchTracks(query string, tracks []m.TrackSearch)
	GetSearchAlbums(query string) ([]m.AlbumSearch, bool)
	SetSearchAlbums(query string, albums []m.AlbumSearch)
	GetSearchArtists(query string) ([]m.ArtistSearch, bool)
	SetSearchArtists(query string, artists []m.ArtistSearch)
	GetSearchPlaylists(query string) ([]m.PlaylistSearch, bool)
	SetSearchPlaylists(query string, playlists []m.PlaylistSearch)

	// OAuth state management
	GetOAuthState(userID string) (string, bool)
	SetOAuthState(userID, state string)
}
