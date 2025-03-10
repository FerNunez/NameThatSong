package base

import (
	"time"
)

type Song struct {
	Id   string
	Name string
}

func NewSong(id, name string) Song {
	return Song{
		id,
		name,
	}
}

// Base
type Image struct {
	URL    string
	Height int
	Width  int
}

type Artist struct {
	TotalFollowers int
	Genres         []string
	Href           string // `json:"href"`   // API endpoint URL for this specific artist
	ID             string
	Images         []Image
	Name           string
	Popularity     int
	Type           string
	URI            string
	AlbumsID       []string
}

type Album struct {
	AlbumType        string
	TotalTracks      int
	AvailableMarkets []string
	Href             string
	ID               string
	Images           []Image
	Name             string
	ReleaseDate      time.Time
	URI              string
	ArtistsFeatureID []string
}

type Track struct {
	DiscNumber  int
	DurationMs  int
	Href        string
	ID          string
	IsPlayable  bool
	Name        string
	Popularity  int
	TrackNumber int
	URI         string
	AlbumID     string
	ArtistsID   []string
}

type TrackJson struct {
	Album struct {
		AlbumType string `json:"album_type"`
		Artists   []struct {
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href string `json:"href"`
			ID   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
			URI  string `json:"uri"`
		} `json:"artists"`

		AvailableMarkets []string `json:"available_markets"`
		ExternalUrls     struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`

		Href   string `json:"href"`
		ID     string `json:"id"`
		Images []struct {
			Height int    `json:"height"`
			Width  int    `json:"width"`
			URL    string `json:"url"`
		} `json:"images"`
		IsPlayable           bool   `json:"is_playable"`
		Name                 string `json:"name"`
		ReleaseDate          string `json:"release_date"`
		ReleaseDatePrecision string `json:"release_date_precision"`
		TotalTracks          int    `json:"total_tracks"`
		Type                 string `json:"type"`
		URI                  string `json:"uri"`
	} `json:"album"`
	Artists []struct {
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href string `json:"href"`
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
		URI  string `json:"uri"`
	} `json:"artists"`
	AvailableMarkets []string `json:"available_markets"`

	DiscNumber  int  `json:"disc_number"`
	DurationMs  int  `json:"duration_ms"`
	Explicit    bool `json:"explicit"`
	ExternalIds struct {
		Isrc string `json:"isrc"`
	} `json:"external_ids"`
	ExternalUrls struct {
		Spotify string `json:"spotify"`
	} `json:"external_urls"`
	Href        string      `json:"href"`
	ID          string      `json:"id"`
	IsLocal     bool        `json:"is_local"`
	IsPlayable  bool        `json:"is_playable"`
	Name        string      `json:"name"`
	Popularity  int         `json:"popularity"`
	PreviewURL  interface{} `json:"preview_url"`
	TrackNumber int         `json:"track_number"`
	Type        string      `json:"type"`
	URI         string      `json:"uri"`
}

type AlbumsJson struct {
	Href     string      `json:"href"`     // API endpoint URL for the album search results
	Limit    int         `json:"limit"`    // Maximum number of items returned per response
	Next     string      `json:"next"`     // URL for the next page of results (null if none)
	Offset   int         `json:"offset"`   // Starting index of the returned items
	Previous interface{} `json:"previous"` // URL for the previous page of results (null if none)
	Total    int         `json:"total"`    // Total number of items available
	Items    []struct {  // Array of album objects
		AlbumType        string   `json:"album_type"`        // Type of album (e.g., "album", "single", "compilation")
		TotalTracks      int      `json:"total_tracks"`      // Total number of tracks in the album
		AvailableMarkets []string `json:"available_markets"` // List of markets where the album is available
		ExternalUrls     struct {
			Spotify string `json:"spotify"` // Spotify's public album page URL
		} `json:"external_urls"`
		Href   string     `json:"href"` // API endpoint URL for this specific album
		ID     string     `json:"id"`   // Unique Spotify album identifier
		Images []struct { // Album cover art in various sizes
			URL    string `json:"url"`    // Image URL
			Height int    `json:"height"` // Image height in pixels
			Width  int    `json:"width"`  // Image width in pixels
		} `json:"images"`
		Name                 string     `json:"name"`                   // Album name
		ReleaseDate          string     `json:"release_date"`           // Release date of the album (format varies)
		ReleaseDatePrecision string     `json:"release_date_precision"` // Precision of the release date (e.g., "year", "month", "day")
		Type                 string     `json:"type"`                   // Resource type ("album")
		URI                  string     `json:"uri"`                    // Spotify URI for direct access
		Artists              []struct { // Array of artists associated with the album
			ExternalUrls struct {
				Spotify string `json:"spotify"` // Spotify's public artist page URL
			} `json:"external_urls"`
			Href string `json:"href"` // API endpoint URL for this specific artist
			ID   string `json:"id"`   // Unique Spotify artist identifier
			Name string `json:"name"` // Artist display name
			Type string `json:"type"` // Resource type ("artist")
			URI  string `json:"uri"`  // Spotify URI for direct access
		} `json:"artists"`
	} `json:"items"`
}

type ArtistsJson struct {
	Href string `json:"href"` // API endpoint URL for the artist search results

	Limit  int    `json:"limit"`  // Maximum number of items returned per response
	Next   string `json:"next"`   // URL for next page of results (null if none)
	Offset int    `json:"offset"` // Starting index of the returned items

	Previous interface{} `json:"previous"` // URL for previous page of results (null if none)
	Total    int         `json:"total"`    // Total number of items available
	Items    []struct {  // Array of artist objects
		ExternalUrls struct {
			Spotify string `json:"spotify"` // Spotify's public artist page URL
		} `json:"external_urls"`
		Followers struct {
			Href  interface{} `json:"href"`  // Deprecated field (typically empty)
			Total int         `json:"total"` // Total number of followers
		} `json:"followers"`
		Genres []string   `json:"genres"` // List of associated music genres
		Href   string     `json:"href"`   // API endpoint URL for this specific artist
		ID     string     `json:"id"`     // Unique Spotify artist identifier
		Images []struct { // Artist profile images in various sizes
			URL    string `json:"url"`
			Height int    `json:"height"` // Image height in pixels
			Width  int    `json:"width"`  // Image width in pixels
		} `json:"images"`
		Name       string `json:"name"`       // Artist display name
		Popularity int    `json:"popularity"` // Popularity score (0-100)
		Type       string `json:"type"`       // Resource type ("artist")
		URI        string `json:"uri"`        // Spotify URI for direct access
	} `json:"items"`
}
