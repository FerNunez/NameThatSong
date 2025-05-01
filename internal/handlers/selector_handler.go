package handlers

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/FerNunez/NameThatSong/internal/manager"
	"github.com/FerNunez/NameThatSong/internal/spotify_api"
	"github.com/FerNunez/NameThatSong/internal/templates"
)

type GetSearchArtists struct {
	gm *manager.GameManager
}

func NewGetSearchArtists(gm *manager.GameManager) *GetSearchArtists {
	return &GetSearchArtists{gm}
}

func (h *GetSearchArtists) ServeHttp(w http.ResponseWriter, r *http.Request) {
	game, err := h.gm.GetGame(r.Context())
	if err != nil {
		fmt.Printf("error getting game : %v", err)
		return
	}
	// Get query term
	query := r.URL.Query().Get("search")
	if query == "" {
		// Return empty results component
		component := templates.SearchResults([]spotify_api.ArtistData{})
		component.Render(r.Context(), w)
		return
	}

	artists, err := game.SearchArtists(query)
	if err != nil || len(artists) == 0 {
		// TODO: SEND ERROR TO REQUEST
		component := templates.SearchResults([]spotify_api.ArtistData{})
		component.Render(r.Context(), w)
		return
	}

	// Sort by popularity
	sort.Slice(artists, func(i, j int) bool {
		return artists[i].Popularity > artists[j].Popularity
	})

	// Return results component
	component := templates.SearchResults(artists)
	component.Render(r.Context(), w)
}

////////////////////////////

type GetArtistAlbums struct {
	gm *manager.GameManager
}

func NewGetArtistAlbums(gm *manager.GameManager) *GetArtistAlbums {
	return &GetArtistAlbums{gm}
}

func (h *GetArtistAlbums) ServeHttp(w http.ResponseWriter, r *http.Request) {
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
