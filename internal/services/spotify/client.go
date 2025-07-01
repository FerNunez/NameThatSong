package spotify

import (
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/config"
	"github.com/FerNunez/NameThatSong/internal/repository"
	"github.com/FerNunez/NameThatSong/internal/services/cache"
)

type Service struct {
	client        *http.Client
	spotifyConfig *config.SpotifyConfig
	tokenManager  *TokenManager
	cache         cache.SpotifyCache
	repo          repository.SpotifyRepository
}

func New(c *http.Client, Config *config.SpotifyConfig, sCache cache.SpotifyCache) (*Service, error) {

	return &Service{
		client:        c,
		spotifyConfig: sConfig,
		State:         state,
		AccessToken:   "",
		RefreshToken:  "",
		cache:         sCache,
	}, nil
}
