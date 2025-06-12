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

type GetSpotifyApiTrackName struct {
	accessToken string
	SpotifyApi  *spotify_api.SpotifySongProvider
}

func NewGetSpotifyApiTrackName(accessToken string, s *spotify_api.SpotifySongProvider) *GetSpotifyApiTrackName {
	return &GetSpotifyApiTrackName{accessToken, s}
}
func (h *GetSpotifyApiTrackName) ServeHttp(w http.ResponseWriter, r *http.Request) {
	trackName := "Re forro"
	fetchTrackName, err := h.SpotifyApi.SearchTracksByName(h.accessToken, trackName)
	if err != nil {
		fmt.Println("couldnt fetch tracks", err)
		return
	}
	fmt.Printf("fetchTrackName%+v\n ", fetchTrackName)
	w.Write(fmt.Appendf(nil, "fetchTrackName %#v", fetchTrackName))
}

type GetSpotifyApiAlbumName struct {
	accessToken string
	SpotifyApi  *spotify_api.SpotifySongProvider
}

func NewGetSpotifyApiAlbumName(accessToken string, s *spotify_api.SpotifySongProvider) *GetSpotifyApiAlbumName {
	return &GetSpotifyApiAlbumName{accessToken, s}
}
func (h *GetSpotifyApiAlbumName) ServeHttp(w http.ResponseWriter, r *http.Request) {
	// albumName := "Papota"
	// fetchAlbumName, err := h.SpotifyApi.SearchAlbumsByName(h.accessToken, albumName)
	// if err != nil {
	// 	fmt.Println("couldnt fetch albums", err)
	// 	return
	// }
	// fmt.Printf("fetchAlbumName%+v\n ", fetchAlbumName)
	// w.Write(fmt.Appendf(nil, "fetchAlbumName %#v", fetchAlbumName))
}

type GetSpotifyApiArtistName struct {
	accessToken string
	SpotifyApi  *spotify_api.SpotifySongProvider
}

func NewGetSpotifyApiArtistName(accessToken string, s *spotify_api.SpotifySongProvider) *GetSpotifyApiArtistName {
	return &GetSpotifyApiArtistName{accessToken, s}
}
func (h *GetSpotifyApiArtistName) ServeHttp(w http.ResponseWriter, r *http.Request) {
	artistName := "Lady Gaga"
	fetchArtistName, err := h.SpotifyApi.SearchArtistsByName(h.accessToken, artistName)
	if err != nil {
		fmt.Println("couldnt fetch artists", err)
		return
	}
	fmt.Printf("[fetchArtistName] for %v\n", artistName)
	w.Write(fmt.Appendf(nil, "fetchArtistName %#v", fetchArtistName))
}

type GetSpotifyApiPlaylistName struct {
	accessToken string
	SpotifyApi  *spotify_api.SpotifySongProvider
}

func NewGetSpotifyApiPlaylistName(accessToken string, s *spotify_api.SpotifySongProvider) *GetSpotifyApiPlaylistName {
	return &GetSpotifyApiPlaylistName{accessToken, s}
}
func (h *GetSpotifyApiPlaylistName) ServeHttp(w http.ResponseWriter, r *http.Request) {
	// playlistName := "Re forro"
	// fetchPlaylistName, err := h.SpotifyApi.SearchPlaylistsByName(h.accessToken, playlistName)
	// if err != nil {
	// 	fmt.Println("couldnt fetch playlists", err)
	// 	return
	// }
	// fmt.Printf("fetchPlaylistName%+v\n ", fetchPlaylistName)
	// w.Write(fmt.Appendf(nil, "fetchPlaylistName %#v", fetchPlaylistName))
}
