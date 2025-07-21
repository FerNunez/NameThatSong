package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/api/middleware"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
)

type GetAuthHandler struct {
	spotifyService spotify.SpotifyService
}

func NewGetAuthHandler(ss spotify.SpotifyService) *GetAuthHandler {
	return &GetAuthHandler{ss}

}
func (h *GetAuthHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		// TODO: add here a way to say "you need to log in"
		return
	}

	urlString, err := h.spotifyService.AuthRequestURL(user.ID.String())
	if err != nil {
		http.Error(w, fmt.Sprintf("error generating the auth request url: %v", err), http.StatusBadRequest)
		return
	}
	// Redirect to Spotify
	w.Header().Set("HX-Redirect", urlString)
}

// //////////////////////////////////////
type GetAuthCallbackHandler struct {
	spotifyService spotify.SpotifyService
}

func NewGetAuthCallbackHandler(ss spotify.SpotifyService) *GetAuthCallbackHandler {
	return &GetAuthCallbackHandler{ss}

}
func (h *GetAuthCallbackHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		// TODO: add here a way to say "you need to log in"
		return
	}
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	tr, err := h.spotifyService.TokenExchange(r.Context(), user.ID.String(), code, state)
	if err != nil {
		fmt.Printf("error exchanging token: %v\n", err)
		http.Error(w, "error exchanging spotify token", http.StatusBadRequest)
		return
	}

	fmt.Printf("tr: %v", tr)
	http.Redirect(w, r, "/", http.StatusFound)
}
