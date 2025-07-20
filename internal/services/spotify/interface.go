package spotify

import (
	"context"

	m "github.com/FerNunez/NameThatSong/internal/models"
)

// SpotifyServiceInterface defines the contract for Spotify service operations
type SpotifyServiceInterface interface {
	// Authentication operations
	AuthRequestURL(userID string) (string, error)
	TokenExchange(ctx context.Context, userID, code, receivedState string) (TokenResponse, error)
	GetValidToken(ctx context.Context, userID string) (string, error)

	// Search operations
	SearchTracks(ctx context.Context, userID, query string) ([]m.TrackSearch, error)
	SearchAlbums(ctx context.Context, userID, query string) ([]m.AlbumSearch, error)
	SearchArtists(ctx context.Context, userID, query string) ([]m.ArtistSearch, error)
	SearchPlaylists(ctx context.Context, userID, query string) ([]m.PlaylistSearch, error)

	// Fetch operations
	FetchTrack(ctx context.Context, userID, trackID string) (m.TrackData, error)
	FetchAlbum(ctx context.Context, userID, albumID string) (m.AlbumData, error)
	FetchArtist(ctx context.Context, userID, artistID string) (m.ArtistData, error)
	FetchPlaylist(ctx context.Context, userID, playlistID string) (m.PlaylistData, error)
	FetchAlbumsFromArtist(ctx context.Context, userID, artistID string) ([]string, error)
	FetchTracksFromAlbum(ctx context.Context, userID, albumID string) ([]string, error)
	FetchTracksFromPlaylist(ctx context.Context, userID, playlistID string) ([]string, error)

	// Player operations
	PlaySong(ctx context.Context, userID, songID string) error
	PausePlayback(ctx context.Context, userID string) error
	ResumePlayback(ctx context.Context, userID string) error
}