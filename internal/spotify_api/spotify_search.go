package spotify_api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

func (p *SpotifySongProvider) SearchTracksByName(accessToken, name string) ([]TrackData, error) {
	limit := "50"
	trackQuery := "track:" + strings.ToLower(name)

	apiURL, err := url.Parse("https://api.spotify.com/v1/search")
	if err != nil {
		return nil, err
	}
	q := apiURL.Query()
	q.Set("type", "track")
	q.Set("q", trackQuery)
	q.Set("limit", limit)
	apiURL.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", apiURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

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
	if err := json.NewDecoder(resp.Body).Decode(&searchTrackResponse); err != nil {
		return nil, err
	}

	// Convert to trackInfo
	tracks := make([]TrackData, 0, len(searchTrackResponse.Tracks.Items))
	for _, t := range searchTrackResponse.Tracks.Items {
		trackInfo := TrackData{
			DiscNumber:  t.DiscNumber,
			DurationMs:  t.DurationMs,
			ID:          t.ID,
			Name:        t.Name,
			TrackNumber: t.TrackNumber,
			//Popularity: 		t.Popularity,
		}
		tracks = append(tracks, trackInfo)
	}
	// sort.Slice(tracks, func(i, j int) bool {
	// 	return tracks[i].Popularity > artists[j].Popularity
	// })
	return tracks, nil
}

func (p *SpotifySongProvider) SearchAlbumsByName(accessToken, name string) ([]AlbumData, error) {
	limit := "50"
	albumQuery := "album:" + strings.ToLower(name)

	apiURL, err := url.Parse("https://api.spotify.com/v1/search")
	if err != nil {
		return nil, err
	}
	q := apiURL.Query()
	q.Set("type", "album")
	q.Set("q", albumQuery)
	q.Set("limit", limit)
	apiURL.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", apiURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	// var searchAlbumResponse struct {
	// 	Albums struct {
	// 		Items []struct {
	// 			ID         string `json:"id"`
	// 			Name       string `json:"name"`
	// 			Popularity int    `json:"popularity"`
	// 			Images     []struct {
	// 				URL    string `json:"url"`
	// 				Height int    `json:"height"`
	// 				Width  int    `json:"width"`
	// 			} `json:"images"`
	// 		} `json:"items"`
	// 	} `json:"artists"`
	// }
	//
	// if err := json.NewDecoder(resp.Body).Decode(&searchAlbumResponse); err != nil {
	// 	return nil, err
	// }

	return []AlbumData{}, nil
}

func (p *SpotifySongProvider) SearchArtistsByName(accessToken, name string) ([]ArtistData, error) {
	limit := "50"
	artistQuery := "artist:" + strings.ToLower(name)

	apiURL, err := url.Parse("https://api.spotify.com/v1/search")
	if err != nil {
		return nil, err
	}
	q := apiURL.Query()
	q.Set("type", "artist")
	q.Set("q", artistQuery)
	q.Set("limit", limit)
	apiURL.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", apiURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

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

	if err := json.NewDecoder(resp.Body).Decode(&searchArtistResponse); err != nil {
		return nil, err
	}

	// Convert to ArtistInfo
	artists := make([]ArtistData, 0, len(searchArtistResponse.Artists.Items))
	for _, a := range searchArtistResponse.Artists.Items {

		// TODO: add here and temp url???
		imageUrl := ""
		if len(a.Images) > 0 {
			imageUrl = a.Images[0].URL
		}
		artistInfo := ArtistData{
			Id:         a.ID,
			Name:       a.Name,
			ImageUrl:   imageUrl,
			Popularity: a.Popularity,
		}
		artists = append(artists, artistInfo)
	}
	sort.Slice(artists, func(i, j int) bool {
		return artists[i].Popularity > artists[j].Popularity
	})
	return artists, nil
}

func (p *SpotifySongProvider) SearchPlaylistsByName(accessToken, name string) ([]PlaylistData, error) {
	limit := "50"
	playlistQuery := "playlist:" + strings.ToLower(name)

	apiURL, err := url.Parse("https://api.spotify.com/v1/search")
	if err != nil {
		return nil, err
	}
	q := apiURL.Query()
	q.Set("type", "playlist")
	q.Set("q", playlistQuery)
	q.Set("limit", limit)
	apiURL.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", apiURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	// var searchPlaylistResponse struct {
	// 	Playlists struct {
	// 		Items []struct {
	// 			ID         string `json:"id"`
	// 			Name       string `json:"name"`
	// 			Popularity int    `json:"popularity"`
	// 			Images     []struct {
	// 				URL    string `json:"url"`
	// 				Height int    `json:"height"`
	// 				Width  int    `json:"width"`
	// 			} `json:"images"`
	// 		} `json:"items"`
	// 	} `json:"artists"`
	// }
	//
	// if err := json.NewDecoder(resp.Body).Decode(&searchPlaylistResponse); err != nil {
	// 	return nil, err
	// }

	return []PlaylistData{}, nil
}
