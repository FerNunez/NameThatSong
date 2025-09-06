package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	m "github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
)

// =============================================================================
// FETCH METHODS
// =============================================================================

// FetchTrack fetches a track using three-tier strategy: Cache → Database → API
func (s *Spotify) FetchTrack(ctx context.Context, userID string, trackID m.SpotifyID) (m.TrackData, error) {
	// GET
	// Tier 1: Check Redis cache first
	if cachedTrack, err := s.cache.GetTrack(trackID); err == nil {
		return cachedTrack, nil
	}
	logger.Debug(ctx, "couldnt find track in cache", logger.F("track_id", trackID))
	// Tier 2: Check database for persistent storage
	if dbTrack, err := s.dataStore.GetTrack(ctx, string(trackID)); err == nil && dbTrack != nil {
		// Found in database, update cache and return
		s.cache.SetTrack(trackID, *dbTrack)
		return *dbTrack, nil
	}
	logger.Debug(ctx, "couldnt find track in db", logger.F("track_id", trackID))
	// Tier 3: Fetch from Spotify API as last resort
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		return m.TrackData{}, fmt.Errorf("failed to get access token: %v", err)
	}

	track, err := s.fetchTrackFromAPI(ctx, accessToken, trackID)
	if err != nil {
		return m.TrackData{}, err
	}

	// STORE
	// go func() { // 	// Async removed cause track db is needed for playlit relations :(
	if err := s.dataStore.StoreTrack(ctx, &track); err != nil {
		// Log error but don't fail the request
		logger.Warn(ctx, "Failed to store track in db", logger.F("track_id", string(trackID)), logger.F("err", err))
	}
	// }()

	// Cache the result for fast future access
	go func() {
		if err := s.cache.SetTrack(trackID, track); err != nil {
			logger.Warn(ctx, "Failed to store track in cache", logger.F("track_id", string(trackID)), logger.F("err", err))
		}
	}()
	return track, nil
}

// FetchMultipleTracks fetches multiple tracks using three-tier strategy with batch processing
func (s *Spotify) FetchMultipleTracks(ctx context.Context, userID string, trackIDs []m.SpotifyID) ([]m.TrackData, error) {
	if len(trackIDs) == 0 {
		return []m.TrackData{}, nil
	}
	// Remove duplicates to avoid unnecessary processing
	uniqueIDs := removeDuplicates(trackIDs)

	remaining := uniqueIDs
	results := make([]m.TrackData, 0, len(uniqueIDs))
	// Tier 1: Check Redis cache with batch operation
	cachedTracks, stillMissing := s.cache.GetMultipleTracks(remaining)
	for _, track := range cachedTracks {
		results = append(results, track)
	}
	remaining = stillMissing

	// All fetched
	if len(remaining) == 0 {
		return results, nil
	}

	// Tier 2: Check database with batch operation
	dbTracks, stillMissing, err := s.dataStore.GetMultipleTracks(ctx, remaining)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tracks from database: %w", err)
	}

	// Update cache with database results and add to results
	cacheMap := make(map[m.SpotifyID]m.TrackData)
	for trackID, track := range dbTracks {
		results = append(results, *track)
		cacheMap[m.SpotifyID(trackID)] = *track
	}
	if len(cacheMap) > 0 {
		s.cache.SetMultipleTracks(cacheMap)
	}
	remaining = stillMissing

	if len(remaining) == 0 {
		return results, nil
	}

	// Tier 3: Fetch remaining tracks from Spotify API in batches of 50
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	apiTracks, err := s.fetchMultipleTracksFromAPI(ctx, accessToken, remaining)
	if err != nil {
		// If we have some tracks from cache/database, return partial success
		if len(results) > 0 {
			fmt.Printf("Warning: API fetch failed but returning %d cached/database tracks: %v\n", len(results), err)
			return results, nil
		}
		// If we have no tracks at all, return the error
		return nil, fmt.Errorf("failed to fetch tracks from API: %w", err)
	}

	// Store API results in database and cache
	if len(apiTracks) > 0 {
		trackPointers := make([]*m.TrackData, len(apiTracks))
		cacheMap := make(map[m.SpotifyID]m.TrackData)
		for i, track := range apiTracks {
			trackPointers[i] = &track
			cacheMap[m.SpotifyID(track.ID)] = track
		}

		// Store in database (async to avoid blocking)
		go func() {
			if err := s.dataStore.StoreMultipleTracks(context.Background(), trackPointers); err != nil {
				fmt.Printf("Warning: failed to store batch tracks in database: %v\n", err)
			}
		}()

		// Cache the results for fast future access
		s.cache.SetMultipleTracks(cacheMap)

		results = append(results, apiTracks...)
	}

	return results, nil
}

// FetchAlbum fetches an album using three-tier strategy: Cache → Database → API
func (s *Spotify) FetchAlbum(ctx context.Context, userID string, albumID m.SpotifyID) (m.AlbumData, error) {
	// Tier 1: Check Redis cache first
	if cachedAlbum, found := s.cache.GetAlbum(albumID); found {
		return cachedAlbum, nil
	}

	// Tier 2: Check database for persistent storage
	if dbAlbum, err := s.dataStore.GetAlbum(ctx, string(albumID)); err == nil && dbAlbum != nil {
		// Found in database, update cache and return
		s.cache.SetAlbum(albumID, *dbAlbum)
		return *dbAlbum, nil
	}

	// Tier 3: Fetch from Spotify API as last resort
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		return m.AlbumData{}, fmt.Errorf("failed to get access token: %v", err)
	}

	album, err := s.fetchAlbumFromAPI(ctx, accessToken, albumID)
	if err != nil {
		return m.AlbumData{}, err
	}

	// Store in database for persistence (async to avoid blocking)
	go func() {
		if err := s.dataStore.StoreAlbum(context.Background(), &album); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: failed to store album %s in database: %v\n", string(albumID), err)
		}
	}()

	// Cache the result for fast future access
	s.cache.SetAlbum(albumID, album)
	return album, nil
}

// FetchArtist fetches an artist using three-tier strategy: Cache → Database → API
func (s *Spotify) FetchArtist(ctx context.Context, userID string, artistID m.SpotifyID) (m.ArtistData, error) {
	// Tier 1: Check Redis cache first
	if cachedArtist, found := s.cache.GetArtist(artistID); found {
		return cachedArtist, nil
	}

	// Tier 2: Check database for persistent storage
	if dbArtist, err := s.dataStore.GetArtist(ctx, string(artistID)); err == nil && dbArtist != nil {
		// Found in database, update cache and return
		s.cache.SetArtist(artistID, *dbArtist)
		return *dbArtist, nil
	}

	// Tier 3: Fetch from Spotify API as last resort
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		return m.ArtistData{}, fmt.Errorf("failed to get access token: %v", err)
	}

	artist, err := s.fetchArtistFromAPI(ctx, accessToken, artistID)
	if err != nil {
		return m.ArtistData{}, err
	}

	// Store in database for persistence (async to avoid blocking)
	go func() {
		if err := s.dataStore.StoreArtist(context.Background(), &artist); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: failed to store artist %s in database: %v\n", string(artistID), err)
		}
	}()

	// Cache the result for fast future access
	s.cache.SetArtist(artistID, artist)
	return artist, nil
}

// FetchPlaylist fetches a playlist using three-tier strategy: Cache → Database → API
func (s *Spotify) FetchPlaylist(ctx context.Context, userID string, playlistID m.SpotifyID) (m.PlaylistData, []string, []string, error) {
	// // Tier 1: Check Redis cache first
	if cachedPlaylist, found := s.cache.GetPlaylist(playlistID); found {
		trackIDs, _ := s.cache.GetPlaylistTracks(string(playlistID))
		albumIDs, _ := s.cache.GetPlaylistAlbums(string(playlistID))
		return cachedPlaylist, trackIDs, albumIDs, nil
	}

	// Tier 2: Check database for persistent storage
	if dbPlaylist, err := s.dataStore.GetPlaylist(ctx, string(playlistID)); err == nil && dbPlaylist != nil {
		trackIDs, _ := s.dataStore.GetPlaylistTracks(ctx, string(playlistID))

		albumIDs := make([]string, 0, len(trackIDs))
		for _, trackID := range trackIDs {
			albumID, err := s.dataStore.GetAlbumByTrackID(ctx, trackID)
			if err != nil {
				return m.PlaylistData{}, []string{}, []string{}, err
			}
			albumIDs = append(albumIDs, albumID)
		}

		// Found in database, update cache and return
		s.cache.SetPlaylist(playlistID, *dbPlaylist)
		s.cache.SetPlaylistTracks(string(playlistID), trackIDs)
		s.cache.SetPlaylistAlbums(string(playlistID), albumIDs)
		return *dbPlaylist, trackIDs, albumIDs, nil
	}
	//
	// Tier 3: Fetch from Spotify API as last resort
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		return m.PlaylistData{}, []string{}, []string{}, fmt.Errorf("failed to get access token: %v", err)
	}

	// Fetch from API
	playlist, trackIDs, albumIDs, err := s.fetchPlaylistFromAPI(ctx, accessToken, playlistID)
	if err != nil {
		return m.PlaylistData{}, []string{}, []string{}, err
	}

	// Convert string track IDs to m.SpotifyID for batch fetching
	spotifyTrackIDs := make([]m.SpotifyID, len(trackIDs))
	for i, id := range trackIDs {
		spotifyTrackIDs[i] = m.SpotifyID(id)
	}
	// Use batch fetching for tracks for better performance
	_, err = s.FetchMultipleTracks(ctx, userID, spotifyTrackIDs)
	if err != nil {
		return m.PlaylistData{}, []string{}, []string{}, err
	}

	// Albums still need individual fetching since we don't have batch album fetching yet
	for _, albumID := range albumIDs {
		_, err := s.FetchAlbum(ctx, userID, m.SpotifyID(albumID))
		if err != nil {
			return m.PlaylistData{}, []string{}, []string{}, err
		}
	}

	// Store in database for persistence (async to avoid blocking)
	go func() {
		if err := s.dataStore.StorePlaylist(context.Background(), &playlist); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: failed to store playlist %s in database: %v\n", playlistID, err)
		}

		if err := s.dataStore.UpsertPlaylistTracks(context.Background(), string(playlistID), trackIDs); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Warning: failed to store playlist tracks %s in database: %v\n", playlistID, err)
		}

		for idx, albumID := range albumIDs {
			if err := s.dataStore.UpsertAlbumTracks(context.Background(), albumID, []string{trackIDs[idx]}); err != nil {
				// Log error but don't fail the request
				fmt.Printf("Warning: failed to store playlist tracks %s in database: %v\n", playlistID, err)
			}
		}
	}()

	// Cache the result for fast future access
	s.cache.SetPlaylist(playlistID, playlist)
	s.cache.SetPlaylistTracks(string(playlistID), trackIDs)
	s.cache.SetPlaylistAlbums(string(playlistID), albumIDs)
	return playlist, trackIDs, albumIDs, nil
}

// Private API fetch methods
func (s *Spotify) fetchTrackFromAPI(ctx context.Context, accessToken string, trackID m.SpotifyID) (m.TrackData, error) {
	fmt.Println("[SpotifyFetchService] FetchTrack: trackID:", trackID)
	requestURL := fmt.Sprintf("%s/tracks/%s", s.config.GetAPIBaseURL(), trackID)
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return m.TrackData{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return m.TrackData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return m.TrackData{}, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	type FetchTrackResponse struct {
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
			AvailableMarkets []any `json:"available_markets"`
			ExternalUrls     struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href   string `json:"href"`
			ID     string `json:"id"`
			Images []struct {
				URL    string `json:"url"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"images"`
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
		AvailableMarkets []any `json:"available_markets"`
		DiscNumber       int   `json:"disc_number"`
		DurationMs       int   `json:"duration_ms"`
		Explicit         bool  `json:"explicit"`
		ExternalIds      struct {
			Isrc string `json:"isrc"`
		} `json:"external_ids"`
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href        string `json:"href"`
		ID          string `json:"id"`
		IsLocal     bool   `json:"is_local"`
		Name        string `json:"name"`
		Popularity  int    `json:"popularity"`
		PreviewURL  any    `json:"preview_url"`
		TrackNumber int    `json:"track_number"`
		Type        string `json:"type"`
		URI         string `json:"uri"`
	}

	var fetchTrackResponse FetchTrackResponse
	if err := json.NewDecoder(resp.Body).Decode(&fetchTrackResponse); err != nil {
		return m.TrackData{}, err
	}

	// Extract track artists IDs
	trackArtistIDs := make([]m.SpotifyID, len(fetchTrackResponse.Artists))
	for i, artist := range fetchTrackResponse.Artists {
		trackArtistIDs[i] = m.SpotifyID(artist.ID)
	}

	return m.TrackData{
		ID:          trackID,
		Name:        fetchTrackResponse.Name,
		DurationMs:  fetchTrackResponse.DurationMs,
		DiscNumber:  fetchTrackResponse.DiscNumber,
		TrackNumber: fetchTrackResponse.TrackNumber,
		Popularity:  fetchTrackResponse.Popularity,
		Explicit:    fetchTrackResponse.Explicit,
		IsLocal:     fetchTrackResponse.IsLocal,
		AlbumID:     m.SpotifyID(fetchTrackResponse.Album.ID),
		ArtistIDs:   trackArtistIDs,
		CachedAt:    time.Now(),
	}, nil
}

// Fetch album by ID: retireves all tracks of the album too
func (s *Spotify) fetchAlbumFromAPI(ctx context.Context, accessToken string, albumID m.SpotifyID) (m.AlbumData, error) {
	requestURL := fmt.Sprintf("%s/albums/%s", s.config.GetAPIBaseURL(), albumID)
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return m.AlbumData{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return m.AlbumData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return m.AlbumData{}, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	type FetchAlbumResponse struct {
		AlbumType        string   `json:"album_type"`
		TotalTracks      int      `json:"total_tracks"`
		AvailableMarkets []string `json:"available_markets"`
		ExternalUrls     struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Href   string `json:"href"`
		ID     string `json:"id"`
		Images []struct {
			URL    string `json:"url"`
			Height int    `json:"height"`
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
		Tracks struct {
			Href     string `json:"href"`
			Limit    int    `json:"limit"`
			Next     any    `json:"next"`
			Offset   int    `json:"offset"`
			Previous any    `json:"previous"`
			Total    int    `json:"total"`
			Items    []struct {
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
				ExternalUrls     struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href        string `json:"href"`
				ID          string `json:"id"`
				Name        string `json:"name"`
				PreviewURL  any    `json:"preview_url"`
				TrackNumber int    `json:"track_number"`
				Type        string `json:"type"`
				URI         string `json:"uri"`
				IsLocal     bool   `json:"is_local"`
			} `json:"items"`
		} `json:"tracks"`
		Copyrights []struct {
			Text string `json:"text"`
			Type string `json:"type"`
		} `json:"copyrights"`
		ExternalIds struct {
			Upc string `json:"upc"`
		} `json:"external_ids"`
		Genres     []any  `json:"genres"`
		Label      string `json:"label"`
		Popularity int    `json:"popularity"`
	}

	var fetchAlbumResponse FetchAlbumResponse
	if err := json.NewDecoder(resp.Body).Decode(&fetchAlbumResponse); err != nil {
		return m.AlbumData{}, err
	}

	imageUrl := ""
	if len(fetchAlbumResponse.Images) > 0 {
		imageUrl = fetchAlbumResponse.Images[len(fetchAlbumResponse.Images)-1].URL
	}

	// Extract artists from album data
	albumArtists := make([]m.ArtistData, len(fetchAlbumResponse.Artists))
	for i, artist := range fetchAlbumResponse.Artists {
		albumArtists[i] = m.ArtistData{
			ID:       artist.ID,
			Name:     artist.Name,
			ImageURL: "", // Not available in album artists response
		}
	}

	return m.AlbumData{
		ID:                   albumID,
		Name:                 fetchAlbumResponse.Name,
		AlbumType:            fetchAlbumResponse.AlbumType,
		ReleaseDate:          fetchAlbumResponse.ReleaseDate,
		ReleaseDatePrecision: fetchAlbumResponse.ReleaseDatePrecision,
		TotalTracks:          fetchAlbumResponse.TotalTracks,
		ImageURL:             imageUrl,
		Label:                fetchAlbumResponse.Label,
		Popularity:           fetchAlbumResponse.Popularity,
		Artists:              albumArtists,
		CachedAt:             time.Now(),
	}, nil
}

func (s *Spotify) fetchArtistFromAPI(ctx context.Context, accessToken string, artistID m.SpotifyID) (m.ArtistData, error) {
	requestURL := fmt.Sprintf("%s/artists/%s", s.config.GetAPIBaseURL(), artistID)
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return m.ArtistData{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return m.ArtistData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return m.ArtistData{}, fmt.Errorf("unexpected status code: %v", resp.StatusCode)

	}

	type FetchArtistResponse struct {
		ExternalUrls struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Followers struct {
			Href  string `json:"href"`
			Total int    `json:"total"`
		} `json:"followers"`
		Genres []string `json:"genres"`
		Href   string   `json:"href"`
		ID     string   `json:"id"`
		Images []struct {
			URL    string `json:"url"`
			Height int    `json:"height"`
			Width  int    `json:"width"`
		} `json:"images"`
		Name       string `json:"name"`
		Popularity int    `json:"popularity"`
		Type       string `json:"type"`
		URI        string `json:"uri"`
	}

	var fetchArtistResponse FetchArtistResponse
	if err := json.NewDecoder(resp.Body).Decode(&fetchArtistResponse); err != nil {
		return m.ArtistData{}, err
	}

	imageUrl := ""
	if len(fetchArtistResponse.Images) > 0 {
		imageUrl = fetchArtistResponse.Images[len(fetchArtistResponse.Images)-1].URL
	}

	return m.ArtistData{
		ID:             artistID,
		Name:           fetchArtistResponse.Name,
		ImageURL:       imageUrl,
		Popularity:     fetchArtistResponse.Popularity,
		FollowersTotal: fetchArtistResponse.Followers.Total,
		Genres:         fetchArtistResponse.Genres,
		CachedAt:       time.Now(),
	}, nil
}

func (s *Spotify) fetchPlaylistFromAPI(ctx context.Context, accessToken string, playlistID m.SpotifyID) (m.PlaylistData, []string, []string, error) {
	baseURL := s.config.GetAPIBaseURL() + "/playlists/" + playlistID
	fmt.Println(baseURL)
	u, err := url.Parse(baseURL)
	if err != nil {
		return m.PlaylistData{}, []string{}, []string{}, err
	}
	q := u.Query()
	//q.Set("fields", "id,name,description,public,followers(total),images(url)")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return m.PlaylistData{}, []string{}, []string{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return m.PlaylistData{}, []string{}, []string{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return m.PlaylistData{}, []string{}, []string{}, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	type FetchPlaylistResponse struct {
		Collaborative bool   `json:"collaborative"`
		Description   string `json:"description"`
		ExternalUrls  struct {
			Spotify string `json:"spotify"`
		} `json:"external_urls"`
		Followers struct {
			Href  interface{} `json:"href"`
			Total int         `json:"total"`
		} `json:"followers"`
		Href   string `json:"href"`
		ID     string `json:"id"`
		Images []struct {
			Height int    `json:"height"`
			URL    string `json:"url"`
			Width  int    `json:"width"`
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
		PrimaryColor interface{} `json:"primary_color"`
		Public       bool        `json:"public"`
		SnapshotID   string      `json:"snapshot_id"`
		Tracks       struct {
			Href  string `json:"href"`
			Items []struct {
				AddedAt time.Time `json:"added_at"`
				AddedBy struct {
					ExternalUrls struct {
						Spotify string `json:"spotify"`
					} `json:"external_urls"`
					Href string `json:"href"`
					ID   string `json:"id"`
					Type string `json:"type"`
					URI  string `json:"uri"`
				} `json:"added_by"`
				IsLocal      bool        `json:"is_local"`
				PrimaryColor interface{} `json:"primary_color"`
				Track        struct {
					PreviewURL       interface{} `json:"preview_url"`
					AvailableMarkets []string    `json:"available_markets"`
					Explicit         bool        `json:"explicit"`
					Type             string      `json:"type"`
					Episode          bool        `json:"episode"`
					Track            bool        `json:"track"`
					Album            struct {
						AvailableMarkets []string `json:"available_markets"`
						Type             string   `json:"type"`
						AlbumType        string   `json:"album_type"`
						Href             string   `json:"href"`
						ID               string   `json:"id"`
						Images           []struct {
							URL    string `json:"url"`
							Width  int    `json:"width"`
							Height int    `json:"height"`
						} `json:"images"`
						Name                 string `json:"name"`
						ReleaseDate          string `json:"release_date"`
						ReleaseDatePrecision string `json:"release_date_precision"`
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
						ExternalUrls struct {
							Spotify string `json:"spotify"`
						} `json:"external_urls"`
						TotalTracks int `json:"total_tracks"`
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
					DiscNumber  int `json:"disc_number"`
					TrackNumber int `json:"track_number"`
					DurationMs  int `json:"duration_ms"`
					ExternalIds struct {
						Isrc string `json:"isrc"`
					} `json:"external_ids"`
					ExternalUrls struct {
						Spotify string `json:"spotify"`
					} `json:"external_urls"`
					Href       string `json:"href"`
					ID         string `json:"id"`
					Name       string `json:"name"`
					Popularity int    `json:"popularity"`
					URI        string `json:"uri"`
					IsLocal    bool   `json:"is_local"`
				} `json:"track"`
				VideoThumbnail struct {
					URL interface{} `json:"url"`
				} `json:"video_thumbnail"`
			} `json:"items"`
			Limit    int         `json:"limit"`
			Next     interface{} `json:"next"`
			Offset   int         `json:"offset"`
			Previous interface{} `json:"previous"`
			Total    int         `json:"total"`
		} `json:"tracks"`
		Type string `json:"type"`
		URI  string `json:"uri"`
	}
	var fetchPlaylistResponse FetchPlaylistResponse
	if err := json.NewDecoder(resp.Body).Decode(&fetchPlaylistResponse); err != nil {
		return m.PlaylistData{}, []string{}, []string{}, err
	}

	imageUrl := ""
	if len(fetchPlaylistResponse.Images) > 0 {
		imageUrl = fetchPlaylistResponse.Images[len(fetchPlaylistResponse.Images)-1].URL
	}

	trackIDs := make([]string, 0, len(fetchPlaylistResponse.Tracks.Items))
	for _, item := range fetchPlaylistResponse.Tracks.Items {
		trackIDs = append(trackIDs, item.Track.ID)
	}
	albumIDs := make([]string, 0, len(fetchPlaylistResponse.Tracks.Items))
	for _, item := range fetchPlaylistResponse.Tracks.Items {
		albumIDs = append(albumIDs, item.Track.Album.ID)
	}

	return m.PlaylistData{
		ID:               playlistID,
		Name:             fetchPlaylistResponse.Name,
		Description:      fetchPlaylistResponse.Description,
		OwnerID:          fetchPlaylistResponse.Owner.ID,
		OwnerDisplayName: fetchPlaylistResponse.Owner.DisplayName,
		Public:           fetchPlaylistResponse.Public,
		Collaborative:    fetchPlaylistResponse.Collaborative,
		FollowersTotal:   0,
		TotalTracks:      fetchPlaylistResponse.Tracks.Total,
		ImageURL:         imageUrl,
		CachedAt:         time.Now(),
	}, trackIDs, albumIDs, nil
}

// FetchAlbumsFromArtist fetches album IDs from an artist
func (s *Spotify) FetchAlbumsFromArtist(ctx context.Context, userID string, artistID m.SpotifyID) ([]string, error) {
	// Get access token for user
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		return []string{}, fmt.Errorf("failed to get access token: %v", err)
	}

	limit := 50
	include_groups := "album"

	requestURL := fmt.Sprintf("%s/artists/%v/albums?include_groups=%v&limit=%v", s.config.GetAPIBaseURL(), artistID, include_groups, limit)
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return []string{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return []string{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []string{}, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	type AlbumsFromArtistResponse struct {
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
				URL    string `json:"url"`
				Height int    `json:"height"`
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
			AlbumGroup string `json:"album_group"`
		} `json:"items"`
	}
	var albumsFromArtistResponse AlbumsFromArtistResponse
	if err := json.NewDecoder(resp.Body).Decode(&albumsFromArtistResponse); err != nil {
		return []string{}, nil
	}

	albumIDs := make([]string, len(albumsFromArtistResponse.Items))
	for idx, album := range albumsFromArtistResponse.Items {
		albumIDs[idx] = album.ID
	}
	return albumIDs, nil
}

// FetchTracksFromAlbum fetches track IDs from an album
func (s *Spotify) FetchTracksFromAlbum(ctx context.Context, userID string, albumID m.SpotifyID) ([]string, error) {
	// Get access token for user
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		return []string{}, fmt.Errorf("failed to get access token: %v", err)
	}

	limit := 50
	requestURL := fmt.Sprintf("%s/albums/%v/tracks?&limit=%v", s.config.GetAPIBaseURL(), albumID, limit)
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return []string{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return []string{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []string{}, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	type TracksFromAlbumResponse struct {
		Href     string `json:"href"`
		Limit    int    `json:"limit"`
		Next     string `json:"next"`
		Offset   int    `json:"offset"`
		Previous string `json:"previous"`
		Total    int    `json:"total"`
		Items    []struct {
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
			ExternalUrls     struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href       string `json:"href"`
			ID         string `json:"id"`
			IsPlayable bool   `json:"is_playable"`
			LinkedFrom struct {
				ExternalUrls struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href string `json:"href"`
				ID   string `json:"id"`
				Type string `json:"type"`
				URI  string `json:"uri"`
			} `json:"linked_from"`
			Restrictions struct {
				Reason string `json:"reason"`
			} `json:"restrictions"`
			Name        string `json:"name"`
			PreviewURL  string `json:"preview_url"`
			TrackNumber int    `json:"track_number"`
			Type        string `json:"type"`
			URI         string `json:"uri"`
			IsLocal     bool   `json:"is_local"`
		} `json:"items"`
	}
	var tracksFromAlbumResponse TracksFromAlbumResponse
	if err := json.NewDecoder(resp.Body).Decode(&tracksFromAlbumResponse); err != nil {
		return []string{}, nil
	}

	trackIDs := make([]string, len(tracksFromAlbumResponse.Items))
	for idx, track := range tracksFromAlbumResponse.Items {
		trackIDs[idx] = track.ID
	}

	return trackIDs, nil
}

// fetchMultipleTracksFromAPI fetches multiple tracks from Spotify API in batches of 50
func (s *Spotify) fetchMultipleTracksFromAPI(ctx context.Context, accessToken string, trackIDs []m.SpotifyID) ([]m.TrackData, error) {
	if len(trackIDs) == 0 {
		return []m.TrackData{}, nil
	}

	var allTracks []m.TrackData
	const batchSize = 50 // Spotify API limit for tracks endpoint

	// Process tracks in batches of 50
	for i := 0; i < len(trackIDs); i += batchSize {
		end := min(i+batchSize, len(trackIDs))

		batch := trackIDs[i:end]
		batchTracks, err := s.fetchTrackBatchFromAPI(ctx, accessToken, batch)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch batch %d-%d: %w", i, end-1, err)
		}

		allTracks = append(allTracks, batchTracks...)
	}

	return allTracks, nil
}

// fetchTrackBatchFromAPI fetches a single batch of tracks (up to 50) from Spotify API
func (s *Spotify) fetchTrackBatchFromAPI(ctx context.Context, accessToken string, trackIDs []m.SpotifyID) ([]m.TrackData, error) {
	if len(trackIDs) == 0 {
		return []m.TrackData{}, nil
	}

	// Join track IDs with comma and URL encode
	idsParam := ""
	for i, id := range trackIDs {
		if i > 0 {
			idsParam += ","
		}
		idsParam += url.QueryEscape(string(id))
	}

	requestURL := fmt.Sprintf("%s/tracks?ids=%s", s.config.GetAPIBaseURL(), idsParam)
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	type BatchTracksResponse struct {
		Tracks []*struct {
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
				AvailableMarkets []any `json:"available_markets"`
				ExternalUrls     struct {
					Spotify string `json:"spotify"`
				} `json:"external_urls"`
				Href   string `json:"href"`
				ID     string `json:"id"`
				Images []struct {
					URL    string `json:"url"`
					Width  int    `json:"width"`
					Height int    `json:"height"`
				} `json:"images"`
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
			AvailableMarkets []any `json:"available_markets"`
			DiscNumber       int   `json:"disc_number"`
			DurationMs       int   `json:"duration_ms"`
			Explicit         bool  `json:"explicit"`
			ExternalIds      struct {
				Isrc string `json:"isrc"`
			} `json:"external_ids"`
			ExternalUrls struct {
				Spotify string `json:"spotify"`
			} `json:"external_urls"`
			Href        string      `json:"href"`
			ID          string      `json:"id"`
			IsLocal     bool        `json:"is_local"`
			Name        string      `json:"name"`
			Popularity  int         `json:"popularity"`
			PreviewURL  interface{} `json:"preview_url"`
			TrackNumber int         `json:"track_number"`
			Type        string      `json:"type"`
			URI         string      `json:"uri"`
		} `json:"tracks"`
	}

	var batchResponse BatchTracksResponse
	if err := json.NewDecoder(resp.Body).Decode(&batchResponse); err != nil {
		return nil, err
	}

	tracks := make([]m.TrackData, 0, len(batchResponse.Tracks))
	for _, trackData := range batchResponse.Tracks {
		// Skip null tracks (can happen with invalid IDs)
		if trackData == nil || trackData.ID == "" {
			continue
		}

		// collect artists IDs
		artistDIs := make([]m.SpotifyID, len(trackData.Artists))
		for i, artist := range trackData.Artists {
			artistDIs[i] = m.SpotifyID(artist.ID) // First artist is considered primary
		}
		track := m.TrackData{
			ID:          m.SpotifyID(trackData.ID),
			Name:        trackData.Name,
			DurationMs:  trackData.DurationMs,
			DiscNumber:  trackData.DiscNumber,
			TrackNumber: trackData.TrackNumber,
			Popularity:  trackData.Popularity,
			Explicit:    trackData.Explicit,
			IsLocal:     trackData.IsLocal,
			AlbumID:     m.SpotifyID(trackData.Album.ID),
			ArtistIDs:   artistDIs,
			CachedAt:    time.Now(),
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

// FetchTracksFromPlaylist fetches track IDs from a playlist
func (s *Spotify) FetchTracksFromPlaylist(ctx context.Context, userID string, playlistID m.SpotifyID) ([]string, error) {
	// Get access token for user
	accessToken, err := s.GetValidToken(ctx, userID)
	if err != nil {
		return []string{}, fmt.Errorf("failed to get access token: %v", err)
	}

	limit := 50
	fields := "items(track(id))"
	requestURL := fmt.Sprintf("%s/playlists/%v/tracks?&fields=%v&limit=%v", s.config.GetAPIBaseURL(), playlistID, fields, limit)
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return []string{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return []string{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []string{}, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	type TracksFromPlaylistResponse struct {
		Items []struct {
			Track struct {
				ID string `json:"id"`
			} `json:"track"`
		} `json:"items"`
	}

	var tracksFromPlaylistResponse TracksFromPlaylistResponse
	if err := json.NewDecoder(resp.Body).Decode(&tracksFromPlaylistResponse); err != nil {
		return []string{}, nil
	}
	trackIDs := make([]string, len(tracksFromPlaylistResponse.Items))
	for idx, item := range tracksFromPlaylistResponse.Items {
		trackIDs[idx] = item.Track.ID
	}
	return trackIDs, nil
}

// // Tools
// removeDuplicates removes duplicate track IDs from a slice
func removeDuplicates(trackIDs []m.SpotifyID) []m.SpotifyID {
	if len(trackIDs) == 0 {
		return []m.SpotifyID{}
	}

	seen := make(map[m.SpotifyID]bool)
	unique := make([]m.SpotifyID, 0, len(trackIDs))

	for _, id := range trackIDs {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}

	return unique
}
