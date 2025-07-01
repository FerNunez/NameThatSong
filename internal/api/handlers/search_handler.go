package handlers

import (
	"fmt"
	"net/http"
	"sort"

	m "github.com/FerNunez/NameThatSong/internal/models"

	"github.com/FerNunez/NameThatSong/internal/services/game"
	"github.com/FerNunez/NameThatSong/web/templates"
)

type GetSearchArtists struct {
	gm *game.GameManager
}

func NewGetSearchArtists(gm *game.GameManager) *GetSearchArtists {
	return &GetSearchArtists{gm}
}

func (h *GetSearchArtists) ServeHttp(w http.ResponseWriter, r *http.Request) {
	game, err := h.gm.GetGame(r.Context())
	if err != nil {
		fmt.Printf("error getting game : %v\n", err)
		return
	}
	// Get query term
	query := r.URL.Query().Get("search")
	if query == "" {
		// Return empty results component
		component := templates.SearchResults([]m.ArtistData{})
		component.Render(r.Context(), w)
		return
	}

	artists, err := game.SearchArtists(r.Context(), query)
	if err != nil || len(artists) == 0 {
		// TODO: SEND ERROR TO REQUEST
		component := templates.SearchResults([]m.ArtistData{})
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
	gm *game.GameManager
}

func NewGetArtistAlbums(gm *game.GameManager) *GetArtistAlbums {
	return &GetArtistAlbums{gm}
}

func (h *GetArtistAlbums) ServeHttp(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Get artist")
	game, err := h.gm.GetGame(r.Context())
	if err != nil {
		fmt.Printf("error getting game : %v", err)
		return
	}
	artistID := r.URL.Query().Get("artist-id")
	fmt.Println("got artist ID", artistID)
	if artistID == "" {
		http.Error(w, "Artist ID is required", http.StatusBadRequest)
		return
	}

	albums, err := game.GetArtistsAlbum(r.Context(), artistID)
	//albums, err := game.SpotifyApi.FetchAlbumByArtistID(artistID)
	if err != nil {
		http.Error(w, "Cant retrieve Artist ID albums", http.StatusBadRequest)
		fmt.Println("no album")
		return
	}

	component := templates.AlbumDropdown(albums, game.AlbumSelection, artistID)
	component.Render(r.Context(), w)
}
