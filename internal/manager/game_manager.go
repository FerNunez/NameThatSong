package manager

import (
	"context"
	"fmt"
	"os"

	"github.com/FerNunez/NameThatSong/internal/middleware"
	"github.com/FerNunez/NameThatSong/internal/service"
	"github.com/joho/godotenv"
)

type GameManager struct {
	Games map[string]*service.GameService
}

func NewGameManager() *GameManager {
	return &GameManager{
		Games: make(map[string]*service.GameService),
	}
}

func (gm *GameManager) CreateGame(userId string) error {

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %v", err)
	}

	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return fmt.Errorf("missing Spotify credentials in .env file")
	}

	//redirectURI := "http://127.0.0.1:8080/auth/callback"
	redirectURI := os.Getenv("SPOTIFY_REDIRECT_URI")
	if redirectURI == "" {
		redirectURI = "http://127.0.0.1:8080/auth/callback"
		//"https://namethatsong.onrender.com/auth/callback"
	}

	gameService, err := service.NewGameService(clientID, clientSecret, redirectURI)
	if err != nil {
		return err
	}

	gm.Games[userId] = gameService
	return nil
}

func (gm GameManager) GetGame(ctx context.Context) (*service.GameService, error) {
	user, ok := middleware.GetUser(ctx)
	if !ok {
		return nil, fmt.Errorf("There is no user in context")
	}

	game, ok := gm.Games[user.ID.String()]
	if !ok {
		return nil, fmt.Errorf("There is no game for user id")
	}

	return game, nil
}
