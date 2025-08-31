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
	SessionName    string
}

func NewPostRegisterHandler(userService user.UserService, spotifyService spotify.SpotifyService, sessionName string) *PostRegisterHandler {
	return &PostRegisterHandler{
		UserService:    userService,
		SpotifyService: spotifyService,
		SessionName:    sessionName,
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

	// redirect to connect to spotify
	http.Redirect(w, r, "/connect-spotify", http.StatusFound)
}
