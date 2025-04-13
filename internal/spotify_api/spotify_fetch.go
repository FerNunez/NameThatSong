package spotify_api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type AlbumData struct {
	AlbumType   string
	TotalTracks int
	ID          string
	ImagesURL   string
	Name        string
	ReleaseDate string
}

// fetch album by ID: retireves all songs
// https://api.spotify.com/v1/artists/{id}/albums&
// id=album&
// include_groups= album
// limit=50
func (p *SpotifySongProvider) FetchAlbumByArtistID(artistId string) ([]AlbumData, error) {

	limit := 50
	include_groups := "album"
	requestURL := fmt.Sprintf("https://api.spotify.com/v1/artists/%s/albums?limit=%d&include_groups=%v", artistId, limit, include_groups)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", p.AccessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	type FetchAlbumByArtistIDResponse struct {
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

	var albumsResponse FetchAlbumByArtistIDResponse
	if err := json.NewDecoder(resp.Body).Decode(&albumsResponse); err != nil {
		return nil, err
	}

	AlbumList := make([]AlbumData, 0, len(albumsResponse.Items))
	for _, item := range albumsResponse.Items {
		if item.AlbumType != "album" {
			continue
		}
		if len(item.Images) <= 0 {
			panic("spotify response of album with no images")
		}

		album := AlbumData{
			AlbumType:   item.AlbumType,
			TotalTracks: item.TotalTracks,
			ID:          item.ID,
			ImagesURL:   item.Images[0].URL,
			Name:        item.Name,
			ReleaseDate: item.ReleaseDate,
		}
		AlbumList = append(AlbumList, album)
	}
	return AlbumList, nil
}

type TrackData struct {
	DiscNumber  int
	DurationMs  int
	ID          string
	Name        string
	TrackNumber int
}

// https://api.spotify.com/v1/albums/{id}/tracks
// id= 4aawyAB9vmqN3uQ7FjRGTy
// limit =50
// fetch artists by ID:
func (p *SpotifySongProvider) FetchTracksByAlbumID(albumId string) ([]TrackData, error) {
	limit := 50
	requestURL := fmt.Sprintf("https://api.spotify.com/v1/albums/%s/tracks?limit=%d", albumId, limit)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", p.AccessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	type FetchTracksByAlbumIDResponse struct {
		Href  string `json:"href"`
		Items []struct {
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
		Limit    int `json:"limit"`
		Next     any `json:"next"`
		Offset   int `json:"offset"`
		Previous any `json:"previous"`
		Total    int `json:"total"`
	}

	var tracksResponse FetchTracksByAlbumIDResponse
	if err := json.NewDecoder(resp.Body).Decode(&tracksResponse); err != nil {
		return nil, err
	}

	trackList := make([]TrackData, 0, len(tracksResponse.Items))
	for _, item := range tracksResponse.Items {

		track := TrackData{
			DiscNumber:  item.DiscNumber,
			DurationMs:  item.DurationMs,
			ID:          item.ID,
			Name:        item.Name,
			TrackNumber: item.TrackNumber,
		}
		trackList = append(trackList, track)
	}
	return trackList, nil
}

func (p *SpotifySongProvider) CreateAlbumFromTopTracks(artistId string) (AlbumData, []TrackData, error) {

	requestURL := fmt.Sprintf("https://api.spotify.com/v1/artists/%v/top-tracks", artistId)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return AlbumData{}, nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", p.AccessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return AlbumData{}, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return AlbumData{}, nil, err
	}

	type TopTracks struct {
		Tracks []struct {
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
					URL    string `json:"url"`
					Height int    `json:"height"`
					Width  int    `json:"width"`
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
		} `json:"tracks"`
	}

	var tt TopTracks
	if err := json.NewDecoder(resp.Body).Decode(&tt); err != nil {
		return AlbumData{}, nil, err
	}

	trackList := make([]TrackData, 0, len(tt.Tracks))
	for _, item := range tt.Tracks {

		track := TrackData{
			DiscNumber:  item.DiscNumber,
			DurationMs:  item.DurationMs,
			ID:          item.ID,
			Name:        item.Name,
			TrackNumber: item.TrackNumber,
		}
		trackList = append(trackList, track)
	}

	fakeID := "topTracks" + artistId
	album := AlbumData{
		AlbumType:   "TopTracks",
		TotalTracks: len(trackList),
		ID:          fakeID,
		ImagesURL:   "",
		Name:        fmt.Sprintf("Top%v", len(trackList)),
		ReleaseDate: "NEW",
	}
	fmt.Println("Fake album: ", album.Name, "trakcs: ", album.TotalTracks, "ID:", album.ID)
	return album, trackList, nil
}
