package models

import (
	"time"
)

type SpotifyID string

// =============================================================================
// SEARCH RESULTS (from Spotify search API)
// =============================================================================
// Sadly, when searching albums we dont get their tracks as response from Spotify Api.. So we need to fetch this info
// from another endpoint

type TrackSearch struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Popularity    int      `json:"popularity"`
	DurationMs    int      `json:"duration_ms"`
	Explicit      bool     `json:"explicit"`
	PreviewURL    string   `json:"preview_url"`
	ArtistNames   []string `json:"artist_names"` // Simple list for search display
	AlbumName     string   `json:"album_name"`   // Simple string for search display
	AlbumImageURL string   `json:"album_image_url"`
}

type AlbumSearch struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	AlbumType   string   `json:"album_type"`
	ReleaseDate string   `json:"release_date"`
	TotalTracks int      `json:"total_tracks"`
	ImageURL    string   `json:"image_url"`
	ArtistNames []string `json:"artist_names"` // Simple list for search display
}

type ArtistSearch struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	ImageURL       string   `json:"image_url"`
	Popularity     int      `json:"popularity"`
	FollowersTotal int      `json:"followers_total"`
	Genres         []string `json:"genres"`
}

type PlaylistSearch struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	ImageURL       string `json:"image_url"`
	OwnerName      string `json:"owner_name"`
	Public         bool   `json:"public"`
	TotalTracks    int    `json:"total_tracks"`
	FollowersTotal int    `json:"followers_total"`
}

// =============================================================================
// FULL DATA MODELS (with relationships, for caching and database)
// =============================================================================

// TrackData represents a complete Spotify track with full relationships
type TrackData struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	DurationMs  int         `json:"duration_ms"`
	DiscNumber  int         `json:"disc_number"`
	TrackNumber int         `json:"track_number"`
	Popularity  int         `json:"popularity"`
	Explicit    bool        `json:"explicit"`
	IsLocal     bool        `json:"is_local"`
	AlbumID     SpotifyID   `json:"album_id"`   // ID of album
	ArtistIDs   []SpotifyID `json:"artist_ids"` // IDs of  artists
	CachedAt    time.Time   `json:"cached_at"`  // when was this track last cached from spotify api
}

// AlbumData represents a complete Spotify album with artist relationships
type AlbumData struct {
	ID                   string      `json:"id"`
	Name                 string      `json:"name"`
	AlbumType            string      `json:"album_type"` // album, single, compilation
	ReleaseDate          string      `json:"release_date"`
	ReleaseDatePrecision string      `json:"release_date_precision"` // year, month, day
	TotalTracks          int         `json:"total_tracks"`
	ImageURL             string      `json:"image_url"`
	Label                string      `json:"label"`
	Popularity           int         `json:"popularity"`
	ArtistIDs            []SpotifyID `json:"artist_ids"` // IDs of artists in this album
	TrackIDs             []SpotifyID `json:"track_ids"`  // IDs of artists in this album
	CachedAt             time.Time   `json:"cached_at"`  // when was this track last cached from spotify api
}

// ArtistData represents a complete Spotify artist with all metadata
type ArtistData struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	ImageURL       string    `json:"image_url"`
	Popularity     int       `json:"popularity"`
	FollowersTotal int       `json:"followers_total"`
	Genres         []string  `json:"genres"`
	CachedAt       time.Time `json:"cached_at"` // when was this track last cached from spotify api
}

// PlaylistData represents Spotify playlist metadata (for caching)
type PlaylistData struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Description      string      `json:"description"`
	OwnerID          string      `json:"owner_id"` // Spotify user ID
	OwnerDisplayName string      `json:"owner_display_name"`
	Public           bool        `json:"public"`
	Collaborative    bool        `json:"collaborative"`
	FollowersTotal   int         `json:"followers_total"`
	TotalTracks      int         `json:"total_tracks"`
	ImageURL         string      `json:"image_url"`
	TrackIDs         []SpotifyID `json:"track_ids"` // IDs of tracks in this album
	CachedAt         time.Time   `json:"cached_at"`
}
