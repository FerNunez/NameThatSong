package main

import (
	"fmt"
	"goth/internal/handlers"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

func main() {

	// Create new router
	r := chi.NewRouter()

	// Create game handler
	gameHandler, err := handlers.NewGameHandler()
	if err != nil {
		log.Fatalf("Error creating game handler: %v", err)
	}

	// Set up static file server
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Routes
	r.Get("/", gameHandler.IndexHandler)
	r.Get("/login", gameHandler.AuthHandler)
	r.Get("/auth/callback", gameHandler.AuthCallbackHandler)

	// Game routes
	r.Get("/search-helper", gameHandler.SearchArtists)
	r.Get("/search-albums", gameHandler.GetArtistAlbums)
	r.Post("/api/select-album", gameHandler.SelectAlbum)
	r.Post("/start-process", gameHandler.StartGame)
	r.Get("/play", gameHandler.PlayGame)
	r.Get("/guess-helper", gameHandler.GuessHelper)
	r.Post("/guess-track", gameHandler.GuessTrack)
	r.Post("/select-track", gameHandler.SelectTrack)
	r.Put("/skip", gameHandler.SkipSong)

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
	fmt.Println("5. Guess the songs as they play!")

	log.Fatal(http.ListenAndServe("127.0.0.1:"+port, r))
}
