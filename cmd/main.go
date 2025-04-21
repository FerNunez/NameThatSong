package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/FerNunez/NameThatSong/internal/handlers"
	"github.com/FerNunez/NameThatSong/internal/store/dbstore"

	m "github.com/FerNunez/NameThatSong/internal/middleware"
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

	sessionStore := dbstore.NewSessionStore()
	//sessionStore.CreateSession()
	authMiddleware := m.NewAuthMiddleware(sessionStore, "test")
	r.Group(func(r chi.Router) {
		r.Use(
			authMiddleware.CreateTempUser,
		)
		r.Get("/", gameHandler.IndexHandler)

	})
	r.Group(func(r chi.Router) {
		r.Use(
			authMiddleware.AddUserToContext,
		)
		// Set up static file server
		fileServer := http.FileServer(http.Dir("./static"))
		r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

		// Auth Routes
		r.Get("/login", gameHandler.AuthHandler)
		r.Get("/auth/callback", gameHandler.AuthCallbackHandler)

		// Game routes
		r.Get("/search-helper", gameHandler.SearchArtists)
		r.Get("/search-albums", gameHandler.GetArtistAlbums)
		r.Post("/api/select-album", gameHandler.SelectAlbum)
		r.Post("/start-process", gameHandler.StartGame)
		//r.Get("/play", gameHandler.PlayGame)
		//TODO: add only songs of artists here
		//r.Get("/guess-helper", gameHandler.GuessHelper)
		r.Post("/guess-track", gameHandler.GuessTrack)
		//r.Post("/select-track", gameHandler.SelectTrack)
		r.Put("/skip", gameHandler.SkipSong)
		r.Post("/clear-queue", gameHandler.ClearQueue)
	})

	r.Get("/song-time", gameHandler.SongTime)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	address := os.Getenv("ADDRESS")
	if address == "" {
		address = "127.0.0.1"
	}
	fmt.Printf("Server starting on http://%s:%s\n", address, port)

	//log.Fatal(http.ListenAndServe("127.0.0.1:"+port, r))
	log.Fatal(http.ListenAndServe(address+":"+port, r))
}
