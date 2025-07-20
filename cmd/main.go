package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/FerNunez/NameThatSong/internal/config"
	"github.com/FerNunez/NameThatSong/internal/repository"
	"github.com/FerNunez/NameThatSong/internal/repository/database"

	"github.com/FerNunez/NameThatSong/internal/api/handlers"
	"github.com/FerNunez/NameThatSong/internal/services/game"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
	"github.com/redis/go-redis/v9"

	"github.com/joho/godotenv"

	m "github.com/FerNunez/NameThatSong/internal/api/middleware"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

func main() {

	err := godotenv.Load()

	// Redis
	// TODO: check TLS/SSL
	//_ := cache.NewRedisSpotifyCache("127.0.0.1:6379", "", 0, time.Hour)
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // default Redis port
		Password: "",               // no password set
		DB:       0,                // use default DB
	})

	//encription key
	// encryptionKey := os.Getenv("SPOTIFY_TOKEN_ENCRYPTION_KEY")
	// if encryptionKey == "" {
	// 	log.Fatalf("SPOTIFY_TOKEN_ENCRYPTION_KEY empty ")
	// }
	// key, err := base64.StdEncoding.DecodeString(encryptionKey)
	// if err != nil {
	// 	log.Fatalf("SPOTIFY_TOKEN_ENCRYPTION_KEY error %v", err)
	// }
	//
	// tokenEncryptor, err := utils.NewTokenEncryptor(key)
	// if err != nil {
	// 	log.Fatalf("Error creating token encrypto: %v", err)
	// }

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
	userStore := repository.NewSQLUserStore(dbQueries)
	// Email configuration and service
	emailConfig, err := config.NewEmailConfig()
	if err != nil {
		log.Fatalf("Failed to load email config: %v", err)
	}

	emailService, err := user.NewEmailService(emailConfig)
	if err != nil {
		log.Fatalf("Failed to initialize email service: %v", err)
	}
	spotifyConf, err := config.NewSpotifyConfig()
	if err != nil {
		fmt.Println("error while creating spotify config", err)
	}

	// NOTE: This is here for testing locally while developing
	ss, err := spotify.NewSpotifyService(spotifyConf, db, rdb)
	if err != nil {
		fmt.Println("error creating sptofy service:", err)
	}

	gm := game.NewGameManager()

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

		// r.Get("/register", handlers.NewGetRegisterHandler().ServeHttp)
		// r.Post("/register", handlers.NewPostRegisterHandler(dbQueries, tokenEncryptor, gm).ServeHttp)
		// // login Routes
		// r.Get("/login", handlers.NewGetLoginHandler().ServeHttp)
		// r.Post("/login", handlers.NewPostLoginHandler(dbQueries, tokenEncryptor, cookieName, gm).ServeHttp)
		// r.Post("/logout", handlers.NewPostLogoutHandler(cookieName).ServeHTTP)

		// Auth
		r.Get("/spotify-auth", handlers.NewGetAuthHandler(gm).ServeHttp)
		r.Get("/auth/callback", handlers.NewGetAuthCallbackHandler(ss).ServeHttp)

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

	// //spotifyCache := cache.NewSpotifyCacheMap()
	// songProvider := spotify.NewSpotifySongProvider("", "", "", "")
	// accessToken := "BQBwioRufaWQxtoe3lkUwq6DaFlsY8mzWdOW18zhx_-bk047G6Pi4Zhd19c_kHDwgTnbA2p0sy7znDirTqUicySw9v3r7a1FyuWkBphj9Enp1Tt-RV9z468nR3UZgMlxZga15EddZUAtlfnqhV-TfSbdt5zSd2CnQV_-PvhXvQIjO2JKNGuLRjPUnXh4bMKU1Tu5ZT9PZt2vkCwJHFF7Qja5Nifhdf1C6sSOdlf0J_PxCoabg8lMkT0zprhdMkQUHAZGFnU"
	// songProvider.AccessToken = accessToken
	//
	// r.Get("/internal/spotifyapi", handlers.NewGetSpotifyApi().ServeHttp)
	// r.Get("/internal/spotifyapi/track", handlers.NewGetSpotifyApiTrack(accessToken).ServeHttp)
	// r.Get("/internal/spotifyapi/album", handlers.NewGetSpotifyApiAlbum(accessToken).ServeHttp)
	// r.Get("/internal/spotifyapi/artist", handlers.NewGetSpotifyApiArtist(accessToken).ServeHttp)
	// r.Get("/internal/spotifyapi/playlist", handlers.NewGetSpotifyApiPlaylist(accessToken).ServeHttp)
	// r.Get("/internal/spotifyapi/track/name", handlers.NewGetSpotifyApiTrackName(accessToken).ServeHttp)
	// r.Get("/internal/spotifyapi/album/name", handlers.NewGetSpotifyApiAlbumName(accessToken).ServeHttp)
	// r.Get("/internal/spotifyapi/artist/name", handlers.NewGetSpotifyApiArtistName(accessToken).ServeHttp)
	// r.Get("/internal/spotifyapi/playlist/name", handlers.NewGetSpotifyApiPlaylistName(accessToken).ServeHttp)
	//
	// r.Get("/internal/spotifycache", handlers.NewGetSpotifyCache().ServeHttp)
	// r.Get("/internal/spotifycache/track", handlers.NewGetSpotifyCacheTrack(accessToken, redisCache).ServeHttp)
	// r.Get("/internal/spotifycache/album", handlers.NewGetSpotifyCacheAlbum(accessToken, redisCache).ServeHttp)
	// r.Get("/internal/spotifycache/artist", handlers.NewGetSpotifyCacheArtist(accessToken, redisCache).ServeHttp)
	// r.Get("/internal/spotifycache/playlist", handlers.NewGetSpotifyCachePlaylist(accessToken, redisCache).ServeHttp)
	// r.Get("/internal/spotifycache/track/name", handlers.NewGetSpotifyCacheTrackName(accessToken, redisCache).ServeHttp)
	// r.Get("/internal/spotifycache/album/name", handlers.NewGetSpotifyCacheAlbumName(accessToken, redisCache).ServeHttp)
	// r.Get("/internal/spotifycache/artist/name", handlers.NewGetSpotifyCacheArtistName(accessToken, redisCache).ServeHttp)
	// r.Get("/internal/spotifycache/playlist/name", handlers.NewGetSpotifyCachePlaylistName(accessToken, redisCache).ServeHttp)
	//
	// r.Get("/internal/frontend/search-music", handlers.NewGetSearchMusic(accessToken, redisCache).ServeHttp)
	// r.Get("/internal/frontend/stack-music", handlers.NewGetStackMusic(accessToken, redisCache).ServeHttp)
	//
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
