package spotify_api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

type TrackData struct {
	DiscNumber  int
	DurationMs  int
	ID          string
	Name        string
	TrackNumber int
	Popularity  int
	Explicit    bool
}

type AlbumData struct {
	AlbumType   string
	TotalTracks int
	ID          string
	ImagesURL   string
	Name        string
	ReleaseDate string
}

type ArtistData struct {
	Id         string
	Name       string
	ImageUrl   string
	Popularity int
}

type PlaylistData struct {
	Description    string
	FollowersTotal int
	ID             string
	ImageUrl       string
	Name           string
	Public         bool
	TotalTracks    int
}

func (p *SpotifySongProvider) FetchTrack(accessToken, trackId string) (TrackData, error) {
	fmt.Println("[SpotifySongProvider] FetchTrack: trackId:", trackId)
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
		return TrackData{}, errors.New(fmt.Sprintf("bad status code:%v", resp.StatusCode))
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
		Popularity:  fetchTrackResponse.Popularity,
		Explicit:    fetchTrackResponse.Explicit,
	}, nil
}

// Fetch album by ID: retireves all tracks of the album too
func (p *SpotifySongProvider) FetchAlbum(accessToken, albumId string) (AlbumData, []TrackData, error) {
	requestURL := fmt.Sprintf("https://api.spotify.com/v1/albums/%s", albumId)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return AlbumData{}, []TrackData{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return AlbumData{}, []TrackData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return AlbumData{}, []TrackData{}, errors.New(fmt.Sprintf("bad status code:%v", resp.StatusCode))

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
		return AlbumData{}, []TrackData{}, err
	}

	imageUrl := ""
	if len(fetchAlbumResponse.Images) > 0 {
		imageUrl = fetchAlbumResponse.Images[len(fetchAlbumResponse.Images)-1].URL
	}

	tracks := make([]TrackData, len(fetchAlbumResponse.Tracks.Items))
	for idx, track := range fetchAlbumResponse.Tracks.Items {
		tracks[idx] = TrackData{
			DiscNumber:  track.DiscNumber,
			DurationMs:  track.DurationMs,
			ID:          track.ID,
			Name:        track.Name,
			TrackNumber: track.TrackNumber,
		}
	}

	return AlbumData{
		AlbumType:   fetchAlbumResponse.AlbumType,
		TotalTracks: fetchAlbumResponse.TotalTracks,
		ID:          albumId,
		ImagesURL:   imageUrl,
		Name:        fetchAlbumResponse.Name,
		ReleaseDate: fetchAlbumResponse.ReleaseDate,
	}, tracks, nil
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
		return ArtistData{}, errors.New(fmt.Sprintf("bad status code:%v", resp.StatusCode))

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

func (p *SpotifySongProvider) FetchPlaylist(accessToken, playlistId string) (PlaylistData, []TrackData, error) {

	baseURL := "https://api.spotify.com/v1/playlists/" + playlistId
	u, err := url.Parse(baseURL)
	if err != nil {
		return PlaylistData{}, []TrackData{}, err
	}
	q := u.Query()
	q.Set("fields", "id,name,description,public,followers(total),images(url),tracks(items(track(id,name,preview_url,external_urls(spotify),album(name,images(url)))))")
	// q.Set("fields", "id,name,description,images(url),external_urls(spotify),followers(total),public,owner(display_name),tracks(items(track(id)))")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return PlaylistData{}, []TrackData{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return PlaylistData{}, []TrackData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PlaylistData{}, []TrackData{}, errors.New(fmt.Sprintf("bad status code:%v", resp.StatusCode))
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
		return PlaylistData{}, []TrackData{}, err
	}

	imageUrl := ""
	if len(fetchPlaylistResponse.Images) > 0 {
		imageUrl = fetchPlaylistResponse.Images[len(fetchPlaylistResponse.Images)-1].URL
	}

	tracks := make([]TrackData, len(fetchPlaylistResponse.Tracks.Items))
	tracksIds := make([]string, len(fetchPlaylistResponse.Tracks.Items))
	for idx, track := range fetchPlaylistResponse.Tracks.Items {
		tracksIds[idx] = track.Track.ID
		tracks[idx] = TrackData{
			DiscNumber:  track.Track.DiscNumber,
			DurationMs:  track.Track.DurationMs,
			ID:          track.Track.ID,
			Name:        track.Track.Name,
			TrackNumber: track.Track.TrackNumber,
		}
	}

	return PlaylistData{
		Description:    fetchPlaylistResponse.Description,
		FollowersTotal: fetchPlaylistResponse.Followers.Total,
		ID:             playlistId,
		ImageUrl:       imageUrl,
		Name:           fetchPlaylistResponse.Name,
		Public:         fetchPlaylistResponse.Public,
		TotalTracks:    len(fetchPlaylistResponse.Tracks.Items),
	}, tracks, nil
}

// ////////////////////
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
		return nil, errors.New(fmt.Sprintf("bad status code:%v", resp.StatusCode))

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
		return nil, errors.New(fmt.Sprintf("bad status code:%v", resp.StatusCode))

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
		return AlbumData{}, nil, errors.New(fmt.Sprintf("bad status code:%v", resp.StatusCode))

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

func (p *SpotifySongProvider) FetchAlbumsFromArtist(accessToken, artistId string) ([]string, error) {
	limit := 50
	include_groups := "album"

	requestURL := fmt.Sprintf("https://api.spotify.com/v1/artists/%v/albums?include_groups=%v&limit=%v", artistId, include_groups, limit)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return []string{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []string{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []string{}, errors.New(fmt.Sprintf("bad status code: %v", resp.StatusCode))
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

	trackIds := make([]string, len(albumsFromArtistResponse.Items))
	for idx, track := range albumsFromArtistResponse.Items {
		trackIds[idx] = track.ID
	}

	return trackIds, nil
}

func (p *SpotifySongProvider) FetchTracksFromAlbum(accessToken, albumId string) ([]string, error) {
	limit := 50

	requestURL := fmt.Sprintf("https://api.spotify.com/v1/albums/%v/tracks?&limit=%v", albumId, limit)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return []string{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []string{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []string{}, errors.New(fmt.Sprintf("bad status code: %v", resp.StatusCode))
	}

	// TODO: Finish
	return []string{}, nil
}
func (p *SpotifySongProvider) FetchTracksFromPlaylist(accessToken, playlistId string) ([]string, error) {
	limit := 50

	requestURL := fmt.Sprintf("https://api.spotify.com/v1/playlists/%v/tracks?&limit=%v", playlistId, limit)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return []string{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []string{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []string{}, errors.New(fmt.Sprintf("bad status code: %v", resp.StatusCode))
	}

	// TODO: Finish
	return []string{}, nil
}
