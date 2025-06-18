package handlers

import (
	"fmt"
	"net/http"
	"strings"

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
	Cache       cache.SpotifyCache
}

func NewGetSpotifyCacheTrack(accessToken string, c cache.SpotifyCache) *GetSpotifyCacheTrack {
	return &GetSpotifyCacheTrack{accessToken, c}
}
func (h *GetSpotifyCacheTrack) ServeHttp(w http.ResponseWriter, r *http.Request) {
	trackID := "11dFghVXANMlKmJXsNCbNl"
	query := r.URL.Query().Get("id")
	if query != "" {
		trackID = query
	}

	fetchTrack, err := h.Cache.GetTrack(h.accessToken, trackID)
	if err != nil {
		fmt.Println("couldnt fetch tracks", err)
		return
	}
	fmt.Printf("[fetchTrack] for %+v\n", trackID)
	w.Write(fmt.Appendf(nil, "fetchTrack %#v", fetchTrack))
}

type GetSpotifyCacheAlbum struct {
	accessToken string
	Cache       cache.SpotifyCache
}

func NewGetSpotifyCacheAlbum(accessToken string, c cache.SpotifyCache) *GetSpotifyCacheAlbum {
	return &GetSpotifyCacheAlbum{accessToken, c}
}
func (h *GetSpotifyCacheAlbum) ServeHttp(w http.ResponseWriter, r *http.Request) {
	albumID := "4aawyAB9vmqN3uQ7FjRGTy"
	query := r.URL.Query().Get("id")
	if query != "" {
		albumID = query
	}

	fetchAlbum, err := h.Cache.GetAlbum(h.accessToken, albumID)
	if err != nil {
		fmt.Println("couldnt fetch albums", err)
		return
	}
	trackList, err := h.Cache.GetTracksFromAlbum(h.accessToken, albumID)
	if err != nil {
		fmt.Println("couldnt fetch tracks from albums: ", err)
		return
	}
	fmt.Printf("[cacheAlbum] for %+v\n", albumID)
	w.Write(fmt.Appendf(nil, "fetchAlbum %#v with tracks: %#v ", fetchAlbum, strings.Join(trackList, ", ")))
}

type GetSpotifyCacheArtist struct {
	accessToken string
	Cache       cache.SpotifyCache
}

func NewGetSpotifyCacheArtist(accessToken string, c cache.SpotifyCache) *GetSpotifyCacheArtist {
	return &GetSpotifyCacheArtist{accessToken, c}
}
func (h *GetSpotifyCacheArtist) ServeHttp(w http.ResponseWriter, r *http.Request) {
	artistID := "0TnOYISbd1XYRBk9myaseg"
	query := r.URL.Query().Get("id")
	if query != "" {
		artistID = query
	}
	fetchArtist, err := h.Cache.GetArtist(h.accessToken, artistID)
	if err != nil {
		fmt.Println("couldnt fetch artists", err)
		return
	}
	albumList, err := h.Cache.GetAlbumsFromArtist(h.accessToken, artistID)
	if err != nil {
		fmt.Println("couldnt fetch tracks from albums: ", err)
		return
	}
	w.Write(fmt.Appendf(nil, "fetchArtist %#v, it has albums: %#v", fetchArtist, strings.Join(albumList, ", ")))
}

type GetSpotifyCachePlaylist struct {
	accessToken string
	Cache       cache.SpotifyCache
}

func NewGetSpotifyCachePlaylist(accessToken string, c cache.SpotifyCache) *GetSpotifyCachePlaylist {
	return &GetSpotifyCachePlaylist{accessToken, c}
}
func (h *GetSpotifyCachePlaylist) ServeHttp(w http.ResponseWriter, r *http.Request) {
	playlistID := "3cEYpjA9oz9GiPac4AsH4n"
	query := r.URL.Query().Get("id")
	if query != "" {
		playlistID = query
	}
	fetchPlaylist, err := h.Cache.GetPlaylist(h.accessToken, playlistID)
	if err != nil {
		fmt.Println("couldnt fetch playlists", err)
		return
	}
	trackList, err := h.Cache.GetTracksFromPlaylist(h.accessToken, playlistID)
	if err != nil {
		fmt.Println("couldnt fetch tracks from playlist: ", err)
		return
	}
	w.Write(fmt.Appendf(nil, "fetchplaylist %#v, with tracks: %#v", fetchPlaylist, strings.Join(trackList, ", ")))
}

// ///
// type GetSpotifyCacheTrackName struct {
// 	accessToken string
// 	SpotifyApi  *spotify_api.SpotifySongProvider
// 	Cache       cache.SpotifyCache
// }
//
// func NewGetSpotifyCacheTrackName(accessToken string, s *spotify_api.SpotifySongProvider, c cache.SpotifyCache) *GetSpotifyCacheTrackName {
// 	return &GetSpotifyCacheTrackName{accessToken, s, c}
// }
// func (h *GetSpotifyCacheTrackName) ServeHttp(w http.ResponseWriter, r *http.Request) {
// 	trackName := "Salsa"
// 	fetchPlaylistName, err := h.SpotifyApi.SearchPlaylistsByName(h.accessToken, playlistName)
//
// 	trackID := "11dFghVXANMlKmJXsNCbNl"
// 	query := r.URL.Query().Get("id")
// 	if query != "" {
// 		trackID = query
// 	}
//
// 	fetchTrackName, err := h.Cache.GetTrack(h.SpotifyApi, h.accessToken, trackID)
// 	if err != nil {
// 		fmt.Println("couldnt fetch tracks", err)
// 		return
// 	}
// 	fmt.Printf("fetchTrackName%+v\n", fetchTrack)
// 	w.Write(fmt.Appendf(nil, "fetchTrackName %#v", fetchTrack))
// }
