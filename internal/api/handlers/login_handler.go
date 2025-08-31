package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
	"github.com/FerNunez/NameThatSong/internal/services/user"
	"github.com/FerNunez/NameThatSong/internal/utils"
	"github.com/FerNunez/NameThatSong/web/templates"
)

type GetLoginHandler struct{}

func NewGetLoginHandler() *GetLoginHandler {
	return &GetLoginHandler{}
}

func (h GetLoginHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {

	t := templates.Login("Login")
	err := templates.AuthLayout(t, "Login - NameThatSong").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}

}

type PostLoginHandler struct {
	UserService    user.UserService
	SpotifyService spotify.SpotifyService
	SessionName    string
	// GameManager       *game.GameManager
}

func NewPostLoginHandler(userService user.UserService, spotifyService spotify.SpotifyService, sessionName string) *PostLoginHandler {
	return &PostLoginHandler{
		UserService:    userService,
		SpotifyService: spotifyService,
		SessionName:    sessionName,
		// GameManager:       gm,
	}
}

func (h PostLoginHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {

	email := r.FormValue("email")
	password := r.FormValue("password")

	// Create login request
	loginReq := models.LoginRequest{
		Email:    email,
		Password: password,
	}

	// Login user through service
	loginResp, err := h.UserService.Login(r.Context(), loginReq)
	if err != nil {
		fmt.Println("login failed:", err)
		w.WriteHeader(http.StatusUnauthorized)
		c := templates.LoginError()
		c.Render(r.Context(), w)
		return
	}

	// Clear any existing session cookie first and set new
	clearCookie := utils.ClearCookie(h.SessionName)
	http.SetCookie(w, &clearCookie)
	newCookie := utils.GenerateCookie(loginResp.SessionID, loginResp.User.ID.String(), h.SessionName, loginResp.ExpiresAt)
	http.SetCookie(w, &newCookie)

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}
