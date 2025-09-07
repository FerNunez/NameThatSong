package spotify

import (
	"context"

	m "github.com/FerNunez/NameThatSong/internal/models"
)

// SpotifyService defines the contract for Spotify service operations
type SpotifyService interface {
	// Authentication operations
	AuthRequestURL(userID string) (string, error)
	TokenExchange(ctx context.Context, userID, code, receivedState string) (TokenResponse, error)
	GetValidToken(ctx context.Context, userID string) (string, error)

	// Search operations
	SearchTracks(ctx context.Context, userID, query string) ([]m.TrackSearch, error)
	SearchAlbums(ctx context.Context, userID, query string) ([]m.AlbumSearch, error)
	SearchArtists(ctx context.Context, userID, query string) ([]m.ArtistSearch, error)
	SearchPlaylists(ctx context.Context, userID, query string) ([]m.PlaylistSearch, error)
	SearchAll(ctx context.Context, userID, query string) (*m.SearchAllResults, error)

	// Fetch operations
	FetchTrack(ctx context.Context, userID string, trackID m.SpotifyID) (m.TrackData, error)
	FetchMultipleTracks(ctx context.Context, userID string, trackIDs []m.SpotifyID) ([]m.TrackData, error)
	FetchAlbum(ctx context.Context, userID string, albumID m.SpotifyID) (m.AlbumData, error)
	FetchMultipleAlbums(ctx context.Context, userID string, albumIDs []m.SpotifyID) ([]m.AlbumData, error)
	FetchArtist(ctx context.Context, userID string, artistID m.SpotifyID) (m.ArtistData, error)
	FetchMultipleArtists(ctx context.Context, userID string, artistIDs []m.SpotifyID) ([]m.ArtistData, error)
	FetchPlaylist(ctx context.Context, userID string, playlistID m.SpotifyID) (m.PlaylistData, error)

	// Player operations
	PlaySong(ctx context.Context, userID, songID string) error
	PausePlayback(ctx context.Context, userID string) error
	ResumePlayback(ctx context.Context, userID string) error

	// Playlist operations
	GetUserPlaylists(ctx context.Context, userID string) ([]m.PlaylistData, error)
	CreatePlaylist(ctx context.Context, userID, name, description string, isPublic bool) (m.PlaylistData, error)
	AddTracksToPlaylist(ctx context.Context, userID, playlistID string, trackIDs []string) error
	RemoveTracksFromPlaylist(ctx context.Context, userID, playlistID string, trackIDs []string) error
}
