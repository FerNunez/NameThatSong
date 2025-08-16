package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
	"github.com/FerNunez/NameThatSong/internal/services/user"
	"github.com/FerNunez/NameThatSong/web/templates"
)

type GetLoginHandler struct{}

func NewGetLoginHandler() *GetLoginHandler {
	return &GetLoginHandler{}
}

func (h GetLoginHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {

	t := templates.Login("Login")
	err := templates.Layout(t, "NameThanSong").Render(r.Context(), w)

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

	// TODO: Game integration will be added later
	// Create/recreate game for user if needed
	// if _, err := h.GameManager.GetGame(r.Context()); err != nil {
	// 	fmt.Println("[PostLoginHandler] ServeHttp: Recreating game for user", loginResp.User.ID.String())
	//
	// 	// Check for existing Spotify token using SpotifyService
	// 	_, err := h.SpotifyService.GetValidToken(r.Context(), loginResp.User.ID.String())
	// 	if err != nil {
	// 		fmt.Println("[PostLoginHandler] ServeHttp: Spotify token not found for", loginResp.User.ID.String())
	// 	}
	//
	// 	// Create game with SpotifyService
	// 	err = h.GameManager.CreateGame(loginResp.User.ID, h.SpotifyService.GetTokenStore())
	// 	if err != nil {
	// 		fmt.Println("could not create game:", err)
	// 	}
	// }

	// Clear any existing session cookie first
	clearCookie := http.Cookie{
		Name:     h.SessionName,
		Value:    "",
		MaxAge:   -1, // Delete immediately
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	}
	http.SetCookie(w, &clearCookie)

	// Set new session cookie
	cookieValue := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", loginResp.SessionID, loginResp.User.ID.String())))
	cookie := http.Cookie{
		Name:     h.SessionName,
		Value:    cookieValue,
		Expires:  loginResp.ExpiresAt,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	}
	http.SetCookie(w, &cookie)

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}
