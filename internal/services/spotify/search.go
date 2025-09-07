package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	m "github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
)

// =============================================================================
// SEARCH METHODS
// =============================================================================

// SearchTracks searches for tracks with caching
func (s *Spotify) SearchTracks(ctx context.Context, userID, query string) ([]m.TrackSearch, error) {
	// Check cache first
	if cachedTracks, found := s.cache.GetSearchTracks(query); found {
		logger.Debug(ctx, "cache hit for track search",
			logger.F("query", query),
			logger.F("results_count", len(cachedTracks)))
		return cachedTracks, nil
	}

	// Get access token for user
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	// Cache miss - search Spotify API
	tracks, err := s.searchTracksFromAPI(ctx, accessToken, query)
	if err != nil {
		return nil, err
	}

	// Cache the results
	s.cache.SetSearchTracks(query, tracks)
	logger.Debug(ctx, "track search completed",
		logger.F("query", query),
		logger.F("results_count", len(tracks)))

	return tracks, nil
}

// SearchAlbums searches for albums with caching
func (s *Spotify) SearchAlbums(ctx context.Context, userID, query string) ([]m.AlbumSearch, error) {
	if cachedAlbums, found := s.cache.GetSearchAlbums(query); found {
		logger.Debug(ctx, "cache hit for album search",
			logger.F("query", query),
			logger.F("results_count", len(cachedAlbums)))
		return cachedAlbums, nil
	}

	// Get access token for user
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	albums, err := s.searchAlbumsFromAPI(ctx, accessToken, query)
	if err != nil {
		return nil, err
	}

	s.cache.SetSearchAlbums(query, albums)
	logger.Debug(ctx, "album search completed",
		logger.F("query", query),
		logger.F("results_count", len(albums)))

	return albums, nil
}

// SearchArtists searches for artists with caching
func (s *Spotify) SearchArtists(ctx context.Context, userID, query string) ([]m.ArtistSearch, error) {
	if cachedArtists, found := s.cache.GetSearchArtists(query); found {
		logger.Debug(ctx, "cache hit for artist search",
			logger.F("query", query),
			logger.F("results_count", len(cachedArtists)))
		return cachedArtists, nil
	}

	// Get access token for user
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	artists, err := s.searchArtistsFromAPI(ctx, accessToken, query)
	if err != nil {
		return nil, err
	}

	s.cache.SetSearchArtists(query, artists)
	logger.Debug(ctx, "artist search completed",
		logger.F("query", query),
		logger.F("results_count", len(artists)))

	return artists, nil
}

// SearchPlaylists searches for playlists with caching
func (s *Spotify) SearchPlaylists(ctx context.Context, userID, query string) ([]m.PlaylistSearch, error) {
	if cachedPlaylists, found := s.cache.GetSearchPlaylists(query); found {
		logger.Debug(ctx, "cache hit for playlist search",
			logger.F("query", query),
			logger.F("results_count", len(cachedPlaylists)))
		return cachedPlaylists, nil
	}

	// Get access token for user
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	playlists, err := s.searchPlaylistsFromAPI(ctx, accessToken, query)
	if err != nil {
		return nil, err
	}

	s.cache.SetSearchPlaylists(query, playlists)
	logger.Debug(ctx, "playlist search completed",
		logger.F("query", query),
		logger.F("results_count", len(playlists)))

	return playlists, nil
}

// SearchAll searches for tracks, albums, artists, and playlists in a single API call
func (s *Spotify) SearchAll(ctx context.Context, userID, query string) (*m.SearchAllResults, error) {
	// Check cache first
	if cachedResults, found := s.cache.GetSearchAll(query); found {
		logger.Debug(ctx, "cache hit for search all",
			logger.F("query", query),
			logger.F("tracks_count", len(cachedResults.Tracks)),
			logger.F("albums_count", len(cachedResults.Albums)),
			logger.F("artists_count", len(cachedResults.Artists)),
			logger.F("playlists_count", len(cachedResults.Playlists)))
		return cachedResults, nil
	}

	// Get access token for user
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	// Cache miss - search Spotify API
	results, err := s.searchAllFromAPI(ctx, accessToken, query)
	if err != nil {
		return nil, err
	}

	// Cache the combined results
	s.cache.SetSearchAll(query, results)

	// Also cache individual results for future single-type searches
	s.cache.SetSearchTracks(query, results.Tracks)
	s.cache.SetSearchAlbums(query, results.Albums)
	s.cache.SetSearchArtists(query, results.Artists)
	s.cache.SetSearchPlaylists(query, results.Playlists)

	logger.Debug(ctx, "search all completed",
		logger.F("query", query),
		logger.F("tracks_count", len(results.Tracks)),
		logger.F("albums_count", len(results.Albums)),
		logger.F("artists_count", len(results.Artists)),
		logger.F("playlists_count", len(results.Playlists)))

	return results, nil
}

// Private helper methods
func (s *Spotify) search(ctx context.Context, accessToken, limit, atype, query string) ([]byte, error) {
	// Handle query format based on whether we have multiple types or single type
	var formattedQuery string
	if strings.Contains(atype, ",") {
		// Multiple types: use raw query without type prefix
		formattedQuery = strings.ToLower(query)
	} else {
		// Single type: use type prefix (existing behavior)
		formattedQuery = atype + ":" + strings.ToLower(query)
	}

	logger.Debug(ctx, "building Spotify search request",
		logger.F("search_type", atype),
		logger.F("query", query),
		logger.F("formatted_query", formattedQuery),
		logger.F("limit", limit),
		logger.F("multiple_types", strings.Contains(atype, ",")))

	apiURL, err := url.Parse(s.config.GetAPIBaseURL() + "/search")
	if err != nil {
		logger.Error(ctx, "failed to parse Spotify API URL",
			logger.F("error", err))
		return nil, err
	}
	q := apiURL.Query()
	q.Set("type", atype)
	q.Set("q", formattedQuery)
	q.Set("limit", limit)
	apiURL.RawQuery = q.Encode()

	logger.Debug(ctx, "making HTTP request to Spotify API",
		logger.F("url", apiURL.String()),
		logger.F("method", "GET"),
		logger.F("search_type", atype))

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL.String(), nil)
	if err != nil {
		logger.Error(ctx, "failed to create HTTP request",
			logger.F("error", err))
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	resp, err := s.httpClient.Do(req)
	if err != nil {
		logger.Error(ctx, "HTTP request to Spotify API failed",
			logger.F("url", apiURL.String()),
			logger.F("error", err))
		return nil, err
	}
	defer resp.Body.Close()

	logger.Debug(ctx, "received HTTP response from Spotify API",
		logger.F("status_code", resp.StatusCode),
		logger.F("content_length", resp.ContentLength),
		logger.F("search_type", atype))

	if resp.StatusCode != http.StatusOK {
		logger.Error(ctx, "Spotify API returned non-200 status code",
			logger.F("status_code", resp.StatusCode),
			logger.F("url", apiURL.String()),
			logger.F("search_type", atype))
		return nil, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(ctx, "failed to read response body from Spotify API",
			logger.F("error", err))
		return nil, err
	}

	logger.Debug(ctx, "successfully read Spotify API response",
		logger.F("response_size_bytes", len(data)),
		logger.F("search_type", atype))

	return data, nil
}

func (s *Spotify) searchTracksFromAPI(ctx context.Context, accessToken, name string) ([]m.TrackSearch, error) {
	limit := "50"
	data, err := s.search(ctx, accessToken, limit, "track", name)
	if err != nil {
		return nil, err
	}
	return s.parseTrackSearchResponse(data)
}

func (s *Spotify) searchAlbumsFromAPI(ctx context.Context, accessToken, name string) ([]m.AlbumSearch, error) {
	limit := "50"
	data, err := s.search(ctx, accessToken, limit, "album", name)
	if err != nil {
		return nil, err
	}
	return s.parseAlbumSearchResponse(data)
}

func (s *Spotify) searchArtistsFromAPI(ctx context.Context, accessToken, name string) ([]m.ArtistSearch, error) {
	limit := "50"
	data, err := s.search(ctx, accessToken, limit, "artist", name)
	if err != nil {
		return nil, err
	}
	return s.parseArtistSearchResponse(data)
}

func (s *Spotify) searchPlaylistsFromAPI(ctx context.Context, accessToken, name string) ([]m.PlaylistSearch, error) {
	limit := "50"
	data, err := s.search(ctx, accessToken, limit, "playlist", name)
	if err != nil {
		return nil, err
	}
	return s.parsePlaylistSearchResponse(data)
}

func (s *Spotify) searchAllFromAPI(ctx context.Context, accessToken, name string) (*m.SearchAllResults, error) {
	limit := "20" // Reduced per-type to manage response size
	multipleTypes := "track,album,artist,playlist"
	data, err := s.search(ctx, accessToken, limit, multipleTypes, name)
	if err != nil {
		return nil, err
	}
	return s.parseSearchAllResponse(ctx, data, name)
}

func (s *Spotify) parseTrackSearchResponse(data []byte) ([]m.TrackSearch, error) {

	type SearchTrachByNameResponse struct {
		Tracks struct {
			Href     string `json:"href"`
			Limit    int    `json:"limit"`
			Next     string `json:"next"`
			Offset   int    `json:"offset"`
			Previous any    `json:"previous"`
			Total    int    `json:"total"`
			Items    []struct {
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
				DiscNumber       int      `json:"disc_number"`
				DurationMs       int      `json:"duration_ms"`
				Explicit         bool     `json:"explicit"`
				ExternalIds      struct {
					Isrc string `json:"isrc"`
				} `json:"external_ids"`
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href        string `json:"href"`
				ID          string `json:"id"`
				IsLocal     bool   `json:"is_local"`
				IsPlayable  bool   `json:"is_playable"`
				Name        string `json:"name"`
				Popularity  int    `json:"popularity"`
				PreviewURL  any    `json:"preview_url"`
				TrackNumber int    `json:"track_number"`
				Type        string `json:"type"`
				URI         string `json:"uri"`
			} `json:"items"`
		} `json:"tracks"`
	}
	var searchTrackResponse SearchTrachByNameResponse
	if err := json.Unmarshal(data, &searchTrackResponse); err != nil {
		return nil, err
	}

	// Convert to trackInfo
	tracks := make([]m.TrackSearch, 0, len(searchTrackResponse.Tracks.Items))
	for _, t := range searchTrackResponse.Tracks.Items {

		artists := make([]string, len(t.Artists))
		for idx, a := range t.Artists {
			artists[idx] = a.Name
		}

		// Get album info
		albumsName := ""
		albumsImageURL := ""
		if len(t.Album.Images) > 0 {
			albumsImageURL = t.Album.Images[0].URL
		}
		if t.Album.Name != "" {
			albumsName = t.Album.Name
		}

		trackInfo := m.TrackSearch{
			ID:         t.ID,
			Name:       t.Name,
			Popularity: t.Popularity,
			DurationMs: t.DurationMs,
			Explicit:   t.Explicit,
			PreviewURL: func() string {
				if url, ok := t.PreviewURL.(string); ok {
					return url
				}
				return ""
			}(),
			ArtistNames:   artists,
			AlbumName:     albumsName,
			AlbumImageURL: albumsImageURL,
		}
		tracks = append(tracks, trackInfo)
	}
	return tracks, nil
}

func (s *Spotify) parseAlbumSearchResponse(data []byte) ([]m.AlbumSearch, error) {

	type SearchAlbumResponse struct {
		Albums struct {
			Href     string `json:"href"`
			Limit    int    `json:"limit"`
			Next     string `json:"next"`
			Offset   int    `json:"offset"`
			Previous any    `json:"previous"`
			Total    int    `json:"total"`
			Items    []struct {
				AlbumType        string   `json:"album_type"`
				TotalTracks      int      `json:"total_tracks"`
				AvailableMarkets []string `json:"available_markets"`
				ExternalUrls     struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href   string `json:"href"`
				ID     string `json:"id"`
				Images []struct {
					Height int    `json:"height"`
					URL    string `json:"url"`
					Width  int    `json:"width"`
				} `json:"images"`
				Name                 string `json:"name"`
				ReleaseDate          string `json:"release_date"`
				ReleaseDatePrecision string `json:"release_date_precision"`
				Type                 string `json:"type"`
				URI                  string `json:"uri"`
				Artists              []struct {
					ExternalUrls struct {
						Spotify string `json:"spotify"`
					} `json:"external_urls"`
					Href string `json:"href"`
					ID   string `json:"id"`
					Name string `json:"name"`
					Type string `json:"type"`
					URI  string `json:"uri"`
				} `json:"artists"`
			} `json:"items"`
		} `json:"albums"`
	}
	var searchAlbumResponse SearchAlbumResponse
	if err := json.Unmarshal(data, &searchAlbumResponse); err != nil {
		return nil, err
	}

	albums := make([]m.AlbumSearch, len(searchAlbumResponse.Albums.Items))
	for idx, alb := range searchAlbumResponse.Albums.Items {

		artists := make([]string, len(alb.Artists))
		for idx, art := range alb.Artists {
			artists[idx] = art.Name
		}

		imageUrl := ""
		if len(alb.Images) > 0 {
			imageUrl = alb.Images[0].URL
		}

		albums[idx] = m.AlbumSearch{
			ID:          alb.ID,
			Name:        alb.Name,
			AlbumType:   alb.AlbumType,
			ReleaseDate: alb.ReleaseDate,
			TotalTracks: alb.TotalTracks,
			ImageURL:    imageUrl,
			ArtistNames: artists,
		}
	}
	return albums, nil
}

func (s *Spotify) parseArtistSearchResponse(data []byte) ([]m.ArtistSearch, error) {

	var searchArtistResponse struct {
		Artists struct {
			Items []struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				Popularity int    `json:"popularity"`
				Images     []struct {
					URL    string `json:"url"`
					Height int    `json:"height"`
					Width  int    `json:"width"`
				} `json:"images"`
			} `json:"items"`
		} `json:"artists"`
	}

	if err := json.Unmarshal(data, &searchArtistResponse); err != nil {
		return nil, err
	}

	// Convert to ArtistInfo
	artists := make([]m.ArtistSearch, 0, len(searchArtistResponse.Artists.Items))
	for _, a := range searchArtistResponse.Artists.Items {

		// TODO: add here and temp url???
		imageUrl := ""
		if len(a.Images) > 0 {
			imageUrl = a.Images[0].URL
		}
		artistInfo := m.ArtistSearch{
			ID:             a.ID,
			Name:           a.Name,
			ImageURL:       imageUrl,
			Popularity:     a.Popularity,
			FollowersTotal: 0,          // Not available in search response
			Genres:         []string{}, // Not available in search response
		}
		artists = append(artists, artistInfo)
	}
	return artists, nil
}

func (s *Spotify) parsePlaylistSearchResponse(data []byte) ([]m.PlaylistSearch, error) {

	type SearchPlaylistResponse struct {
		Playlists struct {
			Href     string  `json:"href"`
			Limit    int     `json:"limit"`
			Next     *string `json:"next"`
			Offset   int     `json:"offset"`
			Previous *string `json:"previous"`
			Total    int     `json:"total"`
			Items    []*struct {
				Collaborative bool   `json:"collaborative"`
				Description   string `json:"description"`
				ExternalUrls  struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href   string `json:"href"`
				ID     string `json:"id"`
				Images []struct {
					Height *int   `json:"height"`
					URL    string `json:"url"`
					Width  *int   `json:"width"`
				} `json:"images"`
				Name  string `json:"name"`
				Owner struct {
					DisplayName  string `json:"display_name"`
					ExternalUrls struct {
						Spotify string `json:"spotify"`
					} `json:"external_urls"`
					Href string `json:"href"`
					ID   string `json:"id"`
					Type string `json:"type"`
					URI  string `json:"uri"`
				} `json:"owner"`
				PrimaryColor *string `json:"primary_color"` // pointer because it can be null
				Public       bool    `json:"public"`
				SnapshotID   string  `json:"snapshot_id"`
				Tracks       struct {
					Href  string `json:"href"`
					Total int    `json:"total"`
				} `json:"tracks"`
				Type string `json:"type"`
				URI  string `json:"uri"`
			} `json:"items"` // slice of pointers to handle null values
		} `json:"playlists"`
	}

	var searchPlaylistResponse SearchPlaylistResponse
	if err := json.Unmarshal(data, &searchPlaylistResponse); err != nil {
		return nil, err
	}

	playlists := make([]m.PlaylistSearch, 0, len(searchPlaylistResponse.Playlists.Items))
	for _, p := range searchPlaylistResponse.Playlists.Items {

		if p == nil {
			continue
		}

		imageUrl := ""
		if len(p.Images) > 0 {
			imageUrl = p.Images[0].URL
		}

		playlists = append(playlists, m.PlaylistSearch{
			ID:             p.ID,
			Name:           p.Name,
			Description:    p.Description,
			ImageURL:       imageUrl,
			OwnerName:      p.Owner.DisplayName,
			Public:         p.Public,
			TotalTracks:    p.Tracks.Total,
			FollowersTotal: 0, // Not available in search response
		})

	}
	return playlists, nil
}

func (s *Spotify) parseSearchAllResponse(ctx context.Context, data []byte, query string) (*m.SearchAllResults, error) {
	// Combined response structure matching Spotify API
	type SearchAllResponse struct {
		Tracks struct {
			Items []struct {
				ID         string      `json:"id"`
				Name       string      `json:"name"`
				Popularity int         `json:"popularity"`
				DurationMs int         `json:"duration_ms"`
				Explicit   bool        `json:"explicit"`
				PreviewURL interface{} `json:"preview_url"`
				Artists    []struct {
					Name string `json:"name"`
				} `json:"artists"`
				Album struct {
					Name   string `json:"name"`
					Images []struct {
						URL string `json:"url"`
					} `json:"images"`
				} `json:"album"`
			} `json:"items"`
		} `json:"tracks"`

		Albums struct {
			Items []struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				AlbumType   string `json:"album_type"`
				ReleaseDate string `json:"release_date"`
				TotalTracks int    `json:"total_tracks"`
				Images      []struct {
					URL string `json:"url"`
				} `json:"images"`
				Artists []struct {
					Name string `json:"name"`
				} `json:"artists"`
			} `json:"items"`
		} `json:"albums"`

		Artists struct {
			Items []struct {
				ID         string `json:"id"`
				Name       string `json:"name"`
				Popularity int    `json:"popularity"`
				Images     []struct {
					URL string `json:"url"`
				} `json:"images"`
			} `json:"items"`
		} `json:"artists"`

		Playlists struct {
			Items []*struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
				Public      bool   `json:"public"`
				Images      []struct {
					URL string `json:"url"`
				} `json:"images"`
				Owner struct {
					DisplayName string `json:"display_name"`
				} `json:"owner"`
				Tracks struct {
					Total int `json:"total"`
				} `json:"tracks"`
			} `json:"items"`
		} `json:"playlists"`
	}

	var response SearchAllResponse
	if err := json.Unmarshal(data, &response); err != nil {
		logger.Error(ctx, "failed to unmarshal search all response",
			logger.F("query", query),
			logger.F("error", err))
		return nil, err
	}

	results := &m.SearchAllResults{
		Query:     query,
		Tracks:    make([]m.TrackSearch, 0, len(response.Tracks.Items)),
		Albums:    make([]m.AlbumSearch, 0, len(response.Albums.Items)),
		Artists:   make([]m.ArtistSearch, 0, len(response.Artists.Items)),
		Playlists: make([]m.PlaylistSearch, 0, len(response.Playlists.Items)),
	}

	// Parse tracks
	for _, t := range response.Tracks.Items {
		artists := make([]string, len(t.Artists))
		for idx, a := range t.Artists {
			artists[idx] = a.Name
		}

		albumImageURL := ""
		if len(t.Album.Images) > 0 {
			albumImageURL = t.Album.Images[0].URL
		}

		previewURL := ""
		if url, ok := t.PreviewURL.(string); ok {
			previewURL = url
		}

		results.Tracks = append(results.Tracks, m.TrackSearch{
			ID:            t.ID,
			Name:          t.Name,
			Popularity:    t.Popularity,
			DurationMs:    t.DurationMs,
			Explicit:      t.Explicit,
			PreviewURL:    previewURL,
			ArtistNames:   artists,
			AlbumName:     t.Album.Name,
			AlbumImageURL: albumImageURL,
		})
	}

	// Parse albums
	for _, alb := range response.Albums.Items {
		artists := make([]string, len(alb.Artists))
		for idx, art := range alb.Artists {
			artists[idx] = art.Name
		}

		imageUrl := ""
		if len(alb.Images) > 0 {
			imageUrl = alb.Images[0].URL
		}

		results.Albums = append(results.Albums, m.AlbumSearch{
			ID:          alb.ID,
			Name:        alb.Name,
			AlbumType:   alb.AlbumType,
			ReleaseDate: alb.ReleaseDate,
			TotalTracks: alb.TotalTracks,
			ImageURL:    imageUrl,
			ArtistNames: artists,
		})
	}

	// Parse artists
	for _, a := range response.Artists.Items {
		imageUrl := ""
		if len(a.Images) > 0 {
			imageUrl = a.Images[0].URL
		}

		results.Artists = append(results.Artists, m.ArtistSearch{
			ID:             a.ID,
			Name:           a.Name,
			ImageURL:       imageUrl,
			Popularity:     a.Popularity,
			FollowersTotal: 0,          // Not available in search response
			Genres:         []string{}, // Not available in search response
		})
	}

	// Parse playlists
	for _, p := range response.Playlists.Items {
		if p == nil {
			continue
		}

		imageUrl := ""
		if len(p.Images) > 0 {
			imageUrl = p.Images[0].URL
		}

		results.Playlists = append(results.Playlists, m.PlaylistSearch{
			ID:             p.ID,
			Name:           p.Name,
			Description:    p.Description,
			ImageURL:       imageUrl,
			OwnerName:      p.Owner.DisplayName,
			Public:         p.Public,
			TotalTracks:    p.Tracks.Total,
			FollowersTotal: 0, // Not available in search response
		})
	}

	logger.Debug(ctx, "parsed search all response",
		logger.F("query", query),
		logger.F("tracks_parsed", len(results.Tracks)),
		logger.F("albums_parsed", len(results.Albums)),
		logger.F("artists_parsed", len(results.Artists)),
		logger.F("playlists_parsed", len(results.Playlists)))

	return results, nil
}
