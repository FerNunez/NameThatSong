package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"goth/internal/base"
	"goth/internal/templates"
	"goth/internal/utils"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// create a struct ApiHandler
// inside ApiHandler add a caché of last responses?
// go from string [query] to object response

func (cfg *SpotifyApi) RequestUserAuthorizationHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("login")

	baseUrl := "https://accounts.spotify.com/authorize"

	url, err := url.Parse(baseUrl)
	if err != nil {
		errmsg := fmt.Sprintf("Error parsing URL: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}

	query := url.Query()
	query.Set("client_id", cfg.Config.ClientID)
	query.Set("response_type", "code")
	query.Set("redirect_uri", "http://127.0.0.1:8080/auth/callback")
	query.Set("state", cfg.Config.State)
	query.Set("scope", "user-read-playback-state user-read-private user-read-email user-modify-playback-state")

	url.RawQuery = query.Encode()
	http.Redirect(w, r, url.String(), http.StatusFound)

}

func (cfg *SpotifyApi) RequestUserAuthorizationCallbackHandler(w http.ResponseWriter, r *http.Request) {

	// Check response
	respError := r.URL.Query().Get("error")
	respCode := r.URL.Query().Get("code")
	respState := r.URL.Query().Get("state")
	if respState != cfg.Config.State {
		errmsg := fmt.Sprintln("could not validate state")
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}
	if respError != "" {
		errmsg := fmt.Sprintf("could not get authorization user: %v", respError)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}
	if respCode == "" {
		errmsg := fmt.Sprintf("could not retrieve code: %v", respError)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}
	cfg.Config.AuthorizationCode = respCode

	// Request AccessToken
	baseUrl := "https://accounts.spotify.com/api/token"
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", cfg.Config.AuthorizationCode)
	data.Set("redirect_uri", "http://127.0.0.1:8080/auth/callback")

	request, err := http.NewRequest("POST", baseUrl, bytes.NewBufferString(data.Encode()))
	if err != nil {
		errmsg := fmt.Sprintf("could not create request: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	encoded := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v:%v", cfg.Config.ClientID, cfg.Config.ClientSecret)))
	request.Header.Set("Authorization", fmt.Sprintf("Basic %v", encoded))
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		errmsg := fmt.Sprintf("could not do request: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}
	defer resp.Body.Close()

	// Check response
	type AccessTokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
	}

	var accessTokenResp AccessTokenResp
	err = json.NewDecoder(resp.Body).Decode(&accessTokenResp)
	if err != nil {
		errmsg := fmt.Sprintf("could not decode response: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}

	cfg.Config.AccessToken = accessTokenResp.AccessToken
	cfg.Config.RefreshToken = accessTokenResp.RefreshToken

	fmt.Printf("acess token: %v\n", cfg.Config.AccessToken)
	fmt.Printf("RefreshToken %v\n", cfg.Config.RefreshToken)

	http.Redirect(w, r, "http://127.0.0.1:8080", http.StatusFound)
}

func (cfg *SpotifyApi) RequestStartHandler(w http.ResponseWriter, r *http.Request) {

	type PlaySongRequest struct {
		Uris       []string `json:"uris"`
		PositionMs int      `json:"position_ms"`
	}

	cfg.MusicPlayer.Next()

	psr := PlaySongRequest{
		Uris:       []string{fmt.Sprintf("spotify:track:%v", cfg.MusicPlayer.Current)},
		PositionMs: 0,
	}

	dat, err := json.Marshal(psr)
	if err != nil {
		fmt.Println("HELP")
	}

	// Set request
	req, err := http.NewRequest("PUT", "https://api.spotify.com/v1/me/player/play", bytes.NewBuffer(dat))
	if err != nil {
		errmsg := fmt.Sprintf("could not create request for me: %v", err)
		fmt.Println(errmsg)
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", cfg.Config.AccessToken))
	req.Header.Set("Content-Type", "application/json")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errmsg := fmt.Sprintf("could not do request for me: %v", err)
		fmt.Println(errmsg)
		return
	}
	if resp.StatusCode != http.StatusNoContent {
		errmsg := fmt.Sprintf("could not get good status code: %v", resp.StatusCode)
		fmt.Println(errmsg)

		var responserError base.ErrorResponse

		err := json.NewDecoder(resp.Body).Decode(&responserError)
		if err == nil {
			fmt.Printf("Status: %v\n Message: %v\n", responserError.Error.Status, responserError.Error.Message)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
}

func (cfg *SpotifyApi) RequestStartHandler2(w http.ResponseWriter, r *http.Request) {

	for _, listTracks := range cfg.SelectedAlbumsIdsToTracksId {

		for _, trackId := range listTracks {

			// add to quey
			newUrl, err := url.Parse("https://api.spotify.com/v1/me/player/queue")
			if err != nil {
				errmsg := fmt.Sprintf("Error parsing URL: %v", err)
				fmt.Println(errmsg)
				return
			}
			urlQuery := newUrl.Query()
			urlQuery.Set("uri", fmt.Sprintf("spotify:track:%v", trackId))
			// Set request
			newUrl.RawQuery = urlQuery.Encode()
			req, err := http.NewRequest("POST", newUrl.String(), nil)
			fmt.Println("newUrl.String(): ", req.URL)
			if err != nil {
				errmsg := fmt.Sprintf("Error parsing URL: %v", err)
				fmt.Println(errmsg)
				return
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", cfg.Config.AccessToken))
			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				errmsg := fmt.Sprintf("could not do request: %v", err)
				fmt.Println(errmsg)
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				errmsg := fmt.Sprintf("could not retrieve status code : %v", err)
				fmt.Println(errmsg)
				return
			}
		}

	}

	type AutoGenerated struct {
		ContextURI string `json:"context_uri"`
		Offset     struct {
			Position int `json:"position"`
		} `json:"offset"`
		PositionMs int `json:"position_ms"`
	}

	// Set request
	req, err := http.NewRequest("PUT", "https://api.spotify.com/v1/me/player/play", nil)
	if err != nil {
		errmsg := fmt.Sprintf("could not create request for me: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", cfg.Config.AccessToken))
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errmsg := fmt.Sprintf("could not do request for me: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}
	if resp.StatusCode != http.StatusNoContent {
		errmsg := fmt.Sprintf("could not get good status code: %v", resp.StatusCode)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Playing"))
}

func (cfg *SpotifyApi) AlbumGridHttp(w http.ResponseWriter, r *http.Request) {

	artistName := strings.ToLower(r.URL.Query().Get("search"))
	fmt.Println("queyr:", artistName)

	artistId := r.URL.Query().Get("artistId")
	fmt.Println("artistId:", artistId)
	if artistId == "" {
		errmsg := fmt.Sprintln("no artist provided")
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errmsg))
		return
	}
	// Set request
	requestUrl := fmt.Sprintf("https://api.spotify.com/v1/artists/%v/albums", artistId)
	fmt.Println(requestUrl)
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		errmsg := fmt.Sprintf("could not create request: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", cfg.Config.AccessToken))

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

	var albumJson base.AlbumsJson
	err = json.NewDecoder(resp.Body).Decode(&albumJson)
	if err != nil {
		errmsg := fmt.Sprintf("could not decode response: %v", err)
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errmsg))
		return
	}

	albums, _ := utils.ParseAlbumsJson(albumJson)
	// TODO: fix err

	// Adding to Cache
	for _, album := range albums {
		cfg.Cache.AlbumsCache[album.ID] = base.AlbumCache{Album: album, LastUpdate: time.Now()}
	}

	// Return the search results component
	component := templates.AlbumGrid(albums)
	component.Render(r.Context(), w)
}

func RequestTracksFromAlbumID(accessToken string, albumID string) ([]base.Track, error) {

	// request
	requestUrl := fmt.Sprintf("https://api.spotify.com/v1/albums/%v", albumID)
	fmt.Println(requestUrl)
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		errmsg := fmt.Sprintf("could not create request: %v", err)
		fmt.Println(errmsg)
		return []base.Track{}, errors.New(errmsg)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errmsg := fmt.Sprintf("could not do request: %v", err)
		fmt.Println(errmsg)
		return []base.Track{}, errors.New(errmsg)
	}
	if resp.StatusCode != http.StatusOK {
		errmsg := fmt.Sprintf("could not get good status code: %v", resp.StatusCode)
		return []base.Track{}, errors.New(errmsg)
	}
	defer resp.Body.Close()

	// decode album
	var getAlbumResponse base.GetAlbumTracksResponse
	err = json.NewDecoder(resp.Body).Decode(&getAlbumResponse)
	if err != nil {
		errmsg := fmt.Sprintf("could not decode response: %v", err)
		fmt.Println(errmsg)
		return []base.Track{}, err
	}
	// parse json tracks into Tracks
	tracks := utils.ParseAlbumResponseToTracks(getAlbumResponse, albumID)

	return tracks, nil
}
