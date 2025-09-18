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
	playlistStore := repository.NewSQLPlaylistStore(dbQueries, db)

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
		r.Get("/", handlers.NewModernHandler().ServeHTTP)
		// Legacy route removed - modern UI is now the default
		// Set up static file server
		fileServer := http.FileServer(http.Dir("./static"))
		r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

		// User: register, Login & Auth
		r.Get("/register", handlers.NewGetRegisterHandler().ServeHttp)
		r.Post("/register", handlers.NewPostRegisterHandler(userService, spotifyService, cookieName).ServeHttp)
		r.Get("/login", handlers.NewGetLoginHandler().ServeHttp)
		r.Post("/login", handlers.NewPostLoginHandler(userService, spotifyService, cookieName).ServeHttp)
		r.Post("/logout", handlers.NewPostLogoutHandler(userService, cookieName).ServeHTTP)
		r.Post("/spotify-auth", handlers.NewGetAuthHandler(spotifyService).ServeHttp)
		r.Get("/auth/callback", handlers.NewGetAuthCallbackHandler(spotifyService, userService, playlistService).ServeHttp)
		r.Get("/connect-spotify", handlers.NewConnectSpotifyHandler().ServeHTTP)

		// Protected routes that require Spotify connection
		r.Group(func(r chi.Router) {
			r.Use(m.RequireSpotifyConnection)

			// Playlist API endpoints
			playlistHandler := handlers.NewPlaylistHandlerWithSpotify(playlistService, spotifyService)
			r.Get("/api/events/playlist-updates", playlistHandler.HandlePlaylistEvents)
			r.Get("/api/import-spotify-playlists", playlistHandler.GetSpotifyPlaylistsForImport)
			r.Get("/api/local-playlists", playlistHandler.GetLocalPlaylists)
			r.Put("/api/spotify-playlists/{id}/update", playlistHandler.UpdateSpotifyPlaylist)
			r.Put("/api/spotify-playlists/refresh", playlistHandler.RefreshSpotifyPlaylists)
			r.Get("/api/playlist/create-form", playlistHandler.ShowCreatePlaylistForm)
			r.Post("/api/local-playlist", playlistHandler.CreateAndShowPlaylist)
			r.Get("/api/playlist/cancel-create", playlistHandler.CancelCreatePlaylist)

			// Game routes (integrated into modern UI)
			r.Get("/api/game/setup", playlistHandler.ShowGameSetup)
			r.Get("/api/game/playlists", playlistHandler.GetGamePlaylists)
			r.Post("/game/start", gameHandler.StartGame)
			r.Post("/game/submit-answer", gameHandler.SubmitAnswer)
			r.Get("/game/timer", gameHandler.GetTimer)
			r.Get("/game/results", gameHandler.GameResults)

			// Utility endpoints
			r.Get("/api/playlist-songs-empty", playlistHandler.ShowPlaylistSongsEmpty)
			r.Get("/api/playlist/{id}/songs", playlistHandler.GetPlaylistSongsView)

			musicSearchHandler := handlers.NewMusicSearchHandler(spotifyService)
			r.Get("/api/music-search", musicSearchHandler.SearchAll)
			r.Get("/api/music-search/artist/{id}", musicSearchHandler.SearchArtistItems)
			r.Get("/api/music-search/album/{id}", musicSearchHandler.SearchAlbumItems)

			// Playlist context and track addition routes
			r.Get("/api/set-playlist-context", playlistHandler.SetPlaylistContext)
			r.Post("/api/add-to-current-playlist", playlistHandler.AddToCurrentPlaylist)
		})

	})

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
