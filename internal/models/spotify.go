package models

import (
	"time"
)

// =============================================================================
// SEARCH RESULTS (from Spotify search API)
// =============================================================================

type TrackSearch struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Popularity   int      `json:"popularity"`
	DurationMs   int      `json:"duration_ms"`
	Explicit     bool     `json:"explicit"`
	PreviewURL   string   `json:"preview_url"`
	ArtistNames  []string `json:"artist_names"`  // Simple list for search display
	AlbumName    string   `json:"album_name"`    // Simple string for search display
	AlbumImageURL string  `json:"album_image_url"`
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

// ArtistData represents a complete Spotify artist with all metadata
type ArtistData struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	ImageURL       string    `json:"image_url"`
	Popularity     int       `json:"popularity"`
	FollowersTotal int       `json:"followers_total"`
	Genres         []string  `json:"genres"`
	CachedAt       time.Time `json:"cached_at"`
	UpdatedAt      time.Time `json:"updated_at"`
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
	Artists              []ArtistData `json:"artists"` // All artists on this album
	CachedAt             time.Time   `json:"cached_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
}

// TrackArtist represents an artist's relationship to a specific track
type TrackArtist struct {
	ArtistData
	IsPrimary bool `json:"is_primary"` // True for main artist, false for features
}

// TrackData represents a complete Spotify track with full relationships
type TrackData struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Album       *AlbumData    `json:"album"`       // Full album info with artists
	Artists     []TrackArtist `json:"artists"`     // All artists (primary + features)
	DurationMs  int           `json:"duration_ms"`
	DiscNumber  int           `json:"disc_number"`
	TrackNumber int           `json:"track_number"`
	Popularity  int           `json:"popularity"`
	Explicit    bool          `json:"explicit"`
	PreviewURL  string        `json:"preview_url"`
	IsLocal     bool          `json:"is_local"`
	CachedAt    time.Time     `json:"cached_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// PlaylistData represents Spotify playlist metadata (for caching)
type PlaylistData struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	OwnerID          string    `json:"owner_id"`           // Spotify user ID
	OwnerDisplayName string    `json:"owner_display_name"`
	Public           bool      `json:"public"`
	Collaborative    bool      `json:"collaborative"`
	FollowersTotal   int       `json:"followers_total"`
	TotalTracks      int       `json:"total_tracks"`
	ImageURL         string    `json:"image_url"`
	CachedAt         time.Time `json:"cached_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// =============================================================================
// CONVENIENCE METHODS
// =============================================================================

// GetPrimaryArtist returns the main artist for a track
func (t *TrackData) GetPrimaryArtist() *ArtistData {
	for _, artist := range t.Artists {
		if artist.IsPrimary {
			return &artist.ArtistData
		}
	}
	// Fallback to first artist if no primary marked
	if len(t.Artists) > 0 {
		return &t.Artists[0].ArtistData
	}
	return nil
}

// GetArtistNames returns all artist names as a slice
func (t *TrackData) GetArtistNames() []string {
	names := make([]string, len(t.Artists))
	for i, artist := range t.Artists {
		names[i] = artist.Name
	}
	return names
}

// GetPrimaryArtistName returns the main artist name, or first artist name
func (t *TrackData) GetPrimaryArtistName() string {
	if primary := t.GetPrimaryArtist(); primary != nil {
		return primary.Name
	}
	return "Unknown Artist"
}

// GetAlbumName returns the album name, with fallback
func (t *TrackData) GetAlbumName() string {
	if t.Album != nil {
		return t.Album.Name
	}
	return "Unknown Album"
}
