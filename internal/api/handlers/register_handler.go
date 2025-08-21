package handlers

import (
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
	err := templates.AuthLayout(t, "Register - NameThatSong").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

type PostRegisterHandler struct {
	UserService    user.UserService
	SpotifyService spotify.SpotifyService
	// GameManager    *game.GameManager
}

func NewPostRegisterHandler(userService user.UserService, spotifyService spotify.SpotifyService) *PostRegisterHandler {
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
	_, err := h.UserService.Register(r.Context(), registerReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c := templates.RegisterError()
		c.Render(r.Context(), w)
		return
	}

	c := templates.RegisterSuccess()
	err = c.Render(r.Context(), w)
	if err != nil {
		http.Error(w, "error rendering template", http.StatusInternalServerError)
		return
	}

	// TODO: add here redirect to main page?
}
