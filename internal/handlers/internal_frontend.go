package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/cache"
	"github.com/FerNunez/NameThatSong/internal/templates"
)

type GetSearchMusic struct {
	accessToken string
	Cache       cache.SpotifyCache
}

func NewGetSearchMusic(accessToken string, c cache.SpotifyCache) *GetSearchMusic {
	return &GetSearchMusic{accessToken, c}
}
func (h *GetSearchMusic) ServeHttp(w http.ResponseWriter, r *http.Request) {
	fmt.Println("im getting called")
	query := r.URL.Query().Get("search")

	trackList, err := h.Cache.SearchTracks(h.accessToken, query)
	if err != nil {
		fmt.Println("couldnt search track", err)
	}
	albumList, err := h.Cache.SearchAlbums(h.accessToken, query)
	if err != nil {
		fmt.Println("couldnt search album", err)
	}
	artistList, err := h.Cache.SearchArtists(h.accessToken, query)
	if err != nil {
		fmt.Println("couldnt search artist", err)
	}
	playlistList, err := h.Cache.SearchPlaylists(h.accessToken, query)
	if err != nil {
		fmt.Println("couldnt search playlist", err)
	}

	fmt.Printf("got %v, %v, %v, %v\n", len(trackList), len(albumList), len(artistList), len(playlistList))
	component := templates.MusicSearch(trackList, albumList, artistList, playlistList)
	component.Render(r.Context(), w)
}
