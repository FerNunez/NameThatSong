package playlist

import (
	"context"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/google/uuid"
)

// PlaylistService defines the contract for playlist service operations
type PlaylistService interface {
	// Local playlist operations
	CreatePlaylist(ctx context.Context, userID uuid.UUID, req models.CreatePlaylistRequest) (*models.Playlist, error)
	GetUserPlaylists(ctx context.Context, userID uuid.UUID) ([]*models.Playlist, error)
	GetPlaylist(ctx context.Context, playlistID, userID uuid.UUID) (*models.Playlist, error)
	UpdatePlaylist(ctx context.Context, playlistID, userID uuid.UUID, req models.UpdatePlaylistRequest) (*models.Playlist, error)
	DeletePlaylist(ctx context.Context, playlistID, userID uuid.UUID) error
	
	// Song management
	AddSongToPlaylist(ctx context.Context, playlistID, userID uuid.UUID, req models.AddSongRequest) error
	RemoveSongFromPlaylist(ctx context.Context, playlistID, userID uuid.UUID, songID uuid.UUID) error
	ReorderPlaylistSongs(ctx context.Context, playlistID, userID uuid.UUID, req models.ReorderSongsRequest) error
	GetPlaylistWithSongs(ctx context.Context, playlistID, userID uuid.UUID) (*models.Playlist, error)
	
	// Spotify integration
	ImportFromSpotify(ctx context.Context, userID uuid.UUID, req models.ImportPlaylistRequest) (*models.Playlist, error)
	ExportToSpotify(ctx context.Context, userID uuid.UUID, req models.ExportPlaylistRequest) (string, error)
	SyncWithSpotify(ctx context.Context, playlistID, userID uuid.UUID) error
}