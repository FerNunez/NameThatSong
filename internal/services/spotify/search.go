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
	"github.com/FerNunez/NameThatSong/internal/pkg/utils"
)

// =============================================================================
// SEARCH METHODS
// =============================================================================

// SearchTracks searches for tracks with caching
func (s *Spotify) SearchTracks(ctx context.Context, userID, query string) ([]m.TrackSearch, error) {
	logger.Info(ctx, "track search initiated",
		logger.F("user_id", userID),
		logger.F("query", query))
	
	normalizedQuery := utils.NormalizeSearchQuery(query)
	sanitizedQuery := utils.SanitizeForCacheKey(normalizedQuery)
	cacheKey := fmt.Sprintf("spotify:search:tracks:%s", sanitizedQuery)
	
	logger.Debug(ctx, "track search query processing",
		logger.F("original_query", query),
		logger.F("normalized_query", normalizedQuery),
		logger.F("sanitized_query", sanitizedQuery),
		logger.F("cache_key", cacheKey))
	
	// Check cache first
	if cachedTracks, found := s.cache.GetSearchTracks(query); found {
		logger.Debug(ctx, "cache hit for track search",
			logger.F("query", query),
			logger.F("normalized_query", normalizedQuery),
			logger.F("cache_key", cacheKey),
			logger.F("results_count", len(cachedTracks)))
		return cachedTracks, nil
	}

	logger.Info(ctx, "cache miss for track search - calling Spotify API",
		logger.F("query", query),
		logger.F("normalized_query", normalizedQuery),
		logger.F("cache_key", cacheKey))

	// Get access token for user
	logger.Debug(ctx, "retrieving access token for track search",
		logger.F("user_id", userID))
	
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		logger.Error(ctx, "failed to get access token for track search",
			logger.F("user_id", userID),
			logger.F("query", query),
			logger.F("error", err))
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	// Cache miss - search Spotify API
	logger.Info(ctx, "calling Spotify API for track search",
		logger.F("user_id", userID),
		logger.F("query", query),
		logger.F("normalized_query", normalizedQuery))
	
	tracks, err := s.searchTracksFromAPI(ctx, accessToken, query)
	if err != nil {
		logger.Error(ctx, "track search API call failed",
			logger.F("user_id", userID),
			logger.F("query", query),
			logger.F("error", err))
		return nil, err
	}

	// Cache the results
	s.cache.SetSearchTracks(query, tracks)
	logger.Info(ctx, "track search completed successfully",
		logger.F("user_id", userID),
		logger.F("query", query),
		logger.F("results_count", len(tracks)),
		logger.F("cache_key", cacheKey))
	
	logger.Debug(ctx, "cached track search results",
		logger.F("query", query),
		logger.F("results_count", len(tracks)))
	return tracks, nil
}

// SearchAlbums searches for albums with caching
func (s *Spotify) SearchAlbums(ctx context.Context, userID, query string) ([]m.AlbumSearch, error) {
	logger.Info(ctx, "album search initiated",
		logger.F("user_id", userID),
		logger.F("query", query))
	
	normalizedQuery := utils.NormalizeSearchQuery(query)
	sanitizedQuery := utils.SanitizeForCacheKey(normalizedQuery)
	cacheKey := fmt.Sprintf("spotify:search:albums:%s", sanitizedQuery)
	
	logger.Debug(ctx, "album search query processing",
		logger.F("original_query", query),
		logger.F("normalized_query", normalizedQuery),
		logger.F("sanitized_query", sanitizedQuery),
		logger.F("cache_key", cacheKey))
	
	if cachedAlbums, found := s.cache.GetSearchAlbums(query); found {
		logger.Debug(ctx, "cache hit for album search",
			logger.F("query", query),
			logger.F("normalized_query", normalizedQuery),
			logger.F("cache_key", cacheKey),
			logger.F("results_count", len(cachedAlbums)))
		return cachedAlbums, nil
	}

	logger.Info(ctx, "cache miss for album search - calling Spotify API",
		logger.F("query", query),
		logger.F("normalized_query", normalizedQuery),
		logger.F("cache_key", cacheKey))

	// Get access token for user
	logger.Debug(ctx, "retrieving access token for album search",
		logger.F("user_id", userID))
	
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		logger.Error(ctx, "failed to get access token for album search",
			logger.F("user_id", userID),
			logger.F("query", query),
			logger.F("error", err))
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	logger.Info(ctx, "calling Spotify API for album search",
		logger.F("user_id", userID),
		logger.F("query", query),
		logger.F("normalized_query", normalizedQuery))
	
	albums, err := s.searchAlbumsFromAPI(ctx, accessToken, query)
	if err != nil {
		logger.Error(ctx, "album search API call failed",
			logger.F("user_id", userID),
			logger.F("query", query),
			logger.F("error", err))
		return nil, err
	}

	s.cache.SetSearchAlbums(query, albums)
	logger.Info(ctx, "album search completed successfully",
		logger.F("user_id", userID),
		logger.F("query", query),
		logger.F("results_count", len(albums)),
		logger.F("cache_key", cacheKey))
	
	logger.Debug(ctx, "cached album search results",
		logger.F("query", query),
		logger.F("results_count", len(albums)))
	return albums, nil
}

// SearchArtists searches for artists with caching
func (s *Spotify) SearchArtists(ctx context.Context, userID, query string) ([]m.ArtistSearch, error) {
	logger.Info(ctx, "artist search initiated",
		logger.F("user_id", userID),
		logger.F("query", query))
	
	if cachedArtists, found := s.cache.GetSearchArtists(query); found {
		logger.Debug(ctx, "cache hit for artist search",
			logger.F("user_id", userID),
			logger.F("query", query),
			logger.F("results_count", len(cachedArtists)))
		return cachedArtists, nil
	}

	logger.Info(ctx, "cache miss for artist search - calling Spotify API",
		logger.F("user_id", userID),
		logger.F("query", query))

	// Get access token for user
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		logger.Error(ctx, "failed to get access token for artist search",
			logger.F("user_id", userID),
			logger.F("error", err))
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	artists, err := s.searchArtistsFromAPI(ctx, accessToken, query)
	if err != nil {
		logger.Error(ctx, "artist search API call failed",
			logger.F("query", query),
			logger.F("error", err))
		return nil, err
	}

	s.cache.SetSearchArtists(query, artists)
	logger.Info(ctx, "artist search completed successfully",
		logger.F("user_id", userID),
		logger.F("query", query),
		logger.F("results_count", len(artists)))
	
	logger.Debug(ctx, "cached artist search results",
		logger.F("query", query),
		logger.F("results_count", len(artists)))
	return artists, nil
}

// SearchPlaylists searches for playlists with caching
func (s *Spotify) SearchPlaylists(ctx context.Context, userID, query string) ([]m.PlaylistSearch, error) {
	logger.Info(ctx, "playlist search initiated",
		logger.F("user_id", userID),
		logger.F("query", query))
	
	if cachedPlaylists, found := s.cache.GetSearchPlaylists(query); found {
		logger.Debug(ctx, "cache hit for playlist search",
			logger.F("user_id", userID),
			logger.F("query", query),
			logger.F("results_count", len(cachedPlaylists)))
		return cachedPlaylists, nil
	}

	logger.Info(ctx, "cache miss for playlist search - calling Spotify API",
		logger.F("user_id", userID),
		logger.F("query", query))

	// Get access token for user
	logger.Debug(ctx, "retrieving access token for playlist search",
		logger.F("user_id", userID))
	
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		logger.Error(ctx, "failed to get access token for playlist search",
			logger.F("user_id", userID),
			logger.F("query", query),
			logger.F("error", err))
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	logger.Info(ctx, "calling Spotify API for playlist search",
		logger.F("user_id", userID),
		logger.F("query", query))
	
	playlists, err := s.searchPlaylistsFromAPI(ctx, accessToken, query)
	if err != nil {
		logger.Error(ctx, "playlist search API call failed",
			logger.F("user_id", userID),
			logger.F("query", query),
			logger.F("error", err))
		return nil, err
	}

	s.cache.SetSearchPlaylists(query, playlists)
	logger.Info(ctx, "playlist search completed successfully",
		logger.F("user_id", userID),
		logger.F("query", query),
		logger.F("results_count", len(playlists)))
	
	return playlists, nil
}

// Private helper methods
func (s *Spotify) search(ctx context.Context, accessToken, limit, atype, query string) ([]byte, error) {
	trackQuery := atype + ":" + strings.ToLower(query)

	logger.Debug(ctx, "building Spotify search request",
		logger.F("search_type", atype),
		logger.F("query", query),
		logger.F("formatted_query", trackQuery),
		logger.F("limit", limit))

	apiURL, err := url.Parse(s.config.GetAPIBaseURL() + "/search")
	if err != nil {
		logger.Error(ctx, "failed to parse Spotify API URL",
			logger.F("error", err))
		return nil, err
	}
	q := apiURL.Query()
	q.Set("type", atype)
	q.Set("q", trackQuery)
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
			ID:            t.ID,
			Name:          t.Name,
			Popularity:    t.Popularity,
			DurationMs:    t.DurationMs,
			Explicit:      t.Explicit,
			PreviewURL:    func() string {
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
			FollowersTotal: 0, // Not available in search response
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
