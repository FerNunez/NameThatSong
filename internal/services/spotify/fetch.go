package spotify

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/FerNunez/NameThatSong/internal/config"
	m "github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/services/cache"
)

// SpotifyFetchService handles fetching individual Spotify entities with caching
type SpotifyFetchService struct {
	config      *config.SpotifyConfig
	cache       cache.SpotifyCache
	authService *SpotifyAuthService
	httpClient  *http.Client
}

// NewSpotifyFetchService creates a new Spotify fetch service
func NewSpotifyFetchService(config *config.SpotifyConfig, cache cache.SpotifyCache, authService *SpotifyAuthService, httpClient *http.Client) *SpotifyFetchService {
	return &SpotifyFetchService{
		config:      config,
		cache:       cache,
		authService: authService,
		httpClient:  httpClient,
	}
}

// FetchTrack fetches a track with caching
func (s *SpotifyFetchService) FetchTrack(ctx context.Context, userID, trackId string) (m.TrackData, error) {
	// Check cache first
	if cachedTrack, found := s.cache.GetTrack(trackId); found {
		return cachedTrack, nil
	}

	// Get access token for user
	accessToken, err := s.authService.GetValidToken(ctx, userID)
	if err != nil {
		return m.TrackData{}, fmt.Errorf("failed to get access token: %v", err)
	}

	// Cache miss - fetch from Spotify API
	track, err := s.fetchTrackFromAPI(accessToken, trackId)
	if err != nil {
		return m.TrackData{}, err
	}

	// Cache the result
	s.cache.SetTrack(trackId, track)
	return track, nil
}

// FetchAlbum fetches an album with caching
func (s *SpotifyFetchService) FetchAlbum(ctx context.Context, userID, albumId string) (m.AlbumData, error) {
	if cachedAlbum, found := s.cache.GetAlbum(albumId); found {
		return cachedAlbum, nil
	}

	// Get access token for user
	accessToken, err := s.authService.GetValidToken(ctx, userID)
	if err != nil {
		return m.AlbumData{}, fmt.Errorf("failed to get access token: %v", err)
	}

	album, err := s.fetchAlbumFromAPI(accessToken, albumId)
	if err != nil {
		return m.AlbumData{}, err
	}

	s.cache.SetAlbum(albumId, album)
	return album, nil
}

// FetchArtist fetches an artist with caching
func (s *SpotifyFetchService) FetchArtist(ctx context.Context, userID, artistId string) (m.ArtistData, error) {
	if cachedArtist, found := s.cache.GetArtist(artistId); found {
		return cachedArtist, nil
	}

	// Get access token for user
	accessToken, err := s.authService.GetValidToken(ctx, userID)
	if err != nil {
		return m.ArtistData{}, fmt.Errorf("failed to get access token: %v", err)
	}

	artist, err := s.fetchArtistFromAPI(accessToken, artistId)
	if err != nil {
		return m.ArtistData{}, err
	}

	s.cache.SetArtist(artistId, artist)
	return artist, nil
}

// FetchPlaylist fetches a playlist with caching
func (s *SpotifyFetchService) FetchPlaylist(ctx context.Context, userID, playlistId string) (m.PlaylistData, error) {
	if cachedPlaylist, found := s.cache.GetPlaylist(playlistId); found {
		return cachedPlaylist, nil
	}

	// Get access token for user
	accessToken, err := s.authService.GetValidToken(ctx, userID)
	if err != nil {
		return m.PlaylistData{}, fmt.Errorf("failed to get access token: %v", err)
	}

	playlist, err := s.fetchPlaylistFromAPI(accessToken, playlistId)
	if err != nil {
		return m.PlaylistData{}, err
	}

	s.cache.SetPlaylist(playlistId, playlist)
	return playlist, nil
}

// Private API fetch methods
func (s *SpotifyFetchService) fetchTrackFromAPI(accessToken, trackId string) (m.TrackData, error) {
	fmt.Println("[SpotifyFetchService] FetchTrack: trackId:", trackId)
	requestURL := fmt.Sprintf("%s/tracks/%s", s.config.GetAPIBaseURL(), trackId)
	req, err := http.NewRequest("GET", requestURL, nil)
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

	return m.TrackData{
		DiscNumber:  fetchTrackResponse.DiscNumber,
		DurationMs:  fetchTrackResponse.DurationMs,
		ID:          trackId,
		Name:        fetchTrackResponse.Name,
		TrackNumber: fetchTrackResponse.TrackNumber,
		Popularity:  fetchTrackResponse.Popularity,
		Explicit:    fetchTrackResponse.Explicit,
	}, nil
}

// Fetch album by ID: retireves all tracks of the album too
func (s *SpotifyFetchService) fetchAlbumFromAPI(accessToken, albumId string) (m.AlbumData, error) {
	requestURL := fmt.Sprintf("%s/albums/%s", s.config.GetAPIBaseURL(), albumId)
	req, err := http.NewRequest("GET", requestURL, nil)
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

	return m.AlbumData{
		AlbumType:   fetchAlbumResponse.AlbumType,
		TotalTracks: fetchAlbumResponse.TotalTracks,
		ID:          albumId,
		ImagesURL:   imageUrl,
		Name:        fetchAlbumResponse.Name,
		ReleaseDate: fetchAlbumResponse.ReleaseDate,
	}, nil
}

func (s *SpotifyFetchService) fetchArtistFromAPI(accessToken, artistId string) (m.ArtistData, error) {
	requestURL := fmt.Sprintf("%s/artists/%s", s.config.GetAPIBaseURL(), artistId)
	req, err := http.NewRequest("GET", requestURL, nil)
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
		Id:         artistId,
		Name:       fetchArtistResponse.Name,
		ImageUrl:   imageUrl,
		Popularity: fetchArtistResponse.Popularity,
	}, nil
}

func (s *SpotifyFetchService) fetchPlaylistFromAPI(accessToken, playlistId string) (m.PlaylistData, error) {
	baseURL := s.config.GetAPIBaseURL() + "/playlists/" + playlistId
	u, err := url.Parse(baseURL)
	if err != nil {
		return m.PlaylistData{}, err
	}
	q := u.Query()
	q.Set("fields", "id,name,description,public,followers(total),images(url)")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return m.PlaylistData{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return m.PlaylistData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return m.PlaylistData{}, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	type FetchPlaylistResponse struct {
		Description string `json:"description"`
		Followers   struct {
			Href  any `json:"href"`
			Total int `json:"total"`
		} `json:"followers"`
		ID     string `json:"id"`
		Images []struct {
			Height any    `json:"height"`
			URL    string `json:"url"`
			Width  any    `json:"width"`
		} `json:"images"`
		Name   string `json:"name"`
		Public bool   `json:"public"`
		Tracks struct {
			Items []struct {
				Track struct {
					PreviewURL       any      `json:"preview_url"`
					AvailableMarkets []string `json:"available_markets"`
					Explicit         bool     `json:"explicit"`
					Type             string   `json:"type"`
					Episode          bool     `json:"episode"`
					Track            bool     `json:"track"`
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
						ExternalUrls         struct {
							Spotify string `json:"spotify"`
						} `json:"external_urls"`
						TotalTracks int `json:"total_tracks"`
					} `json:"album"`
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
			} `json:"items"`
		} `json:"tracks"`
	}

	var fetchPlaylistResponse FetchPlaylistResponse
	if err := json.NewDecoder(resp.Body).Decode(&fetchPlaylistResponse); err != nil {
		return m.PlaylistData{}, err
	}

	imageUrl := ""
	if len(fetchPlaylistResponse.Images) > 0 {
		imageUrl = fetchPlaylistResponse.Images[len(fetchPlaylistResponse.Images)-1].URL
	}

	return m.PlaylistData{
		Description:    fetchPlaylistResponse.Description,
		FollowersTotal: fetchPlaylistResponse.Followers.Total,
		ID:             playlistId,
		ImageUrl:       imageUrl,
		Name:           fetchPlaylistResponse.Name,
		Public:         fetchPlaylistResponse.Public,
		TotalTracks:    0, // Will be set when tracks are fetched separately
	}, nil
}

// FetchAlbumsFromArtist fetches album IDs from an artist
func (s *SpotifyFetchService) FetchAlbumsFromArtist(ctx context.Context, userID, artistId string) ([]string, error) {
	// Get access token for user
	accessToken, err := s.authService.GetValidToken(ctx, userID)
	if err != nil {
		return []string{}, fmt.Errorf("failed to get access token: %v", err)
	}

	limit := 50
	include_groups := "album"

	requestURL := fmt.Sprintf("%s/artists/%v/albums?include_groups=%v&limit=%v", s.config.GetAPIBaseURL(), artistId, include_groups, limit)
	req, err := http.NewRequest("GET", requestURL, nil)
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

	albumIds := make([]string, len(albumsFromArtistResponse.Items))
	for idx, album := range albumsFromArtistResponse.Items {
		albumIds[idx] = album.ID
	}
	return albumIds, nil
}

// FetchTracksFromAlbum fetches track IDs from an album
func (s *SpotifyFetchService) FetchTracksFromAlbum(ctx context.Context, userID, albumId string) ([]string, error) {
	// Get access token for user
	accessToken, err := s.authService.GetValidToken(ctx, userID)
	if err != nil {
		return []string{}, fmt.Errorf("failed to get access token: %v", err)
	}

	limit := 50
	requestURL := fmt.Sprintf("%s/albums/%v/tracks?&limit=%v", s.config.GetAPIBaseURL(), albumId, limit)
	req, err := http.NewRequest("GET", requestURL, nil)
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

	trackIds := make([]string, len(tracksFromAlbumResponse.Items))
	for idx, track := range tracksFromAlbumResponse.Items {
		trackIds[idx] = track.ID
	}

	return trackIds, nil
}

// FetchTracksFromPlaylist fetches track IDs from a playlist
func (s *SpotifyFetchService) FetchTracksFromPlaylist(ctx context.Context, userID, playlistId string) ([]string, error) {
	// Get access token for user
	accessToken, err := s.authService.GetValidToken(ctx, userID)
	if err != nil {
		return []string{}, fmt.Errorf("failed to get access token: %v", err)
	}

	limit := 50
	fields := "items(track(id))"
	requestURL := fmt.Sprintf("%s/playlists/%v/tracks?&fields=%v&limit=%v", s.config.GetAPIBaseURL(), playlistId, fields, limit)
	req, err := http.NewRequest("GET", requestURL, nil)
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
	trackIds := make([]string, len(tracksFromPlaylistResponse.Items))
	for idx, item := range tracksFromPlaylistResponse.Items {
		trackIds[idx] = item.Track.ID
	}
	return trackIds, nil
}

// ////////////////////
// func  FetchAlbumByArtistID(accesToken, artistId string) ([]m.AlbumData, error) {
//
// 	limit := 50
// 	include_groups := "album"
// 	requestURL := fmt.Sprintf("https://api.spotify.com/v1/artists/%s/albums?limit=%d&include_groups=%v", artistId, limit, include_groups)
// 	req, err := http.NewRequest("GET", requestURL, nil)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accesToken))
//
// 	client := http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()
//
// 	if resp.StatusCode != http.StatusOK {
// 		return nil, errors.New(fmt.Sprintf("bad status code:%v", resp.StatusCode))
//
// 	}
//
// 	type FetchAlbumByArtistIDResponse struct {
// 		Href     string `json:"href"`
// 		Limit    int    `json:"limit"`
// 		Next     string `json:"next"`
// 		Offset   int    `json:"offset"`
// 		Previous any    `json:"previous"`
// 		Total    int    `json:"total"`
// 		Items    []struct {
// 			AlbumType        string   `json:"album_type"`
// 			TotalTracks      int      `json:"total_tracks"`
// 			AvailableMarkets []string `json:"available_markets"`
// 			ExternalUrls     struct {
// 				Spotify string `json:"spotify"`
// 			} `json:"external_urls"`
// 			Href   string `json:"href"`
// 			ID     string `json:"id"`
// 			Images []struct {
// 				URL    string `json:"url"`
// 				Height int    `json:"height"`
// 				Width  int    `json:"width"`
// 			} `json:"images"`
// 			Name                 string `json:"name"`
// 			ReleaseDate          string `json:"release_date"`
// 			ReleaseDatePrecision string `json:"release_date_precision"`
// 			Type                 string `json:"type"`
// 			URI                  string `json:"uri"`
// 			Artists              []struct {
// 				ExternalUrls struct {
// 					Spotify string `json:"spotify"`
// 				} `json:"external_urls"`
// 				Href string `json:"href"`
// 				ID   string `json:"id"`
// 				Name string `json:"name"`
// 				Type string `json:"type"`
// 				URI  string `json:"uri"`
// 			} `json:"artists"`
// 			AlbumGroup string `json:"album_group"`
// 		} `json:"items"`
// 	}
//
// 	var albumsResponse FetchAlbumByArtistIDResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&albumsResponse); err != nil {
// 		return nil, err
// 	}
//
// 	AlbumList := make([]m.AlbumData, 0, len(albumsResponse.Items))
// 	for _, item := range albumsResponse.Items {
// 		if item.AlbumType != "album" {
// 			continue
// 		}
// 		if len(item.Images) <= 0 {
// 			panic("spotify response of album with no images")
// 		}
//
// 		album := m.AlbumData{
// 			AlbumType:   item.AlbumType,
// 			TotalTracks: item.TotalTracks,
// 			ID:          item.ID,
// 			ImagesURL:   item.Images[0].URL,
// 			Name:        item.Name,
// 			ReleaseDate: item.ReleaseDate,
// 		}
// 		AlbumList = append(AlbumList, album)
// 	}
// 	return AlbumList, nil
// }
//
// // https://api.spotify.com/v1/albums/{id}/tracks
// // id= 4aawyAB9vmqN3uQ7FjRGTy
// // limit =50
// // fetch artists by ID:
// func  FetchTracksByAlbumID(accessToken, albumId string) ([]m.TrackData, error) {
// 	limit := 50
// 	requestURL := fmt.Sprintf("https://api.spotify.com/v1/albums/%s/tracks?limit=%d", albumId, limit)
// 	req, err := http.NewRequest("GET", requestURL, nil)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))
//
// 	client := http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()
//
// 	if resp.StatusCode != http.StatusOK {
// 		return nil, errors.New(fmt.Sprintf("bad status code:%v", resp.StatusCode))
//
// 	}
//
// 	type FetchTracksByAlbumIDResponse struct {
// 		Href  string `json:"href"`
// 		Items []struct {
// 			Artists []struct {
// 				ExternalUrls struct {
// 					Spotify string `json:"spotify"`
// 				} `json:"external_urls"`
// 				Href string `json:"href"`
// 				ID   string `json:"id"`
// 				Name string `json:"name"`
// 				Type string `json:"type"`
// 				URI  string `json:"uri"`
// 			} `json:"artists"`
// 			AvailableMarkets []string `json:"available_markets"`
// 			DiscNumber       int      `json:"disc_number"`
// 			DurationMs       int      `json:"duration_ms"`
// 			Explicit         bool     `json:"explicit"`
// 			ExternalUrls     struct {
// 				Spotify string `json:"spotify"`
// 			} `json:"external_urls"`
// 			Href        string `json:"href"`
// 			ID          string `json:"id"`
// 			Name        string `json:"name"`
// 			PreviewURL  any    `json:"preview_url"`
// 			TrackNumber int    `json:"track_number"`
// 			Type        string `json:"type"`
// 			URI         string `json:"uri"`
// 			IsLocal     bool   `json:"is_local"`
// 		} `json:"items"`
// 		Limit    int `json:"limit"`
// 		Next     any `json:"next"`
// 		Offset   int `json:"offset"`
// 		Previous any `json:"previous"`
// 		Total    int `json:"total"`
// 	}
//
// 	var tracksResponse FetchTracksByAlbumIDResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&tracksResponse); err != nil {
// 		return nil, err
// 	}
//
// 	trackList := make([]m.TrackData, 0, len(tracksResponse.Items))
// 	for _, item := range tracksResponse.Items {
//
// 		track := m.TrackData{
// 			DiscNumber:  item.DiscNumber,
// 			DurationMs:  item.DurationMs,
// 			ID:          item.ID,
// 			Name:        item.Name,
// 			TrackNumber: item.TrackNumber,
// 		}
// 		trackList = append(trackList, track)
// 	}
// 	return trackList, nil
// }
//
// // Creates an album with the top tracks of an artist
// func  CreateTopTracksAlbum(accessToken, artistId string) (m.AlbumData, []m.TrackData, error) {
//
// 	requestURL := fmt.Sprintf("https://api.spotify.com/v1/artists/%v/top-tracks", artistId)
// 	req, err := http.NewRequest("GET", requestURL, nil)
// 	if err != nil {
// 		return m.AlbumData{}, nil, err
// 	}
//
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))
//
// 	client := http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return m.AlbumData{}, nil, err
// 	}
// 	defer resp.Body.Close()
//
// 	if resp.StatusCode != http.StatusOK {
// 		return m.AlbumData{}, nil, errors.New(fmt.Sprintf("bad status code:%v", resp.StatusCode))
//
// 	}
//
// 	type TopTracks struct {
// 		Tracks []struct {
// 			Album struct {
// 				AlbumType string `json:"album_type"`
// 				Artists   []struct {
// 					ExternalUrls struct {
// 						Spotify string `json:"spotify"`
// 					} `json:"external_urls"`
// 					Href string `json:"href"`
// 					ID   string `json:"id"`
// 					Name string `json:"name"`
// 					Type string `json:"type"`
// 					URI  string `json:"uri"`
// 				} `json:"artists"`
// 				AvailableMarkets []string `json:"available_markets"`
// 				ExternalUrls     struct {
// 					Spotify string `json:"spotify"`
// 				} `json:"external_urls"`
// 				Href   string `json:"href"`
// 				ID     string `json:"id"`
// 				Images []struct {
// 					URL    string `json:"url"`
// 					Height int    `json:"height"`
// 					Width  int    `json:"width"`
// 				} `json:"images"`
// 				IsPlayable           bool   `json:"is_playable"`
// 				Name                 string `json:"name"`
// 				ReleaseDate          string `json:"release_date"`
// 				ReleaseDatePrecision string `json:"release_date_precision"`
// 				TotalTracks          int    `json:"total_tracks"`
// 				Type                 string `json:"type"`
// 				URI                  string `json:"uri"`
// 			} `json:"album"`
// 			Artists []struct {
// 				ExternalUrls struct {
// 					Spotify string `json:"spotify"`
// 				} `json:"external_urls"`
// 				Href string `json:"href"`
// 				ID   string `json:"id"`
// 				Name string `json:"name"`
// 				Type string `json:"type"`
// 				URI  string `json:"uri"`
// 			} `json:"artists"`
// 			AvailableMarkets []string `json:"available_markets"`
// 			DiscNumber       int      `json:"disc_number"`
// 			DurationMs       int      `json:"duration_ms"`
// 			Explicit         bool     `json:"explicit"`
// 			ExternalIds      struct {
// 				Isrc string `json:"isrc"`
// 			} `json:"external_ids"`
// 			ExternalUrls struct {
// 				Spotify string `json:"spotify"`
// 			} `json:"external_urls"`
// 			Href        string `json:"href"`
// 			ID          string `json:"id"`
// 			IsLocal     bool   `json:"is_local"`
// 			IsPlayable  bool   `json:"is_playable"`
// 			Name        string `json:"name"`
// 			Popularity  int    `json:"popularity"`
// 			PreviewURL  any    `json:"preview_url"`
// 			TrackNumber int    `json:"track_number"`
// 			Type        string `json:"type"`
// 			URI         string `json:"uri"`
// 		} `json:"tracks"`
// 	}
//
// 	var tt TopTracks
// 	if err := json.NewDecoder(resp.Body).Decode(&tt); err != nil {
// 		return m.AlbumData{}, nil, err
// 	}
//
// 	trackList := make([]m.TrackData, 0, len(tt.Tracks))
// 	for _, item := range tt.Tracks {
//
// 		track := m.TrackData{
// 			DiscNumber:  item.DiscNumber,
// 			DurationMs:  item.DurationMs,
// 			ID:          item.ID,
// 			Name:        item.Name,
// 			TrackNumber: item.TrackNumber,
// 		}
// 		trackList = append(trackList, track)
// 	}
//
// 	fakeID := "topTracks" + artistId
// 	album := m.AlbumData{
// 		AlbumType:   "TopTracks",
// 		TotalTracks: len(trackList),
// 		ID:          fakeID,
// 		ImagesURL:   "",
// 		Name:        fmt.Sprintf("Top%v", len(trackList)),
// 		ReleaseDate: "NEW",
// 	}
// 	//fmt.Println("Fake album: ", album.Name, "trakcs: ", album.TotalTracks, "ID:", album.ID)
// 	return album, trackList, nil
// }
