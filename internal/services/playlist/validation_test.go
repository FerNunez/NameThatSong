package playlist

import (
	"testing"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestValidateCreatePlaylistRequest(t *testing.T) {
	p := &PlaylistProvider{} // Create instance to call validation methods

	tests := []struct {
		name        string
		request     models.CreatePlaylistRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid create request",
			request: models.CreatePlaylistRequest{
				Name:        "My Test Playlist",
				Description: "A great playlist for testing",
				IsPublic:    true,
			},
			expectError: false,
		},
		{
			name: "valid create request with empty description",
			request: models.CreatePlaylistRequest{
				Name:        "My Test Playlist",
				Description: "",
				IsPublic:    false,
			},
			expectError: false,
		},
		{
			name: "invalid create request - empty name",
			request: models.CreatePlaylistRequest{
				Name:        "",
				Description: "A great playlist",
				IsPublic:    true,
			},
			expectError: true,
			errorMsg:    "invalid name",
		},
		{
			name: "invalid create request - name too long",
			request: models.CreatePlaylistRequest{
				Name:        string(make([]byte, 101)), // 101 characters
				Description: "A great playlist",
				IsPublic:    true,
			},
			expectError: true,
			errorMsg:    "invalid name",
		},
		{
			name: "invalid create request - description too long",
			request: models.CreatePlaylistRequest{
				Name:        "Valid Name",
				Description: string(make([]byte, 501)), // 501 characters
				IsPublic:    true,
			},
			expectError: true,
			errorMsg:    "invalid description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.validateCreatePlaylistRequest(tt.request)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUpdatePlaylistRequest(t *testing.T) {
	p := &PlaylistProvider{}

	tests := []struct {
		name        string
		request     models.UpdatePlaylistRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid update request",
			request: models.UpdatePlaylistRequest{
				Name:        "Updated Playlist Name",
				Description: "Updated description",
				IsPublic:    false,
			},
			expectError: false,
		},
		{
			name: "invalid update request - empty name",
			request: models.UpdatePlaylistRequest{
				Name:        "",
				Description: "Updated description",
				IsPublic:    false,
			},
			expectError: true,
			errorMsg:    "invalid name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.validateUpdatePlaylistRequest(tt.request)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateImportPlaylistRequest(t *testing.T) {
	p := &PlaylistProvider{}

	tests := []struct {
		name        string
		request     models.ImportPlaylistRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid import request",
			request: models.ImportPlaylistRequest{
				SpotifyPlaylistID: "37i9dQZF1DXcBWIGoYBM5M",
			},
			expectError: false,
		},
		{
			name: "invalid import request - empty spotify ID",
			request: models.ImportPlaylistRequest{
				SpotifyPlaylistID: "",
			},
			expectError: true,
			errorMsg:    "invalid spotify playlist ID",
		},
		{
			name: "invalid import request - invalid spotify ID",
			request: models.ImportPlaylistRequest{
				SpotifyPlaylistID: "invalid-id!",
			},
			expectError: true,
			errorMsg:    "invalid spotify playlist ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.validateImportPlaylistRequest(tt.request)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateExportPlaylistRequest(t *testing.T) {
	p := &PlaylistProvider{}

	validUUID := uuid.New()

	tests := []struct {
		name        string
		request     models.ExportPlaylistRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid export request",
			request: models.ExportPlaylistRequest{
				PlaylistID: validUUID,
				Name:       "Exported Playlist",
				IsPublic:   true,
			},
			expectError: false,
		},
		{
			name: "invalid export request - nil playlist ID",
			request: models.ExportPlaylistRequest{
				PlaylistID: uuid.Nil,
				Name:       "Exported Playlist",
				IsPublic:   true,
			},
			expectError: true,
			errorMsg:    "playlist ID is required",
		},
		{
			name: "invalid export request - empty name",
			request: models.ExportPlaylistRequest{
				PlaylistID: validUUID,
				Name:       "",
				IsPublic:   true,
			},
			expectError: true,
			errorMsg:    "invalid export name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.validateExportPlaylistRequest(tt.request)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAddSongRequest(t *testing.T) {
	p := &PlaylistProvider{}

	tests := []struct {
		name        string
		request     models.AddSongRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid add song request",
			request: models.AddSongRequest{
				SpotifyTrackID: "4iV5W9uYEdYUVa79Axb7Rh",
			},
			expectError: false,
		},
		{
			name: "invalid add song request - empty track ID",
			request: models.AddSongRequest{
				SpotifyTrackID: "",
			},
			expectError: true,
			errorMsg:    "invalid spotify track ID",
		},
		{
			name: "invalid add song request - invalid track ID",
			request: models.AddSongRequest{
				SpotifyTrackID: "invalid-track-id!",
			},
			expectError: true,
			errorMsg:    "invalid spotify track ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.validateAddSongRequest(tt.request)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateReorderSongsRequest(t *testing.T) {
	p := &PlaylistProvider{}

	uuid1 := uuid.New()
	uuid2 := uuid.New()
	uuid3 := uuid.New()

	tests := []struct {
		name        string
		request     models.ReorderSongsRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid reorder request",
			request: models.ReorderSongsRequest{
				SongOrder: []uuid.UUID{uuid1, uuid2, uuid3},
			},
			expectError: false,
		},
		{
			name: "valid reorder request - single song",
			request: models.ReorderSongsRequest{
				SongOrder: []uuid.UUID{uuid1},
			},
			expectError: false,
		},
		{
			name: "invalid reorder request - empty order",
			request: models.ReorderSongsRequest{
				SongOrder: []uuid.UUID{},
			},
			expectError: true,
			errorMsg:    "song order cannot be empty",
		},
		{
			name: "invalid reorder request - nil UUID",
			request: models.ReorderSongsRequest{
				SongOrder: []uuid.UUID{uuid1, uuid.Nil, uuid3},
			},
			expectError: true,
			errorMsg:    "song ID at position 1 is invalid",
		},
		{
			name: "invalid reorder request - duplicate UUIDs",
			request: models.ReorderSongsRequest{
				SongOrder: []uuid.UUID{uuid1, uuid2, uuid1},
			},
			expectError: true,
			errorMsg:    "duplicate song ID found at position 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := p.validateReorderSongsRequest(tt.request)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
