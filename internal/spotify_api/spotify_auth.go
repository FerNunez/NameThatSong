package spotify_api

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

func (p *SpotifySongProvider) AuthRequestURL() (string, error) {
	// Build the authorization URL
	authURL := "https://accounts.spotify.com/authorize"
	u, err := url.Parse(authURL)
	if err != nil {
		return "", err
	}

	// Add query parameters
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", p.ClientID)
	q.Set("scope", "user-read-private user-read-email streaming user-modify-playback-state user-read-playback-state")
	q.Set("redirect_uri", p.RedirectURI)
	q.Set("state", p.State)
	u.RawQuery = q.Encode()

	// Redirect to Spotify
	//http.Redirect(w, r, u.String(), http.StatusFound)
	return u.String(), nil
}

func (p *SpotifySongProvider) ValidateState(state string) error {
	if state != p.State {
		return errors.New("state is not validated")
	}
	return nil
}

// Exchange code for tokens
func (p *SpotifySongProvider) TokenExchange(code string) (TokenResponse, error) {
	tokenURL := "https://accounts.spotify.com/api/token"
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", p.RedirectURI)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Printf("Error creating request: %v", err)
		return TokenResponse{}, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	auth := base64.StdEncoding.EncodeToString([]byte(p.ClientID + ":" + p.ClientSecret))
	req.Header.Set("Authorization", "Basic "+auth)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error getting token: %v", err)
		return TokenResponse{}, err
	}
	defer resp.Body.Close()

	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		fmt.Printf("Error parsing token response: %v", err)
		return TokenResponse{}, err
	}

	// Store tokens
	// TODO: Set AccessToken to spotify API
	// TODO: Set RefreshToken to spotify API
	// p.AccessToken = tokenResponse.AccessToken
	// p.RefreshToken = tokenResponse.RefreshToken
	// fmt.Printf("Got the tokens %v, %v", p.AccessToken, p.RefreshToken)

	return tokenResponse, nil
}

func (p *SpotifySongProvider) RegenerateToken() (TokenResponse, error) {
	tokenURL := "https://accounts.spotify.com/api/token"
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", p.RefreshToken)
	data.Set("client_id", p.ClientID)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Printf("Error creating request: %v", err)
		return TokenResponse{}, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	auth := base64.StdEncoding.EncodeToString([]byte(p.ClientID + ":" + p.ClientSecret))
	req.Header.Set("Authorization", "Basic "+auth)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error getting token: %v", err)
		return TokenResponse{}, err
	}
	defer resp.Body.Close()

	var tokenResponse TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		fmt.Printf("Error parsing token response: %v", err)
		return TokenResponse{}, err
	}

	return tokenResponse, nil
}
