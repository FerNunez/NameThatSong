package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/FerNunez/NameThatSong/internal/config"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
	"github.com/FerNunez/NameThatSong/internal/repository"
	"github.com/FerNunez/NameThatSong/internal/repository/database"

	"github.com/FerNunez/NameThatSong/internal/api/handlers"
	"github.com/FerNunez/NameThatSong/internal/services/playlist"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
	"github.com/FerNunez/NameThatSong/internal/services/user"
	"github.com/redis/go-redis/v9"

	"github.com/joho/godotenv"

	m "github.com/FerNunez/NameThatSong/internal/api/middleware"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

func main() {
	// Initialize logger
	logLevel := logger.GetLogLevelFromEnv()
	logger.Init(logLevel)
	logger.Info(nil, "starting NameThatSong application",
		logger.F("log_level", logLevel))

	err := godotenv.Load()

	// Redis
	// TODO: check TLS/SSL
	//_ := cache.NewRedisSpotifyCache("127.0.0.1:6379", "", 0, time.Hour)
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // default Redis port
		Password: "",               // no password set
		DB:       0,                // use default DB
	})

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		fmt.Println("**Please define the DB_RUL in environtment.")
		fmt.Println("Setting dev dbUrl:", dbURL)
		dbURL = "postgres://postgres:postgres@localhost:5432/nts?sslmode=disable"
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		logger.Error(nil, "failed to open database connection",
			logger.F("error", err),
			logger.F("db_url", dbURL))
		log.Fatalf("Error opening db: %v", err)
		return
	}
	logger.Info(nil, "database connection established",
		logger.F("db_url", dbURL))
	// Db and Stores
	dbQueries := database.New(db)
	userStore := repository.NewSQLUserStore(dbQueries)
	emailVerificationStore := repository.NewSQLEmailVerificationStore(dbQueries)
	passwordResetStore := repository.NewSQLPasswordResetStore(dbQueries)
	sessionStore := repository.NewSQLSessionStore(dbQueries)
	playlistStore := repository.NewSQLPlaylistStore(dbQueries)

	// Email configuration and service
	emailConfig, err := config.NewEmailConfig()
	if err != nil {
		log.Fatalf("Failed to load email config: %v", err)
	}
	emailService, err := user.NewEmailService(emailConfig)
	if err != nil {
		log.Fatalf("Failed to initialize email service: %v", err)
	}

	// User service with all dependencies
	userService := user.NewUserService(
		userStore,
		emailVerificationStore,
		passwordResetStore,
		sessionStore,
		emailService,
	)

	spotifyConf, err := config.NewSpotifyConfig()
	if err != nil {
		fmt.Println("error while creating spotify config", err)
		return
	}

	// NOTE: This is here for testing locally while developing
	spotifyService, err := spotify.NewSpotifyService(spotifyConf, db, rdb)
	if err != nil {
		fmt.Println("error creating spotify service:", err)
		return
	}

	// SpotifyService now handles token management internally

	// Playlist service
	playlistService := playlist.NewPlaylistService(playlistStore, spotifyService)

	// Game handler
	gameHandler := handlers.NewGameHandler(playlistService)

	// Create new router
	r := chi.NewRouter()

	cookieName := "CookieName"
	authMiddleware := m.NewAuthMiddleware(userService, cookieName)
	r.Group(func(r chi.Router) {
		r.Use(
			authMiddleware.AddUserToCtxt,
		)
		r.Get("/", handlers.NewGetIndexHandler().ServeHttp)
		r.Get("/modern", handlers.NewModernHandler().ServeHTTP)
		// Set up static file server
		fileServer := http.FileServer(http.Dir("./static"))
		r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

		// User: register, Login & Auth
		r.Get("/register", handlers.NewGetRegisterHandler().ServeHttp)
		r.Post("/register", handlers.NewPostRegisterHandler(userService, spotifyService).ServeHttp)
		r.Get("/login", handlers.NewGetLoginHandler().ServeHttp)
		r.Post("/login", handlers.NewPostLoginHandler(userService, spotifyService, cookieName).ServeHttp)
		r.Post("/logout", handlers.NewPostLogoutHandler(userService, cookieName).ServeHTTP)
		r.Post("/spotify-auth", handlers.NewGetAuthHandler(spotifyService).ServeHttp)
		r.Get("/auth/callback", handlers.NewGetAuthCallbackHandler(spotifyService, userService).ServeHttp)
		r.Get("/connect-spotify", handlers.NewConnectSpotifyHandler().ServeHTTP)

		// Protected routes that require Spotify connection
		r.Group(func(r chi.Router) {
			r.Use(m.RequireSpotifyConnection)

			// Playlist API endpoints
			playlistHandler := handlers.NewPlaylistHandler(playlistService)
			r.Get("/playlists", playlistHandler.GetUserPlaylists)
			r.Post("/playlists", playlistHandler.CreatePlaylist)
			r.Get("/playlists/{id}", playlistHandler.GetPlaylist)
			r.Put("/playlists/{id}", playlistHandler.UpdatePlaylist)
			r.Delete("/playlists/{id}", playlistHandler.DeletePlaylist)
			r.Post("/playlists/{id}/songs", playlistHandler.AddSongToPlaylist)
			r.Delete("/playlists/{id}/songs/{songId}", playlistHandler.RemoveSongFromPlaylist)
			r.Put("/playlists/{id}/songs/reorder", playlistHandler.ReorderPlaylistSongs)
			r.Post("/playlists/import", playlistHandler.ImportFromSpotify)
			r.Post("/playlists/{id}/export", playlistHandler.ExportToSpotify)
			r.Post("/playlists/{id}/sync", playlistHandler.SyncWithSpotify)

			// Debug routes
			r.Get("/debug/search", handlers.NewDebugSearchHandler(spotifyService).ServeHttp)

			// Simple HTMX Search Interface
			r.Get("/search", handlers.NewSearchPageHandler().ServeHTTP)
			r.Get("/api/search", handlers.NewSimpleSearchHandler(spotifyService).ServeHTTP)

			// Action endpoints for the search interface
			actionHandler := handlers.NewActionHandler(spotifyService)
			r.Post("/api/add-track", actionHandler.AddTrackHandler)
			r.Post("/api/play-track", actionHandler.PlayTrackHandler)
			r.Post("/api/add-album", actionHandler.AddAlbumHandler)
			r.Post("/api/play-album", actionHandler.PlayAlbumHandler)
			r.Post("/api/play-artist", actionHandler.PlayArtistHandler)
			r.Post("/api/play-playlist", actionHandler.PlayPlaylistHandler)
			r.Get("/api/artist/{artistId}", handlers.NewArtistDetailHandler(spotifyService).ServeHTTP)

			// Game routes
			r.Get("/game/setup", gameHandler.GameSetupPage)
			r.Get("/api/playlists", gameHandler.GetUserPlaylists)
			r.Post("/game/start", gameHandler.StartGame)
			r.Post("/game/submit-answer", gameHandler.SubmitAnswer)
			r.Get("/game/timer", gameHandler.GetTimer)
			r.Get("/game/results", gameHandler.GameResults)
		})

		// Search
		// r.Get("/search-helper", handlers.NewGetSearchArtists().ServeHttp)
		// r.Get("/search-albums", handlers.NewGetArtistAlbums().ServeHttp)

		// Select
		// r.Post("/api/select-album", handlers.NewPostSelectAlbum().ServeHttp)
		// r.Post("/start-game", handlers.NewPostStartGame().ServeHttp)

		// Guess
		// r.Post("/guess-track", handlers.NewPostGuessTrack().ServeHttp)

		// Player
		// r.Post("/play-pause", handlers.NewPostPlayPause().ServeHttp)
		// r.Post("/skip", handlers.NewPostSkip().ServeHttp)
		// r.Post("/clear-queue", handlers.NewPostClearQueue().ServeHttp)
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
