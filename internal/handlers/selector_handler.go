package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/manager"
	"github.com/FerNunez/NameThatSong/internal/templates"
)

type PostSelectAlbum struct {
	gm *manager.GameManager
}

func NewPostSelectAlbum(gm *manager.GameManager) *PostSelectAlbum {
	return &PostSelectAlbum{gm}
}

func (h *PostSelectAlbum) ServeHttp(w http.ResponseWriter, r *http.Request) {

	game, err := h.gm.GetGame(r.Context())
	if err != nil {
		fmt.Printf("error getting game : %v", err)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	albumID := r.Form.Get("albumID")
	if albumID == "" {
		http.Error(w, "Album ID is required", http.StatusBadRequest)
		return
	}
	artistID := r.Form.Get("artistID")
	if artistID == "" {
		http.Error(w, "Album ID is required", http.StatusBadRequest)
		return
	}
	fmt.Println("artistID", artistID)

	// Toggle album selection
	toggle := game.ToggleAlbumSelection(albumID, artistID)

	album, ok := game.Cache.AlbumMap[albumID]
	if !ok {
		panic("album should be in cache")
	}
	component := templates.AlbumCard(album, toggle, artistID)
	component.Render(r.Context(), w)
}

//////////////////

type PostClearQueue struct {
	gm *manager.GameManager
}

func NewPostClearQueue(gm *manager.GameManager) *PostClearQueue {
	return &PostClearQueue{gm}
}

func (h *PostClearQueue) ServeHttp(w http.ResponseWriter, r *http.Request) {

	game, err := h.gm.GetGame(r.Context())
	if err != nil {
		fmt.Printf("error getting game : %v", err)
		return
	}
	game.ClearQueue()
	mp := templates.MusicPlayer(game)
	mp.Render(r.Context(), w)
}
