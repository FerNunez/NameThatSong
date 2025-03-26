package handlers

import (
	"encoding/json"
	"fmt"
	"goth/internal/base"
	"goth/internal/templates"
	"goth/internal/utils"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

func (api *SpotifyApi) StartProcess(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Processing albums: %v\n", api.SelectedAlbumsIdsToTracksId)
	// Set Music Player songs
	listSongsId := []string{}
	for _, listSongsIdByAlbum := range api.SelectedAlbumsIdsToTracksId {
		listSongsId = append(listSongsId, listSongsIdByAlbum...)
	}
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(listSongsId), func(i, j int) {
		listSongsId[i], listSongsId[j] = listSongsId[j], listSongsId[i]
	})
	api.MusicPlayer.Queue = listSongsId[:10]

	for _, id := range api.MusicPlayer.Queue {
		fmt.Printf("%v ", api.Cache.TracksCache[id].Track.Name)
	}

	w.Write([]byte("Added songs :)"))

}

func (cfg *SpotifyApi) SearchHelper(w http.ResponseWriter, r *http.Request) {

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
		if _, exists := cfg.Cache.ArtistsCache[artist.ID]; !exists {
			cfg.Cache.ArtistsCache[artist.ID] = base.ArtistCache{
				Artist:     artist,
				LastUpdate: time.Now(),
			}
		}
	}

	component := templates.SearchResults(artistList)
	component.Render(r.Context(), w)

}

func (cfg *SpotifyApi) AlbumSelection(w http.ResponseWriter, r *http.Request) {

	// Get Album IDs
	err := r.ParseForm()
	if err != nil {
		return
	}
	albumID := r.Form.Get("albumID")

	// get tracks of album
	tracks, err := getAlbumsTracks(cfg.Config.AccessToken, albumID)
	if err != nil {
		errmsg := fmt.Sprintf("could not get tracks for album: %v", err)
		fmt.Println(errmsg)
	}

	// Update tracks cache
	tracksIds := make([]string, 0, len(tracks))
	for _, track := range tracks {
		cfg.Cache.TracksCache[track.ID] = base.TrackCache{
			Track:      track,
			LastUpdate: time.Now(),
		}
		tracksIds = append(tracksIds, track.ID)
	}

	// Add or remove from list
	_, exists := cfg.SelectedAlbumsIdsToTracksId[albumID]
	if !exists {
		cfg.SelectedAlbumsIdsToTracksId[albumID] = tracksIds
	} else {
		delete(cfg.SelectedAlbumsIdsToTracksId, albumID)
	}

	// Update cache
	albumCache := cfg.Cache.AlbumsCache[albumID]
	albumCache.ToggleSelect()
	cfg.Cache.AlbumsCache[albumID] = albumCache

	for albumID, tracksIds := range cfg.SelectedAlbumsIdsToTracksId {
		fmt.Printf("albumId: %v, tracksIds: %v\n", albumID, tracksIds)
	}

	component := templates.AlbumCard(albumCache.Album)
	component.Render(r.Context(), w)

}

func getAlbumsTracks(accessToken, albumId string) ([]base.Track, error) {

	requestUrl := fmt.Sprintf("https://api.spotify.com/v1/albums/%v/tracks", albumId)
	fmt.Println(requestUrl)
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		errmsg := fmt.Sprintf("could not create request: %v", err)
		fmt.Println(errmsg)
		return []base.Track{}, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errmsg := fmt.Sprintf("could not do request: %v", err)
		fmt.Println(errmsg)
		return []base.Track{}, err
	}
	if resp.StatusCode != http.StatusOK {
		errmsg := fmt.Sprintf("could not get good status code: %v", resp.StatusCode)
		fmt.Println(errmsg)
		return []base.Track{}, err
	}
	defer resp.Body.Close()

	var albumTracksResponse base.GetAlbumTracksResponse
	err = json.NewDecoder(resp.Body).Decode(&albumTracksResponse)
	if err != nil {
		errmsg := fmt.Sprintf("could not decode response: %v", err)
		fmt.Println(errmsg)
		return []base.Track{}, err
	}

	tracks := utils.ParseAlbumResponseToTracks(albumTracksResponse, albumId)
	return tracks, nil
}

func (cfg *SpotifyApi) GuessHelper(w http.ResponseWriter, r *http.Request) {

	// get query
	requestQuery := r.URL.Query().Get("guess")
	if requestQuery == "" {
		component := templates.SearchResults([]base.Artist{})
		component.Render(r.Context(), w)

	}

	// Parse track /a artist /b album
	cleanedQuery := utils.ParseGuessText(requestQuery)

	// hit search endpoint spoti
	newUrl, err := url.Parse("https://api.spotify.com/v1/search")
	if err != nil {
		errmsg := fmt.Sprintf("Error parsing URL: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}
	urlQuery := newUrl.Query()
	urlQuery.Set("type", "track")
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

	var searchTrackResponse base.SearchTrackResponse
	err = json.NewDecoder(resp.Body).Decode(&searchTrackResponse)
	if err != nil {
		errmsg := fmt.Sprintf("could not decode response: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}

	tracks, artistNames, albumUris := cfg.ParserTracksResponse(searchTrackResponse)

	// Note: Tracks are already sorted by popularity in descending order in ParserTracksResponse

	guesserHelperList := templates.GuessHelperList(tracks, artistNames, albumUris)
	guesserHelperList.Render(r.Context(), w)
}

func (cfg *SpotifyApi) GuessTrack(w http.ResponseWriter, r *http.Request) {

	// get query user written
	requestQuery := r.URL.Query().Get("search")
	if requestQuery == "" {
		component := templates.SearchResults([]base.Artist{})
		component.Render(r.Context(), w)

	}
	lowerQuery := strings.ToLower(requestQuery)
	cleanedQuery := "artist:" + lowerQuery

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
		if _, exists := cfg.Cache.ArtistsCache[artist.ID]; !exists {
			cfg.Cache.ArtistsCache[artist.ID] = base.ArtistCache{
				Artist:     artist,
				LastUpdate: time.Now(),
			}
		}
	}

	component := templates.SearchResults(artistList)
	component.Render(r.Context(), w)

}

func (cfg *SpotifyApi) SelectTrack(w http.ResponseWriter, r *http.Request) {

	// Get Album IDs
	err := r.ParseForm()
	if err != nil {
		return
	}
	trackID := r.Form.Get("trackID")

	track, err := cfg.Cache.RetrieveTrack(trackID)
	if err != nil {
		errmsg := fmt.Errorf("cant find track %v", err)
		println(errmsg)
	}
	track.Selected = !track.Selected
	cfg.Cache.InsertTrack(track.ID, track)

	album, err := cfg.Cache.RetrieveAlbum(track.AlbumID)
	if err != nil {
		errmsg := fmt.Errorf("cant find album %v", err)
		println(errmsg)
	}

	if len(track.ArtistsID) <= 0 {
		panic("no artists available")
	}
	artist, err := cfg.Cache.RetrieveArtist(track.ArtistsID[0])
	if err != nil {
		errmsg := fmt.Errorf("cant find artist %v", err)
		println(errmsg)
	}

	fmt.Printf("track name: %v\n", track.Name)
	fmt.Printf("album name: %v\n", album.Name)
	fmt.Printf("artist name: %v\n", artist.Name)
	if len(album.Images) <= 0 {
		panic("album has no image in select track")
	}
	fmt.Println("url: ", album.Images[0].URL)
	gocComponent := templates.GuessOptionCard(track.ID, track.Name, album.Images[0].URL, artist.Name, track.Selected)
	gocComponent.Render(r.Context(), w)
	//templates.GuessOptionCard()
	////component := templates.GuessOptionCard(albumCache.Album)
	//component.Render(r.Context(), w)

}
