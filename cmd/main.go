package main

import (
	"goth/internal/handlers"
	"net/http"
	"github.com/go-chi/chi/v5"
)

func main() {

	r := chi.NewRouter()
	cfg := handlers.NewSpotifyApi()
	fileServer := http.FileServer(http.Dir("./static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	r.Get("/", handlers.IndexHttp)
	r.Get("/login", cfg.RequestUserAuthorizationHandler)
	r.Get("/auth/callback", cfg.RequestUserAuthorizationCallbackHandler)

	r.Get("/search", cfg.RequestArtistListByNameHandler)
	r.Get("/search-albums", cfg.AlbumGridHttp)
	r.Post("/start-process", cfg.StartProcess)

	r.Get("/start", cfg.RequestStartHandler)

	server := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8080",
	}

	server.ListenAndServe()
}
