package handlers

import (
	"net/http"
)

type PostGuessTrack struct {
}

func NewPostGuessTrack() *PostGuessTrack {
	return &PostGuessTrack{}
}

func (h *PostGuessTrack) ServeHttp(w http.ResponseWriter, r *http.Request) {

	// guess := r.FormValue("guess")
	// if guess == "" {
	// 	http.Error(w, "Guess is required", http.StatusBadRequest)
	// 	return
	// }

	// mp := templates.MusicPlayer(game)
	// mp.Render(r.Context(), w)
}
