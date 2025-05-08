package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/FerNunez/NameThatSong/internal/auth"
	"github.com/FerNunez/NameThatSong/internal/manager"
	"github.com/FerNunez/NameThatSong/internal/store"
	"github.com/FerNunez/NameThatSong/internal/store/database"
	"github.com/FerNunez/NameThatSong/internal/templates"
)

type GetRegisterHandler struct {
}

func NewGetRegisterHandler() *GetRegisterHandler {
	return &GetRegisterHandler{}
}
func (h GetRegisterHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {
	t := templates.RegisterPage()
	err := templates.Layout(t, "NameThanSong").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

type PostRegisterHandler struct {
	UserStore         store.UserStore
	SpotifyTokenStore store.SpotifyTokenStore
	//dbQuery *database.Queries
	// sessionStore      store.SessionStore
	// passwordhash      hash.PasswordHash
	// sessionCookieName string
	GameManager *manager.GameManager
}

func NewPostRegisterHandler(dbQuery *database.Queries, gm *manager.GameManager) *PostRegisterHandler {
	return &PostRegisterHandler{
		UserStore:         store.NewSQLUserStore(dbQuery),
		SpotifyTokenStore: store.NewSQLSpotifyTokenStore(dbQuery),
		GameManager:       gm,
	}
}

func (h PostRegisterHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {

	email := r.FormValue("email")
	password := r.FormValue("password")

	// hash password
	hashedPass, err := auth.HashPassword(password)

	// add user to DB
	dbUser, err := h.UserStore.Create(r.Context(), email, hashedPass)
	if err != nil {
		fmt.Println("error creating user err:", err)
		w.WriteHeader(http.StatusBadRequest)
		c := templates.RegisterError()
		c.Render(r.Context(), w)
		return
	}

	err = h.GameManager.CreateGame(dbUser.ID, h.SpotifyTokenStore)
	if err != nil {
		fmt.Println("could not create game", err)
		w.WriteHeader(http.StatusInternalServerError)
		c := templates.RegisterError()
		c.Render(r.Context(), w)
		return
	}
	fmt.Println("Game added for:", dbUser.ID.String())

	// add a empty spotify token for user
	err = h.SpotifyTokenStore.Create(r.Context(), dbUser.ID, "", "", "", "", time.Now())
	if err != nil {
		fmt.Println("Could not create Empty Spotify Token: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		c := templates.RegisterError()
		c.Render(r.Context(), w)
		return
	}
	fmt.Println("Token Created for", dbUser.ID.String())

	c := templates.RegisterSuccess()
	err = c.Render(r.Context(), w)

	if err != nil {
		http.Error(w, "error rendering template", http.StatusInternalServerError)
		return
	}

	// TODO: add here redirect to main page?
}
