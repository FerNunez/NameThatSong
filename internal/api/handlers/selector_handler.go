package handlers

import (
	"fmt"
	"net/http"
)

type PostSelectAlbum struct {
	//TODO:
}

func NewPostSelectAlbum() *PostSelectAlbum {
	return &PostSelectAlbum{}
}

func (h *PostSelectAlbum) ServeHttp(w http.ResponseWriter, r *http.Request) {

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

}

type PostClearQueue struct {
}

func NewPostClearQueue() *PostClearQueue {
	return &PostClearQueue{}
}

func (h *PostClearQueue) ServeHttp(w http.ResponseWriter, r *http.Request) {

	// TODO
}
