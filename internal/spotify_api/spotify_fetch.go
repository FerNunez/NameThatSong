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

func (p *SpotifySongProvider) FetchTrack(accessToken, trackId string) (TrackData, error) {
	requestURL := fmt.Sprintf("https://api.spotify.com/v1/tracks/%s", trackId)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return TrackData{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return TrackData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return TrackData{}, err
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
		return TrackData{}, err
	}

	return TrackData{
		DiscNumber:  fetchTrackResponse.DiscNumber,
		DurationMs:  fetchTrackResponse.DurationMs,
		ID:          trackId,
		Name:        fetchTrackResponse.Name,
		TrackNumber: fetchTrackResponse.TrackNumber,
	}, nil
}
func (p *SpotifySongProvider) FetchAlbum(accessToken, albumId string) (AlbumData, error) {
	requestURL := fmt.Sprintf("https://api.spotify.com/v1/albums/%s", albumId)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return AlbumData{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return AlbumData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return AlbumData{}, err
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
		return AlbumData{}, err
	}

	imageUrl := ""
	if len(fetchAlbumResponse.Images) > 0 {
		imageUrl = fetchAlbumResponse.Images[len(fetchAlbumResponse.Images)-1].URL
	}

	return AlbumData{
		AlbumType:   fetchAlbumResponse.AlbumType,
		TotalTracks: fetchAlbumResponse.TotalTracks,
		ID:          albumId,
		ImagesURL:   imageUrl,
		Name:        fetchAlbumResponse.Name,
		ReleaseDate: fetchAlbumResponse.ReleaseDate,
	}, nil
}
func (p *SpotifySongProvider) FetchArtist(accessToken, artistId string) (ArtistData, error) {

	requestURL := fmt.Sprintf("https://api.spotify.com/v1/artists/%s", artistId)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return ArtistData{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ArtistData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ArtistData{}, err
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
		return ArtistData{}, err
	}

	imageUrl := ""
	if len(fetchArtistResponse.Images) > 0 {
		imageUrl = fetchArtistResponse.Images[len(fetchArtistResponse.Images)-1].URL
	}

	return ArtistData{
		Id:         artistId,
		Name:       fetchArtistResponse.Name,
		ImageUrl:   imageUrl,
		Popularity: fetchArtistResponse.Popularity,
	}, nil
}

// fetch album by ID: retireves all songs
// https://api.spotify.com/v1/artists/{id}/albums&
// id=album&
// include_groups= album
// limit=50
func (p *SpotifySongProvider) FetchAlbumByArtistID(accesToken, artistId string) ([]AlbumData, error) {

	limit := 50
	include_groups := "album"
	requestURL := fmt.Sprintf("https://api.spotify.com/v1/artists/%s/albums?limit=%d&include_groups=%v", artistId, limit, include_groups)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accesToken))

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
func (p *SpotifySongProvider) FetchTracksByAlbumID(accessToken, albumId string) ([]TrackData, error) {
	limit := 50
	requestURL := fmt.Sprintf("https://api.spotify.com/v1/albums/%s/tracks?limit=%d", albumId, limit)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

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

// Creates an album with the top tracks of an artist
func (p *SpotifySongProvider) CreateTopTracksAlbum(accessToken, artistId string) (AlbumData, []TrackData, error) {

	requestURL := fmt.Sprintf("https://api.spotify.com/v1/artists/%v/top-tracks", artistId)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return AlbumData{}, nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

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
	//fmt.Println("Fake album: ", album.Name, "trakcs: ", album.TotalTracks, "ID:", album.ID)
	return album, trackList, nil
}
