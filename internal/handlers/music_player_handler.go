package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/manager"
	"github.com/FerNunez/NameThatSong/internal/templates"
)

type PostPlayPause struct {
	gm *manager.GameManager
}

func NewPostPlayPause(gm *manager.GameManager) *PostPlayPause {
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
	gm *manager.GameManager
}

func NewPostSkip(gm *manager.GameManager) *PostSkip {
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
	gm *manager.GameManager
}

func NewGetSongTime(gm *manager.GameManager) *GetSongTime {
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
