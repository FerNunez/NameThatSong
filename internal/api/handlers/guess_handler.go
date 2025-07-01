package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/services/game"
	"github.com/FerNunez/NameThatSong/web/templates"
)

type PostGuessTrack struct {
	gm *game.GameManager
}

func NewPostGuessTrack(gm *game.GameManager) *PostGuessTrack {
	return &PostGuessTrack{gm}
}

func (h *PostGuessTrack) ServeHttp(w http.ResponseWriter, r *http.Request) {

	game, err := h.gm.GetGame(r.Context())
	if err != nil {
		fmt.Printf("error getting game : %v", err)
		return
	}

	guess := r.FormValue("guess")
	if guess == "" {
		http.Error(w, "Guess is required", http.StatusBadRequest)
		return
	}

	_, err = game.UserGuess(guess)
	if err != nil {
		http.Error(w, "Guess user error", http.StatusBadRequest)
		return
	}

	mp := templates.MusicPlayer(game)
	mp.Render(r.Context(), w)
}
