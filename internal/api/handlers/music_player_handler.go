package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/services/game"
	"github.com/FerNunez/NameThatSong/web/templates"
)

type PostPlayPause struct {
	gm *game.GameManager
}

func NewPostPlayPause(gm *game.GameManager) *PostPlayPause {
	return &PostPlayPause{gm}
}

func (h *PostPlayPause) ServeHttp(w http.ResponseWriter, r *http.Request) {

	game, err := h.gm.GetGame(r.Context())
	if err != nil {
		fmt.Printf("error getting game : %v", err)
		return
	}

	err = game.SpotifyApi.PausePlayback(game.SpotifyToken.AccessToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error play game: %v", err), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Playback started"))
}

// /////////////////////////////////////
type PostSkip struct {
	gm *game.GameManager
}

func NewPostSkip(gm *game.GameManager) *PostSkip {
	return &PostSkip{gm}
}

func (h *PostSkip) ServeHttp(w http.ResponseWriter, r *http.Request) {

	game, err := h.gm.GetGame(r.Context())
	if err != nil {
		fmt.Printf("error getting game : %v", err)
		return
	}

	err = game.SkipSong(r.Context())
	if err != nil {
		return
	}
	mp := templates.MusicPlayer(game)
	mp.Render(r.Context(), w)
}

// /////////////////////////////////////
type GetSongTime struct {
	gm *game.GameManager
}

func NewGetSongTime(gm *game.GameManager) *GetSongTime {
	return &GetSongTime{gm}
}

func (h *GetSongTime) ServeHttp(w http.ResponseWriter, r *http.Request) {

	game, err := h.gm.GetGame(r.Context())
	if err != nil {
		fmt.Printf("error getting game : %v", err)
		return
	}
	w.Write([]byte(game.MusicPlayer.GetTimerAsString()))
}
