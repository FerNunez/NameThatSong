package main

import (
	"fmt"
	"goth/internal/handlers"
	"goth/internal/player"
	"goth/internal/provider"
	"goth/internal/service"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get Spotify credentials from environment
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	if clientID == "" || clientSecret == "" {
		log.Fatal("Missing Spotify credentials in .env file")
	}

	// Create new router
	r := chi.NewRouter()

	// Legacy handler (will be used for compatibility until fully migrated)
	legacyApi := handlers.NewSpotifyApi()

	// Create song provider
	songProvider := provider.NewSpotifySongProvider(
		legacyApi.Config.AccessToken,
		legacyApi.Config.RefreshToken,
		clientID,
		clientSecret,
	)

	// Create music player
	musicPlayer := player.NewSpotifyPlayer(legacyApi.Config.DeviceId, legacyApi.Config.AccessToken)

	// Create game service - will be passed to the handler once we modify the NewGameHandler function
	_ = service.NewGameService(musicPlayer, songProvider)

	// Create game handler - in the future, we'll modify this to accept the gameService
	gameHandler, err := handlers.NewGameHandler()
	if err != nil {
		log.Fatalf("Error creating game handler: %v", err)
	}

	// Set up static file server
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Auth routes (still using legacy API for now)
	r.Get("/", handlers.IndexHttp)
	r.Get("/login", legacyApi.RequestUserAuthorizationHandler)
	r.Get("/auth/callback", legacyApi.RequestUserAuthorizationCallbackHandler)

	// New routes using the game handler
	r.Get("/search-helper", gameHandler.SearchArtists)
	r.Get("/search-albums", gameHandler.GetArtistAlbums)
	r.Post("/api/select-album", gameHandler.SelectAlbum)
	r.Post("/start-process", gameHandler.StartGame)
	r.Get("/play", gameHandler.PlayGame)
	r.Get("/guess", gameHandler.MakeGuess)
	r.Get("/skip", gameHandler.SkipSong)

	// Legacy routes (keeping for backward compatibility)
	r.Get("/guess-helper", legacyApi.GuessHelper)
	r.Post("/guess-track", legacyApi.GuessTrack)
	r.Post("/select-track", legacyApi.SelectTrack)
	r.Get("/start", legacyApi.RequestStartHandler)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on http://127.0.0.1:%s\n", port)
	fmt.Println("\n===== Guess The Song Game =====")
	fmt.Println("1. Open your browser and go to: http://127.0.0.1:" + port)
	fmt.Println("2. Log in with your Spotify account")
	fmt.Println("3. Search for an artist and select their albums")
	fmt.Println("4. Click Start to begin the game")
	fmt.Println("5. Guess the songs as they play!\n")

	log.Fatal(http.ListenAndServe("127.0.0.1:"+port, r))
}
