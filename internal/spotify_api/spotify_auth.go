package spotify_api

import "net/url"

func (p *SpotifySongProvider) AuthRequest(state string, redirectUrl string) error {
	// Build the authorization URL
	authURL := "https://accounts.spotify.com/authorize"
	u, err := url.Parse(authURL)
	if err != nil {
		return err
	}

	// Add query parameters
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", p.ClientID)
	q.Set("scope", "user-read-private user-read-email streaming user-modify-playback-state user-read-playback-state")
	q.Set("redirect_uri", redirectUrl)
	q.Set("state", state)
	u.RawQuery = q.Encode()

	// Redirect to Spotify
	//http.Redirect(w, r, u.String(), http.StatusFound)
	return nil
}
