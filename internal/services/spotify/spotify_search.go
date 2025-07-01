package spotify

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	m "github.com/FerNunez/NameThatSong/internal/models"
)

func search(accessToken, limit, atype, query string) ([]byte, error) {
	trackQuery := atype + ":" + strings.ToLower(query)

	apiURL, err := url.Parse("https://api.spotify.com/v1/search")
	if err != nil {
		return nil, err
	}
	q := apiURL.Query()
	q.Set("type", atype)
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

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func SearchTracksByName(accessToken, name string) ([]m.TrackSearch, error) {
	limit := "50"

	data, err := search(accessToken, limit, "track", name)
	if err != nil {
		return nil, err
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

		trackInfo := m.TrackSearch{
			Id:         t.ID,
			Name:       t.Name,
			Popularity: t.Popularity,
			ArtistList: artists,
		}
		tracks = append(tracks, trackInfo)
	}
	return tracks, nil
}

func SearchAlbumsByName(accessToken, name string) ([]m.AlbumSearch, error) {
	limit := "50"

	data, err := search(accessToken, limit, "album", name)
	if err != nil {
		return nil, err
	}

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
			Id:         alb.ID,
			ImageUrl:   imageUrl,
			Name:       alb.Name,
			ArtistList: artists,
		}
	}
	return albums, nil
}

func SearchArtistsByName(accessToken, name string) ([]m.ArtistSearch, error) {
	limit := "50"

	data, err := search(accessToken, limit, "artist", name)
	if err != nil {
		return nil, err
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
			Id:         a.ID,
			Name:       a.Name,
			ImageUrl:   imageUrl,
			Popularity: a.Popularity,
		}
		artists = append(artists, artistInfo)
	}
	return artists, nil
}

func SearchPlaylistsByName(accessToken, name string) ([]m.PlaylistSearch, error) {
	limit := "50"
	data, err := search(accessToken, limit, "playlist", name)
	if err != nil {
		return nil, err
	}

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
			Id:       p.ID,
			ImageUrl: imageUrl,
			Name:     p.Name,
		})

	}
	return playlists, nil
}
