package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/spotify_api"

	"github.com/FerNunez/NameThatSong/internal/cache"
	"github.com/FerNunez/NameThatSong/internal/templates"
)

type GetSpotifyCache struct {
}

func NewGetSpotifyCache() *GetSpotifyCache {
	return &GetSpotifyCache{}
}

func (h *GetSpotifyCache) ServeHttp(w http.ResponseWriter, r *http.Request) {
	component := templates.InternalSpotifyCache()
	component.Render(r.Context(), w)

}

type GetSpotifyCacheTrack struct {
	accessToken string
	SpotifyApi  *spotify_api.SpotifySongProvider
	Cache       cache.SpotifyCache
}

func NewGetSpotifyCacheTrack(accessToken string, s *spotify_api.SpotifySongProvider, c cache.SpotifyCache) *GetSpotifyCacheTrack {
	return &GetSpotifyCacheTrack{accessToken, s, c}
}
func (h *GetSpotifyCacheTrack) ServeHttp(w http.ResponseWriter, r *http.Request) {
	trackID := "11dFghVXANMlKmJXsNCbNl"
	query := r.URL.Query().Get("id")
	if query != "" {
		trackID = query
	}

	fetchTrack, err := h.Cache.GetTrack(h.SpotifyApi, h.accessToken, trackID)
	if err != nil {
		fmt.Println("couldnt fetch tracks", err)
		return
	}
	fmt.Printf("fetchTrack%+v\n", fetchTrack)
	w.Write(fmt.Appendf(nil, "fetchTrack %#v", fetchTrack))
}

type GetSpotifyCacheAlbum struct {
	accessToken string
	SpotifyApi  *spotify_api.SpotifySongProvider
	Cache       cache.SpotifyCache
}

func NewGetSpotifyCacheAlbum(accessToken string, s *spotify_api.SpotifySongProvider, c cache.SpotifyCache) *GetSpotifyCacheAlbum {
	return &GetSpotifyCacheAlbum{accessToken, s, c}
}
func (h *GetSpotifyCacheAlbum) ServeHttp(w http.ResponseWriter, r *http.Request) {
	albumID := "4aawyAB9vmqN3uQ7FjRGTy"
	query := r.URL.Query().Get("id")
	if query != "" {
		albumID = query
	}

	fetchAlbum, err := h.Cache.GetAlbum(h.SpotifyApi, h.accessToken, albumID)
	if err != nil {
		fmt.Println("couldnt fetch albums", err)
		return
	}

	w.Write(fmt.Appendf(nil, "fetchAlbum %#v with \n", fetchAlbum))
}

type GetSpotifyCacheArtist struct {
	accessToken string
	SpotifyApi  *spotify_api.SpotifySongProvider
	Cache       cache.SpotifyCache
}

func NewGetSpotifyCacheArtist(accessToken string, s *spotify_api.SpotifySongProvider, c cache.SpotifyCache) *GetSpotifyCacheArtist {
	return &GetSpotifyCacheArtist{accessToken, s, c}
}
func (h *GetSpotifyCacheArtist) ServeHttp(w http.ResponseWriter, r *http.Request) {
	artistID := "0TnOYISbd1XYRBk9myaseg"
	query := r.URL.Query().Get("id")
	if query != "" {
		artistID = query
	}
	fetchArtist, err := h.Cache.GetArtist(h.SpotifyApi, h.accessToken, artistID)
	if err != nil {
		fmt.Println("couldnt fetch artists", err)
		return
	}
	w.Write(fmt.Appendf(nil, "fetchArtist %#v", fetchArtist))
}

type GetSpotifyCachePlaylist struct {
	accessToken string
	SpotifyApi  *spotify_api.SpotifySongProvider
	Cache       cache.SpotifyCache
}

func NewGetSpotifyCachePlaylist(accessToken string, s *spotify_api.SpotifySongProvider, c cache.SpotifyCache) *GetSpotifyCachePlaylist {
	return &GetSpotifyCachePlaylist{accessToken, s, c}
}
func (h *GetSpotifyCachePlaylist) ServeHttp(w http.ResponseWriter, r *http.Request) {
	playlistID := "3cEYpjA9oz9GiPac4AsH4n"
	query := r.URL.Query().Get("id")
	if query != "" {
		playlistID = query
	}
	fetchPlaylist, err := h.Cache.GetPlaylist(h.SpotifyApi, h.accessToken, playlistID)
	if err != nil {
		fmt.Println("couldnt fetch playlists", err)
		return
	}
	w.Write(fmt.Appendf(nil, "fetchplaylist %#v", fetchPlaylist))
}
