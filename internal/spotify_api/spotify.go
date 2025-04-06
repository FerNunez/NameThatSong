package spotify_api

type ArtistData struct {
	Id         string
	Name       string
	ImageUrl   string
	Popularity int
}

// SpotifySongProvider implements the SongProvider interface for Spotify
type SpotifySongProvider struct {
	AccessToken  string
	RefreshToken string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	State        string
}

// NewSpotifySongProvider creates a new SpotifySongProvider
func NewSpotifySongProvider(clientID, clientSecret string, redirectURI string, state string) *SpotifySongProvider {
	return &SpotifySongProvider{
		AccessToken:  "",
		RefreshToken: "",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  redirectURI,
		State:        state,
	}
}

func (p *SpotifySongProvider) SetToken(accessToken string, refreshToken string) {
	p.AccessToken = accessToken
	p.RefreshToken = refreshToken
}
