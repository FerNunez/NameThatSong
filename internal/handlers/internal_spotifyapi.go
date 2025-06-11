package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/spotify_api"
	"github.com/FerNunez/NameThatSong/internal/templates"
)

type GetSpotifyApi struct {
}

func NewGetSpotifyApi() *GetSpotifyApi {
	return &GetSpotifyApi{}
}

func (h *GetSpotifyApi) ServeHttp(w http.ResponseWriter, r *http.Request) {
	component := templates.InternalSpotifyApi()
	component.Render(r.Context(), w)

}

type GetSpotifyApiTrack struct {
	accessToken string
	SpotifyApi  *spotify_api.SpotifySongProvider
}

func NewGetSpotifyApiTrack(accessToken string, s *spotify_api.SpotifySongProvider) *GetSpotifyApiTrack {
	return &GetSpotifyApiTrack{accessToken, s}
}
func (h *GetSpotifyApiTrack) ServeHttp(w http.ResponseWriter, r *http.Request) {
	tracklID := "11dFghVXANMlKmJXsNCbNl"
	fetchTrack, err := h.SpotifyApi.FetchTrack(h.accessToken, tracklID)
	if err != nil {
		fmt.Println("couldnt fetch tracks", err)
		return
	}
	fmt.Printf("fetchTrack%+v\n ", fetchTrack)
	w.Write(fmt.Appendf(nil, "fetchTrack %#v", fetchTrack))
}

type GetSpotifyApiAlbum struct {
	accessToken string
	SpotifyApi  *spotify_api.SpotifySongProvider
}

func NewGetSpotifyApiAlbum(accessToken string, s *spotify_api.SpotifySongProvider) *GetSpotifyApiAlbum {
	return &GetSpotifyApiAlbum{accessToken, s}
}
func (h *GetSpotifyApiAlbum) ServeHttp(w http.ResponseWriter, r *http.Request) {
	albumID := "4aawyAB9vmqN3uQ7FjRGTy"
	fetchAlbum, tracks, err := h.SpotifyApi.FetchAlbum(h.accessToken, albumID)
	if err != nil {
		fmt.Println("couldnt fetch albums", err)
		return
	}
	w.Write(fmt.Appendf(nil, "fetchAlbum %#v with \n **tracks: %#v ", fetchAlbum, tracks))
}

type GetSpotifyApiArtist struct {
	accessToken string
	SpotifyApi  *spotify_api.SpotifySongProvider
}

func NewGetSpotifyApiArtist(accessToken string, s *spotify_api.SpotifySongProvider) *GetSpotifyApiArtist {
	return &GetSpotifyApiArtist{accessToken, s}
}
func (h *GetSpotifyApiArtist) ServeHttp(w http.ResponseWriter, r *http.Request) {
	artistID := "0TnOYISbd1XYRBk9myaseg"
	fetchArtist, err := h.SpotifyApi.FetchArtist(h.accessToken, artistID)
	if err != nil {
		fmt.Println("couldnt fetch artists", err)
		return
	}
	w.Write(fmt.Appendf(nil, "fetchArtist %#v", fetchArtist))
}

type GetSpotifyApiPlaylist struct {
	accessToken string
	SpotifyApi  *spotify_api.SpotifySongProvider
}

func NewGetSpotifyApiPlaylist(accessToken string, s *spotify_api.SpotifySongProvider) *GetSpotifyApiPlaylist {
	return &GetSpotifyApiPlaylist{accessToken, s}
}
func (h *GetSpotifyApiPlaylist) ServeHttp(w http.ResponseWriter, r *http.Request) {
	playlistID := "3cEYpjA9oz9GiPac4AsH4n"
	fetchPlaylist, fetchTracks, err := h.SpotifyApi.FetchPlaylist(h.accessToken, playlistID)
	if err != nil {
		fmt.Println("couldnt fetch playlists", err)
		return
	}
	w.Write(fmt.Appendf(nil, "fetchplaylist %#v, with tracks: %#v", fetchPlaylist, fetchTracks))
}
