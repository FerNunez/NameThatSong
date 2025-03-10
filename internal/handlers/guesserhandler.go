package handlers

import (
	"encoding/json"
	"fmt"
	"goth/internal/base"
	"goth/internal/templates"
	"goth/internal/utils"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (api SpotifyApi) StartProcess(w http.ResponseWriter, r *http.Request) {
	fmt.Println("playing")

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	// 2. Get selected album IDs
	selectedIDs := r.Form["selectedAlbums"]
	fmt.Println("q: ", selectedIDs)
	var selectedAlbums []templates.Album

	output := ""
	for _, id := range selectedIDs {
		// Find matching album in your data store
		idInt, _ := strconv.Atoi(id)
		for _, album := range albums {
			if album.ID == idInt {
				selectedAlbums = append(selectedAlbums, album)
				output += fmt.Sprintf("%v\n", album.Title)
				break
			}
		}
	}
	// TODO: Add songs to a QUEUE

	// 4. Process selected albums (your logic here)
	fmt.Printf("Processing albums: %+v\n", selectedAlbums)

	musicPlayer := templates.MusicPlayer()
	layout := templates.GuesserLayout(musicPlayer, "Player")

	//component := templates.Index(output)
	//layout := templates.Layout(component, "game")
	layout.Render(r.Context(), w)

}

func (cfg *SpotifyApi) RequestArtistListByNameHandler(w http.ResponseWriter, r *http.Request) {

	// get query user written
	requestQuery := r.URL.Query().Get("search")
	if requestQuery == "" {
		component := templates.SearchResults([]base.Artist{})
		component.Render(r.Context(), w)

	}
	lowerQuery := strings.ToLower(requestQuery)
	cleanedQuery := "artist:" + lowerQuery

	// hittinh spotify for least
	newUrl, err := url.Parse("https://api.spotify.com/v1/search")
	if err != nil {
		errmsg := fmt.Sprintf("Error parsing URL: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}
	urlQuery := newUrl.Query()
	urlQuery.Set("type", "artist")
	urlQuery.Set("q", cleanedQuery)

	// Set request
	newUrl.RawQuery = urlQuery.Encode()
	req, err := http.NewRequest("GET", newUrl.String(), nil)
	if err != nil {
		errmsg := fmt.Sprintf("could not create request: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", cfg.Config.AccessToken))

	// Set Response
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errmsg := fmt.Sprintf("could not do request: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}
	if resp.StatusCode != http.StatusOK {
		errmsg := fmt.Sprintf("could not get good status code: %v", resp.StatusCode)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}
	defer resp.Body.Close()

	// Decode & Parse
	var searchedArtistResp struct {
		Artists base.ArtistsJson `json:"artists"`
	}
	err = json.NewDecoder(resp.Body).Decode(&searchedArtistResp)
	if err != nil {
		errmsg := fmt.Sprintf("could not decode response: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}
	artistList, _ := utils.ParseArtistsJson(searchedArtistResp.Artists)
	sort.Slice(artistList, func(i, j int) bool {
		return artistList[i].Popularity > artistList[j].Popularity
	})

	// Update Cache
	for _, artist := range artistList {
		if _, exists := cfg.Cache.ArtistCache[artist.ID]; !exists {
			cfg.Cache.ArtistCache[artist.ID] = base.ArtistCache{
				Artist:     artist,
				LastUpdate: time.Now(),
			}
		}
	}

	component := templates.SearchResults(artistList)
	component.Render(r.Context(), w)

}
