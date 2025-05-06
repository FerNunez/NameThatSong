package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/manager"
)

type GetAuthHandler struct {
	gm *manager.GameManager
}

func NewGetAuthHandler(gm *manager.GameManager) *GetAuthHandler {
	return &GetAuthHandler{gm}

}
func (h *GetAuthHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Hello getting called")
	game, err := h.gm.GetGame(r.Context())
	if err != nil {
		fmt.Printf("eror getting game: %v", err)
		http.Error(w, "error generating state", http.StatusBadRequest)
		return
	}

	urlString, err := game.RequestUserAuthoritazion()
	if err != nil {
		fmt.Printf("error getting auth: %v", err)
		http.Error(w, "error generating state", http.StatusBadRequest)
		return
	}
	// Redirect to Spotify
	w.Header().Set("HX-Redirect", urlString)
}

// //////////////////////////////////////
type GetAuthCallbackHandler struct {
	gm *manager.GameManager
}

func NewGetAuthCallbackHandler(gm *manager.GameManager) *GetAuthCallbackHandler {
	return &GetAuthCallbackHandler{gm}

}
func (h *GetAuthCallbackHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {
	game, err := h.gm.GetGame(r.Context())
	if err != nil {
		fmt.Printf("error generating state: %v\n", err)
		http.Error(w, "error generating state", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")
	err = game.ExchangeToken(state, code)
	if err != nil {
		fmt.Printf("error exchanging token: %v\n", err)
		http.Error(w, "error exchanging spotify token", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
