package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/FerNunez/NameThatSong/internal/services/cache"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
	"github.com/FerNunez/NameThatSong/web/templates"
)

type GetSearchMusic struct {
	accessToken string
	Cache       cache.SpotifyCache
}

func NewGetSearchMusic(accessToken string, c cache.SpotifyCache) *GetSearchMusic {
	return &GetSearchMusic{accessToken, c}
}
func (h *GetSearchMusic) ServeHttp(w http.ResponseWriter, r *http.Request) {
	// TODO: add as request/depending on front end?
	numElems := 5
	startIdx := 1

	query := r.URL.Query().Get("search")
	if query == "" || len(query) < 2 {
		// TODO: Should I change to Result?
		component := templates.MusicSearch([]spotify.TrackSearch{}, []spotify.AlbumSearch{}, []spotify.ArtistSearch{}, []spotify.PlaylistSearch{})
		component.Render(r.Context(), w)
	}

	tracksChan := make(chan []spotify.TrackSearch)
	albumsChan := make(chan []spotify.AlbumSearch)
	artistsChan := make(chan []spotify.ArtistSearch)
	playlistsChan := make(chan []spotify.PlaylistSearch)
	errorsChan := make(chan error, 4)

	go func() {
		now := time.Now()
		fmt.Println("execute duration start: ", now)
		trackList, err := h.Cache.SearchTracks(h.accessToken, query)

		fmt.Println("over: ", time.Since(now))
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
		trackList    []spotify.TrackSearch
		albumList    []spotify.AlbumSearch
		artistList   []spotify.ArtistSearch
		playlistList []spotify.PlaylistSearch
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

	slicedTrackList := trackList[min(startIdx, len(trackList)):min(len(trackList), numElems+startIdx)]
	slicedAlbumList := albumList[0:min(len(albumList), numElems)]
	slicedArtistList := artistList[0:min(len(artistList), numElems)]
	slicedPlaylistList := playlistList[0:min(len(playlistList), numElems)]

	sort.Slice(slicedTrackList, func(i, j int) bool {
		return slicedTrackList[i].Popularity > slicedTrackList[j].Popularity
	})

	component := templates.MusicSearch(slicedTrackList, slicedAlbumList, slicedArtistList, slicedPlaylistList)
	component.Render(r.Context(), w)

}

type GetStackMusic struct {
	accessToken string
	Cache       cache.SpotifyCache
}

func NewGetStackMusic(accessToken string, c cache.SpotifyCache) *GetStackMusic {
	return &GetStackMusic{accessToken, c}
}
func (h *GetStackMusic) ServeHttp(w http.ResponseWriter, r *http.Request) {

	query := "Example"
	trackList, err := h.Cache.SearchTracks(h.accessToken, query)
	if err != nil {
		return
	}

	component := templates.StackMusic(trackList)
	component.Render(r.Context(), w)
}
