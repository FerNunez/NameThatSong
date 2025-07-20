package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
	"github.com/FerNunez/NameThatSong/internal/services/user"
	"github.com/FerNunez/NameThatSong/web/templates"
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
	UserService    user.UserService
	SpotifyService spotify.Spotify
	// GameManager    *game.GameManager
}

func NewPostRegisterHandler(userService user.UserService, spotifyService spotify.Spotify) *PostRegisterHandler {
	return &PostRegisterHandler{
		UserService:    userService,
		SpotifyService: spotifyService,
		// GameManager:    gm,
	}
}

func (h PostRegisterHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {

	email := r.FormValue("email")
	password := r.FormValue("password")
	displayName := r.FormValue("display_name")

	// Create registration request
	registerReq := models.RegisterRequest{
		Email:       email,
		Password:    password,
		DisplayName: displayName,
	}

	// Register user through service
	dbUser, err := h.UserService.Register(r.Context(), registerReq)
	if err != nil {
		fmt.Println("error registering user:", err)
		w.WriteHeader(http.StatusBadRequest)
		c := templates.RegisterError()
		c.Render(r.Context(), w)
		return
	}

	// TODO: Game integration will be added later
	// Create game for user
	// err = h.GameManager.CreateGame(dbUser.ID, h.SpotifyService.GetTokenStore())
	// if err != nil {
	// 	fmt.Println("could not create game", err)
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	c := templates.RegisterError()
	// 	c.Render(r.Context(), w)
	// 	return
	// }

	c := templates.RegisterSuccess()
	err = c.Render(r.Context(), w)
	if err != nil {
		http.Error(w, "error rendering template", http.StatusInternalServerError)
		return
	}

	// TODO: add here redirect to main page?
}
