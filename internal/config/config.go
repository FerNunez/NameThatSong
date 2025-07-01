package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

//TODO: Here a config file for Post and DatabaseUrl

type Config struct {
	Address            string
	Port               string
	DatabaseURL        string
	RedisURL           string
	Spotify            SpotifyConfig
	SpotifyTokenSecret string
	JWTSecret          string
}

type SpotifyConfig struct {
	ClientID     string
	ClientSecret string
	redirectURL  string
}

func New() (*Config, error) {

	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")
	if clientID == "" || clientSecret == "" || redirectURI == "" {
		return nil, fmt.Errorf("missing Spotify vars in .env file")
	}
	//redirectURI = "http://127.0.0.1:8080/auth/callback"
	//"https://namethatsong.onrender.com/auth/callback"
	//redirectURI := "http://127.0.0.1:8080/auth/callback"

	address := os.Getenv("ADDRESS")
	port := os.Getenv("PORT")
	if address == "" || port == "" {
		return nil, fmt.Errorf("missing address & port in .env file")
	}

	databaseURL := os.Getenv("DB_URL")
	redisURL := os.Getenv("REDIS_URL")
	if databaseURL == "" || redisURL == "" {
		return nil, fmt.Errorf("missing Database vars in .env file")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	spotifyTokenSecret := os.Getenv("SPOTIFY_TOKEN_SECRET")

	return &Config{
		Address:     address,
		Port:        port,
		DatabaseURL: databaseURL,
		RedisURL:    redisURL,
		Spotify: SpotifyConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			redirectURL:  redirectURI,
		},
		SpotifyTokenSecret: spotifyTokenSecret,
		JWTSecret:          jwtSecret,
	}, nil

}
