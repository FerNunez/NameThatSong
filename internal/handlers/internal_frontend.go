package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/FerNunez/NameThatSong/internal/cache"
	"github.com/FerNunez/NameThatSong/internal/spotify_api"
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
	query := r.URL.Query().Get("search")
	if query == "" {
		// TODO: Should I change to Result?
		component := templates.MusicSearch([]spotify_api.TrackSearch{}, []spotify_api.AlbumSearch{}, []spotify_api.ArtistSearch{}, []spotify_api.PlaylistSearch{})
		component.Render(r.Context(), w)
	}

	tracksChan := make(chan []spotify_api.TrackSearch)
	albumsChan := make(chan []spotify_api.AlbumSearch)
	artistsChan := make(chan []spotify_api.ArtistSearch)
	playlistsChan := make(chan []spotify_api.PlaylistSearch)
	errorsChan := make(chan error, 4)

	go func() {
		trackList, err := h.Cache.SearchTracks(h.accessToken, query)
		if err != nil {
			errorsChan <- fmt.Errorf("track search error: %v", err)
			tracksChan <- nil
		}
		tracksChan <- trackList
	}()

	go func() {
		albumList, err := h.Cache.SearchAlbums(h.accessToken, query)
		if err != nil {
			errorsChan <- fmt.Errorf("album search error: %v", err)
			albumsChan <- nil
		}
		albumsChan <- albumList
	}()
	go func() {
		artistList, err := h.Cache.SearchArtists(h.accessToken, query)
		if err != nil {
			errorsChan <- fmt.Errorf("artist search error: %v", err)
			artistsChan <- nil
		}
		artistsChan <- artistList
	}()

	go func() {
		playlistList, err := h.Cache.SearchPlaylists(h.accessToken, query)
		if err != nil {
			errorsChan <- fmt.Errorf("playlist search error: %v", err)
			playlistsChan <- nil
		}
		playlistsChan <- playlistList
	}()

	var (
		trackList    []spotify_api.TrackSearch
		albumList    []spotify_api.AlbumSearch
		artistList   []spotify_api.ArtistSearch
		playlistList []spotify_api.PlaylistSearch
	)

	timeout := time.After(5 * time.Second)

	// TODO: To change to a sync.WaitGroup ?
	//This for is to wait for all of them
	for range 4 {
		select {
		case trackList = <-tracksChan:
		case albumList = <-albumsChan:
		case artistList = <-artistsChan:
		case playlistList = <-playlistsChan:
		case err := <-errorsChan:
			fmt.Printf("Search error: %v\n", err)
		case <-timeout:
			http.Error(w, "Search timeout", http.StatusGatewayTimeout)
			return
		}

	}
	fmt.Printf("got %v tracks, %v album, %v artist, %v playlist\n", len(trackList), len(albumList), len(artistList), len(playlistList))
	component := templates.MusicSearch(trackList, albumList, artistList, playlistList)
	component.Render(r.Context(), w)
}
