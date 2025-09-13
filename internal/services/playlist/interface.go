package playlist

import (
	"context"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/google/uuid"
)

// PlaylistService defines the contract for playlist service operations
type Service interface {
	// Local playlist operations
	CreatePlaylist(ctx context.Context, userID uuid.UUID, req models.CreatePlaylistRequest) (*models.LocalPlaylist, error)
	GetUserPlaylists(ctx context.Context, userID uuid.UUID) ([]*models.LocalPlaylist, error)
	GetPlaylist(ctx context.Context, playlistID, userID uuid.UUID) (*models.LocalPlaylist, error)
	UpdatePlaylist(ctx context.Context, playlistID, userID uuid.UUID, req models.UpdatePlaylistRequest) (*models.LocalPlaylist, error)
	DeletePlaylist(ctx context.Context, playlistID, userID uuid.UUID) error

	// Song management
	AddSongToPlaylist(ctx context.Context, userID string, playlistID uuid.UUID, req models.AddSongRequest) error
	RemoveSongFromPlaylist(ctx context.Context, playlistID, userID uuid.UUID, spotifyTrackID string) error
	//ReorderPlaylistSongs(ctx context.Context, playlistID, userID uuid.UUID, req models.ReorderSongsRequest) error
	GetPlaylistSongs(ctx context.Context, userID string, playlistID uuid.UUID) ([]*models.PlaylistTrack, error)
	GetPlaylistSongsWithDetails(ctx context.Context, userID string, playlistID uuid.UUID) ([]*models.PlaylistTrackWithDetails, error)

	// Spotify integration
	ImportFromSpotify(ctx context.Context, userID uuid.UUID, req models.ImportPlaylistRequest) (*models.LocalPlaylist, error)
	ExportToSpotify(ctx context.Context, userID uuid.UUID, req models.ExportPlaylistRequest) (string, error)
	//SyncWithSpotify(ctx context.Context, playlistID, userID uuid.UUID) error
	ImportUsersPlaylistsFromSpotify(ctx context.Context, userID uuid.UUID) ([]models.LocalPlaylist, error)
}
