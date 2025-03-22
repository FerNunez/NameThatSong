package main

import (
	"github.com/go-chi/chi/v5"
	"goth/internal/handlers"
	"net/http"
)

func main() {

	r := chi.NewRouter()
	cfg := handlers.NewSpotifyApi()
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	r.Get("/", handlers.IndexHttp)
	r.Get("/login", cfg.RequestUserAuthorizationHandler)
	r.Get("/auth/callback", cfg.RequestUserAuthorizationCallbackHandler)

	r.Post("/api/select-album", cfg.AlbumSelection)

	r.Get("/search-helper", cfg.SearchHelper)
	r.Get("/search-albums", cfg.AlbumGridHttp)
	r.Post("/start-process", cfg.StartProcess)

	r.Get("/start", cfg.RequestStartHandler)

	r.Get("/guess-helper", cfg.GuessHelper)
	r.Post("/guess-track", cfg.GuessTrack)
	//r.PUT("/select-track", cfg.GuessTrack)
	server := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8080",
	}

	server.ListenAndServe()
}
