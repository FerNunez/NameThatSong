package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/manager"
	"github.com/FerNunez/NameThatSong/internal/templates"
)

type PostStartGame struct {
	gm *manager.GameManager
}

func NewPostStartGame(gm *manager.GameManager) *PostStartGame {
	return &PostStartGame{gm}
}

func (h *PostStartGame) ServeHttp(w http.ResponseWriter, r *http.Request) {

	game, err := h.gm.GetGame(r.Context())
	if err != nil {
		fmt.Printf("error getting game : %v", err)
		return
	}

	// Start the game
	err = game.StartGame(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Error starting game: %v", err), http.StatusInternalServerError)
		return
	}

	mp := templates.MusicPlayer(game)
	mp.Render(r.Context(), w)
}
