package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/FerNunez/NameThatSong/internal/services/cache"
	"github.com/FerNunez/NameThatSong/web/templates"
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

// /
type GetSpotifyCacheTrackName struct {
	accessToken string
	Cache       cache.SpotifyCache
}

func NewGetSpotifyCacheTrackName(accessToken string, c cache.SpotifyCache) *GetSpotifyCacheTrackName {
	return &GetSpotifyCacheTrackName{accessToken, c}
}
func (h *GetSpotifyCacheTrackName) ServeHttp(w http.ResponseWriter, r *http.Request) {
	trackName := "Re forro"
	query := r.URL.Query().Get("name")
	if query != "" {
		trackName = query
	}
	trackList, err := h.Cache.SearchTracks(h.accessToken, trackName)
	if err != nil {
		fmt.Println("couldnt fetch tracks", err)
		return
	}
	fmt.Printf("search tracks cache for:  %+v\n", trackName)
	w.Write(fmt.Appendf(nil, "search track %#v", trackList))
}

type GetSpotifyCacheAlbumName struct {
	accessToken string
	Cache       cache.SpotifyCache
}

func NewGetSpotifyCacheAlbumName(accessToken string, c cache.SpotifyCache) *GetSpotifyCacheAlbumName {
	return &GetSpotifyCacheAlbumName{accessToken, c}
}
func (h *GetSpotifyCacheAlbumName) ServeHttp(w http.ResponseWriter, r *http.Request) {
	albumName := "Papota"
	query := r.URL.Query().Get("name")
	if query != "" {
		albumName = query
	}
	albumList, err := h.Cache.SearchAlbums(h.accessToken, albumName)
	if err != nil {
		fmt.Println("couldnt fetch albums", err)
		return
	}
	fmt.Printf("search albums cache for:  %+v\n", albumName)
	w.Write(fmt.Appendf(nil, "search album %#v", albumList))
}

type GetSpotifyCacheArtistName struct {
	accessToken string
	Cache       cache.SpotifyCache
}

func NewGetSpotifyCacheArtistName(accessToken string, c cache.SpotifyCache) *GetSpotifyCacheArtistName {
	return &GetSpotifyCacheArtistName{accessToken, c}
}
func (h *GetSpotifyCacheArtistName) ServeHttp(w http.ResponseWriter, r *http.Request) {
	artistName := "Lady Gaga"
	query := r.URL.Query().Get("name")
	if query != "" {
		artistName = query
	}
	artistList, err := h.Cache.SearchArtists(h.accessToken, artistName)
	if err != nil {
		fmt.Println("couldnt fetch artists", err)
		return
	}
	fmt.Printf("search artists cache for:  %+v\n", artistName)
	w.Write(fmt.Appendf(nil, "search artist %#v", artistList))
}

type GetSpotifyCachePlaylistName struct {
	accessToken string
	Cache       cache.SpotifyCache
}

func NewGetSpotifyCachePlaylistName(accessToken string, c cache.SpotifyCache) *GetSpotifyCachePlaylistName {
	return &GetSpotifyCachePlaylistName{accessToken, c}
}
func (h *GetSpotifyCachePlaylistName) ServeHttp(w http.ResponseWriter, r *http.Request) {
	playlistName := "Salsa"
	query := r.URL.Query().Get("name")
	if query != "" {
		playlistName = query
	}
	playlistList, err := h.Cache.SearchPlaylists(h.accessToken, playlistName)
	if err != nil {
		fmt.Println("couldnt fetch playlists", err)
		return
	}
	fmt.Printf("search playlists cache for:  %+v\n", playlistName)
	w.Write(fmt.Appendf(nil, "search playlist %#v", playlistList))
}
