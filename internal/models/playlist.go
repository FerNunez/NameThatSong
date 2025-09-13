package models

import (
	"time"

	"github.com/google/uuid"
)

// Business domain models following existing patterns
type LocalPlaylist struct {
	ID                uuid.UUID   `json:"id"`
	UserID            uuid.UUID   `json:"user_id"`
	Name              string      `json:"name"`
	Description       string      `json:"description"`
	SpotifyPlaylistID *string     `json:"spotify_playlist_id"`
	IsPublic          bool        `json:"is_public"`
	SnapshotID        *string     `json:"snapshot_id"`
	LastSyncedAt      *time.Time  `json:"last_synced_at"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	Tracks            []TrackData `json:"tracks"`
}

// PlaylistTrack represents a track in a playlist with basic metadata from spotify_tracks
type PlaylistTrack struct {
	SpotifyTrackID string    `json:"spotify_track_id"`
	Position       int32     `json:"position"`
	UpdatedAt      time.Time `json:"updated_at"`
	TrackName      string    `json:"track_name"`
	DurationMs     int32     `json:"duration_ms"`
	AlbumID        string    `json:"album_id"`
	ArtistIds      []string  `json:"artist_ids"`
}

// PlaylistTrackWithDetails represents enriched track data with album/artist names via 3-tier caching
type PlaylistTrackWithDetails struct {
	PlaylistTrack
	AlbumName             string   `json:"album_name,omitempty"`
	AlbumImageUrl         string   `json:"album_image_url,omitempty"`
	ArtistNames           []string `json:"artist_names,omitempty"`
	PrimaryArtistImageUrl string   `json:"primary_artist_image_url,omitempty"`
}

// Request/Response DTOs
type CreatePlaylistRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
	SnapshotID  string `json:"snapshot_id"`
}

type UpdatePlaylistRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
	SnapshotID  string `json:"snapshot_id"`
}

type ImportPlaylistRequest struct {
	SpotifyPlaylistID string `json:"spotify_playlist_id"`
	SnapshotID        string `json:"snapshot_id"`
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
