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
	Songs             []PlaylistSong `json:"songs,omitempty"`
}

type PlaylistSong struct {
	ID             uuid.UUID `json:"id"`
	PlaylistID     uuid.UUID `json:"playlist_id"`
	SpotifyTrackID string    `json:"spotify_track_id"`
	Position       int       `json:"position"`
	TrackName      string    `json:"track_name"`
	ArtistName     string    `json:"artist_name"`
	AlbumName      string    `json:"album_name"`
	DurationMs     int       `json:"duration_ms"`
	AddedAt        time.Time `json:"added_at"`
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