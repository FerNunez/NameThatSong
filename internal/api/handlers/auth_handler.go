package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/services/game"
)

type GetAuthHandler struct {
	gm *game.GameManager
}

func NewGetAuthHandler(gm *game.GameManager) *GetAuthHandler {
	return &GetAuthHandler{gm}

}
func (h *GetAuthHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {
	game, err := h.gm.GetGame(r.Context())
	if err != nil {
		fmt.Printf("[GetAuthHandler] could not retriever game: %v", err)
		return
	}
	urlString, err := game.RequestUserAuthoritazion()
	if err != nil {
		fmt.Printf("[GetAuthHandler] error getting spotify auth: %v", err)
		return
	}
	// Redirect to Spotify
	w.Header().Set("HX-Redirect", urlString)
}

// //////////////////////////////////////
type GetAuthCallbackHandler struct {
	gm *game.GameManager
}

func NewGetAuthCallbackHandler(gm *game.GameManager) *GetAuthCallbackHandler {
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
	err = game.ExchangeToken(r.Context(), state, code)
	if err != nil {
		fmt.Printf("error exchanging token: %v\n", err)
		http.Error(w, "error exchanging spotify token", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
