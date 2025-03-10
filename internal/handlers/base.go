package handlers

import (
	"fmt"
	"goth/internal/base"
	"goth/internal/guesser"
	"goth/internal/utils"
	"os"

	"github.com/joho/godotenv"
)

type ApiConfig struct {
	ClientID          string
	ClientSecret      string
	State             string
	AuthorizationCode string
	AccessToken       string
	RefreshToken      string
	DeviceId          string
}

type SpotifyApi struct {
	Config      ApiConfig
	Cache       base.ApiCache
	MusicPlayer guesser.MusicPlayer
}

func NewSpotifyApi() SpotifyApi {
	godotenv.Load()
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	apiConfig := ApiConfig{
		ClientID:          clientID,
		ClientSecret:      clientSecret,
		State:             "",
		AuthorizationCode: "",
		AccessToken:       "",
		RefreshToken:      "",
		DeviceId:          "",
	}

	state, err := utils.GenerateState(16)
	if err != nil {
		errmsg := fmt.Sprintf("could not generate random state: %v", state)
		fmt.Println(errmsg)
	}

	return SpotifyApi{
		Config:      apiConfig,
		Cache:       base.NewApiCache(),
		MusicPlayer: guesser.MusicPlayer{},
	}
}
