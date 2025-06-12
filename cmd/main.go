package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"database/sql"

	"github.com/FerNunez/NameThatSong/internal/cache"
	"github.com/FerNunez/NameThatSong/internal/crypto"
	"github.com/FerNunez/NameThatSong/internal/handlers"
	"github.com/FerNunez/NameThatSong/internal/manager"
	"github.com/FerNunez/NameThatSong/internal/spotify_api"
	"github.com/FerNunez/NameThatSong/internal/store"
	"github.com/FerNunez/NameThatSong/internal/store/database"
	"github.com/joho/godotenv"

	m "github.com/FerNunez/NameThatSong/internal/middleware"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {

	err := godotenv.Load()

	// Redis
	// TODO: Change to connection string
	// TODO: check TLS/SSL
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	fmt.Println("redis db created: ", rdb)

	//encription key
	encryptionKey := os.Getenv("SPOTIFY_TOKEN_ENCRYPTION_KEY")

	key, err := base64.StdEncoding.DecodeString(encryptionKey)
	if err != nil {
		log.Fatalf("SPOTIFY_TOKEN_ENCRYPTION_KEY error %v", err)
	}

	tokenEncryptor, err := crypto.NewTokenEncryptor(key)
	if err != nil {
		log.Fatalf("Error creating token encrypto: %v", err)
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		fmt.Println("**Please define the DB_RUL in environtment.")
		fmt.Println("Setting dev dbUrl:", dbURL)
		dbURL = "postgres://postgres:postgres@localhost:5432/nts"
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening db: %v", err)
	}
	dbQueries := database.New(db)
	userStore := store.NewSQLUserStore(dbQueries)

	gm := manager.NewGameManager()

	// Create new router
	r := chi.NewRouter()

	cookieName := "CookieName"
	authMiddleware := m.NewAuthMiddleware(userStore, cookieName)
	r.Group(func(r chi.Router) {
		r.Use(
			authMiddleware.AddUserToCtxt,
		)
		r.Get("/", handlers.NewGetIndexHandler(gm).ServeHttp)
		// Set up static file server
		fileServer := http.FileServer(http.Dir("./static"))
		r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

		r.Get("/register", handlers.NewGetRegisterHandler().ServeHttp)
		r.Post("/register", handlers.NewPostRegisterHandler(dbQueries, tokenEncryptor, gm).ServeHttp)
		// login Routes
		r.Get("/login", handlers.NewGetLoginHandler().ServeHttp)
		r.Post("/login", handlers.NewPostLoginHandler(dbQueries, tokenEncryptor, cookieName, gm).ServeHttp)
		r.Post("/logout", handlers.NewPostLogoutHandler(cookieName).ServeHTTP)

		// Auth
		r.Get("/spotify-auth", handlers.NewGetAuthHandler(gm).ServeHttp)
		r.Get("/auth/callback", handlers.NewGetAuthCallbackHandler(gm).ServeHttp)

		// Search
		r.Get("/search-helper", handlers.NewGetSearchArtists(gm).ServeHttp)
		r.Get("/search-albums", handlers.NewGetArtistAlbums(gm).ServeHttp)

		// Select
		r.Post("/api/select-album", handlers.NewPostSelectAlbum(gm).ServeHttp)
		r.Post("/start-game", handlers.NewPostStartGame(gm).ServeHttp)

		// Guess
		r.Post("/guess-track", handlers.NewPostGuessTrack(gm).ServeHttp)

		// Player
		r.Post("/play-pause", handlers.NewPostPlayPause(gm).ServeHttp)
		r.Post("/skip", handlers.NewPostSkip(gm).ServeHttp)
		r.Post("/clear-queue", handlers.NewPostClearQueue(gm).ServeHttp)

		r.Post("/clear-queue", handlers.NewPostClearQueue(gm).ServeHttp)
		//r.Get("/song-time", handlers.NewGetSongTime(gm).ServeHttp)
	})

	spotifyCache := cache.NewSpotifyCacheMap()
	songProvider := spotify_api.NewSpotifySongProvider("", "", "", "")
	accessToken := "TODO"
	songProvider.AccessToken = accessToken

	r.Get("/internal/spotifyapi", handlers.NewGetSpotifyApi().ServeHttp)
	r.Get("/internal/spotifyapi/track", handlers.NewGetSpotifyApiTrack(accessToken, songProvider).ServeHttp)
	r.Get("/internal/spotifyapi/album", handlers.NewGetSpotifyApiAlbum(accessToken, songProvider).ServeHttp)
	r.Get("/internal/spotifyapi/artist", handlers.NewGetSpotifyApiArtist(accessToken, songProvider).ServeHttp)
	r.Get("/internal/spotifyapi/playlist", handlers.NewGetSpotifyApiPlaylist(accessToken, songProvider).ServeHttp)
	r.Get("/internal/spotifyapi/track/name", handlers.NewGetSpotifyApiTrackName(accessToken, songProvider).ServeHttp)
	r.Get("/internal/spotifyapi/album/name", handlers.NewGetSpotifyApiAlbumName(accessToken, songProvider).ServeHttp)
	r.Get("/internal/spotifyapi/artist/name", handlers.NewGetSpotifyApiArtistName(accessToken, songProvider).ServeHttp)
	r.Get("/internal/spotifyapi/playlist/name", handlers.NewGetSpotifyApiPlaylistName(accessToken, songProvider).ServeHttp)

	r.Get("/internal/spotifycache", handlers.NewGetSpotifyCache().ServeHttp)
	r.Get("/internal/spotifycache/track", handlers.NewGetSpotifyCacheTrack(accessToken, songProvider, spotifyCache).ServeHttp)
	r.Get("/internal/spotifycache/album", handlers.NewGetSpotifyCacheAlbum(accessToken, songProvider, spotifyCache).ServeHttp)
	r.Get("/internal/spotifycache/artist", handlers.NewGetSpotifyCacheArtist(accessToken, songProvider, spotifyCache).ServeHttp)
	r.Get("/internal/spotifycache/playlist", handlers.NewGetSpotifyCachePlaylist(accessToken, songProvider, spotifyCache).ServeHttp)

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
