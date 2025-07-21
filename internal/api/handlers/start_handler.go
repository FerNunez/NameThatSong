package handlers

import (
	"net/http"
)

type PostStartGame struct {
}

func NewPostStartGame() *PostStartGame {
	return &PostStartGame{}
}

func (h *PostStartGame) ServeHttp(w http.ResponseWriter, r *http.Request) {

	// mp := templates.MusicPlayer(game)
	// mp.Render(r.Context(), w)
}
