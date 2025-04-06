package spotify_api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

func (p *SpotifySongProvider) SearchArtistsByName(name string) ([]ArtistData, error) {

	fmt.Println("Got query: name, ", name)
	artistQuery := "artist:" + strings.ToLower(name)

	// Call Spotify API
	apiURL, err := url.Parse("https://api.spotify.com/v1/search")
	if err != nil {
		return nil, err
	}
	q := apiURL.Query()
	q.Set("type", "artist")
	q.Set("q", artistQuery)
	apiURL.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", apiURL.String(), nil)
	if err != nil {
		return nil, err
	}

	// Set Authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.AccessToken))

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err

	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	// Parse response
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

		artistInfo := ArtistData{
			Id:         a.ID,
			Name:       a.Name,
			ImageUrl:   a.Images[0].URL,
			Popularity: a.Popularity,
		}
		artists = append(artists, artistInfo)
	}

	sort.Slice(artists, func(i, j int) bool {
		return artists[i].Popularity > artists[j].Popularity
	})

	fmt.Println(artists)

	return artists, nil
}
