package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"database/sql"

	"github.com/FerNunez/NameThatSong/internal/handlers"
	"github.com/FerNunez/NameThatSong/internal/manager"
	"github.com/FerNunez/NameThatSong/internal/store"
	"github.com/FerNunez/NameThatSong/internal/store/database"
	"github.com/joho/godotenv"

	m "github.com/FerNunez/NameThatSong/internal/middleware"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

func main() {

	gm := manager.NewGameManager()

	err := godotenv.Load()
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

	// Create new router
	r := chi.NewRouter()

	// Create game handler
	// gameHandler, err := handlers.NewGameHandler()
	// if err != nil {
	// 	log.Fatalf("Error creating game handler: %v", err)
	// }

	cookieName := "CookieName"
	// sessionStore := dbstore.NewSessionStore()
	//sessionStore.CreateSession()
	authMiddleware := m.NewAuthMiddleware(userStore, cookieName)
	r.Group(func(r chi.Router) {
		r.Use(
			authMiddleware.AddUserToContext,
		)
		r.Get("/", handlers.NewGetIndexHandler(gm).ServeHttp)
		// Set up static file server
		fileServer := http.FileServer(http.Dir("./static"))
		r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

		r.Get("/register", handlers.NewGetRegisterHandler().ServeHttp)
		r.Post("/register", handlers.NewPostRegisterHandler(dbQueries, gm).ServeHttp)
		// login Routes
		r.Get("/login", handlers.NewGetLoginHandler().ServeHttp)
		r.Post("/login", handlers.NewPostLoginHandler(dbQueries, cookieName, gm).ServeHttp)
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

		r.Get("/song-time", handlers.NewGetSongTime(gm).ServeHttp)

	})

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
