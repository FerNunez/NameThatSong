package manager

import (
	"context"
	"fmt"
	"os"

	"github.com/FerNunez/NameThatSong/internal/middleware"
	"github.com/FerNunez/NameThatSong/internal/service"
	"github.com/FerNunez/NameThatSong/internal/store"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type GameManager struct {
	Games             map[string]*service.GameService
	SpotifyTokenStore store.SpotifyTokenStore
}

func NewGameManager() *GameManager {
	return &GameManager{
		Games: make(map[string]*service.GameService),
	}
}

func (gm *GameManager) CreateGame(userId uuid.UUID, spotifyTokenStore store.SpotifyTokenStore) error {

	// TODO: Pass all this into a internal config struct?
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

	gameService, err := service.NewGameService(clientID, clientSecret, redirectURI, userId, spotifyTokenStore)
	if err != nil {
		return err
	}

	gm.Games[userId.String()] = gameService
	fmt.Println("[GameManager] CreateGame: New game created for user", userId.String())
	return nil
}

func (gm GameManager) GetGame(ctx context.Context) (*service.GameService, error) {
	user, ok := middleware.GetUser(ctx)
	if !ok {
		return nil, fmt.Errorf("There is no user in context")
	}

	// TODO: Add a DB here to extract a "game"
	game, ok := gm.Games[user.ID.String()]
	if !ok {
		return nil, fmt.Errorf("There is no game for user id")
	}
	fmt.Println("[GameManager] GetGame found game for user:", user.ID.String())
	return game, nil
}
