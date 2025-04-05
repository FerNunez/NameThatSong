package spotify_api

type ArtistInfo struct {
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
}

// NewSpotifySongProvider creates a new SpotifySongProvider
func NewSpotifySongProvider(accessToken, refreshToken, clientID, clientSecret string) *SpotifySongProvider {
	return &SpotifySongProvider{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

func (p *SpotifySongProvider) SetToken(accessToken string, refreshToken string) {
	p.AccessToken = accessToken
	p.RefreshToken = refreshToken
}
