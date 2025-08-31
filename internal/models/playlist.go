package models

import (
	"time"

	"github.com/google/uuid"
)

// Business domain models following existing patterns
type Playlist struct {
	ID                uuid.UUID  `json:"id"`
	UserID            uuid.UUID  `json:"user_id"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	SpotifyPlaylistID *string    `json:"spotify_playlist_id"`
	IsPublic          bool       `json:"is_public"`
	SyncWithSpotify   bool       `json:"sync_with_spotify"`
	LastSyncedAt      *time.Time `json:"last_synced_at"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	Songs             []Song     `json:"songs,omitempty"`
}

type Song struct {
	SpotifyTrackID   string    `json:"spotify_track_id"`
	SpotifyAlbumID   string    `json:"spotify_album_id"`
	SpotifyArtistID  string    `json:"spotify_artist_id"`
	TrackName        string    `json:"track_name"`
	ArtistName       string    `json:"artist_name"`
	AlbumName        string    `json:"album_name"`
	SpotifyAlbumURL  string    `json:"spotify_album_url"`
	SpotifyArtistURL string    `json:"spotify_artist_url"`
	DurationMs       int       `json:"duration_ms"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Request/Response DTOs
type CreatePlaylistRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	IsPublic        bool   `json:"is_public"`
	SyncWithSpotify bool   `json:"sync_with_spotify"`
}

type UpdatePlaylistRequest struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	IsPublic        bool   `json:"is_public"`
	SyncWithSpotify bool   `json:"sync_with_spotify"`
}

type ImportPlaylistRequest struct {
	SpotifyPlaylistID string `json:"spotify_playlist_id"`
	SyncWithSpotify   bool   `json:"sync_with_spotify"`
}

type ExportPlaylistRequest struct {
	PlaylistID uuid.UUID `json:"playlist_id"`
	Name       string    `json:"name" `
	IsPublic   bool      `json:"is_public"`
}

type AddSongRequest struct {
	SpotifyTrackID string `json:"spotify_track_id"`
}

type ReorderSongsRequest struct {
	SongOrder []uuid.UUID `json:"song_order"`
}
