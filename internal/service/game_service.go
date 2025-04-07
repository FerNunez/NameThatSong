package service

import (
	"errors"
	"fmt"
	"goth/internal/cache"
	"goth/internal/music_player"
	"goth/internal/spotify_api"
)

// GameService coordinates the song provider and music player
type GameService struct {
	MusicPlayer     *player.MusicPlayer
	SpotifyApi      *spotify_api.SpotifySongProvider
	AlbumSelection  map[string]bool
	ArtistSelection map[string]uint8
	TrackPoolId     []string
	Cache           *cache.SpotifyCache
}

// NewGameService creates a new game service
func NewGameService(player *player.MusicPlayer, provider *spotify_api.SpotifySongProvider) *GameService {
	return &GameService{
		MusicPlayer:     player,
		SpotifyApi:      provider,
		AlbumSelection:  make(map[string]bool),
		ArtistSelection: make(map[string]uint8),
		Cache:           cache.NewSpotifyCache(),
	}
}

// SelectAlbum selects or deselects an album
func (s *GameService) ToggleAlbumSelection(albumID string, artistId string) bool {

	isSelected := false
	if _, exists := s.AlbumSelection[albumID]; exists {
		delete(s.AlbumSelection, albumID)

		if occurrances, exists := s.ArtistSelection[artistId]; exists {
			s.ArtistSelection[artistId] -= 1
			if occurrances == 1 {
				delete(s.ArtistSelection, artistId)
			}
			return false
		}
	} else {
		s.AlbumSelection[albumID] = true
		isSelected = true
		s.ArtistSelection[artistId] += 1
	}

	s.AlbumSelection[albumID] = true
	return isSelected
}

// GetSelectedAlbums returns the currently selected albums
func (s *GameService) GetSelectedAlbums() []string {
	albums := make([]string, 0, len(s.AlbumSelection))
	for id := range s.AlbumSelection {
		albums = append(albums, id)
	}
	return albums
}

func (s GameService) SearchArtists(artist string) ([]spotify_api.ArtistData, error) {
	artists, err := s.SpotifyApi.SearchArtistsByName(artist)
	return artists, err
}

func (s GameService) GetArtistsAlbum(artistId string) ([]spotify_api.AlbumData, error) {
	// check if artist already known
	albumsIds, exist := s.Cache.ArtistToAlbumsMap[artistId]
	if !exist {
		albums, err := s.SpotifyApi.FetchAlbumByArtistID(artistId)
		if err != nil {
			return nil, err
		}

		// update Artist to albumMaps
		albumsIds = make([]string, 0, len(albums))
		for _, album := range albums {
			s.Cache.AlbumMap[album.ID] = album
			albumsIds = append(albumsIds, album.ID)
		}
		s.Cache.ArtistToAlbumsMap[artistId] = albumsIds
	}

	albums := make([]spotify_api.AlbumData, 0, len(albumsIds))
	for _, albumId := range albumsIds {
		album, _ := s.Cache.AlbumMap[albumId]
		albums = append(albums, album)
	}

	return albums, nil
}

func (s GameService) GetAlbumTracks(albumId string) ([]spotify_api.TrackData, error) {
	tracksIds, exist := s.Cache.AlbumToTracksMap[albumId]
	if !exist {
		tracks, err := s.SpotifyApi.FetchTracksByAlbumID(albumId)
		if err != nil {
			return nil, err
		}
		tracksIds = make([]string, 0, len(tracks))
		for _, track := range tracks {
			s.Cache.TrackMap[track.ID] = track
			tracksIds = append(tracksIds, track.ID)
		}
		s.Cache.AlbumToTracksMap[albumId] = tracksIds
	}

	tracks := make([]spotify_api.TrackData, 0, len(tracksIds))
	for _, trackId := range tracksIds {
		track, _ := s.Cache.TrackMap[trackId]
		tracks = append(tracks, track)
	}
	return tracks, nil
}

// StartGame prepares the game with selected albums
func (s *GameService) StartGame() error {
	fmt.Println("Game Started!")
	if len(s.AlbumSelection) <= 0 {
		return errors.New("Empty album selection")
	}

	fmt.Println(len(s.AlbumSelection))
	//trackId := s.MusicPlayer.Queue[s.MusicPlayer.CurrentIndex]
	//s.SpotifyApi.PlaySong(trackId)
	return nil
}

// PlayGame starts playback
//func (s *GameService) PlayGame() error {}

// MakeGuess checks if the user's guess matches the current song
//func (s *GameService) MakeGuess(guess string) (bool, string, error) { }

// GetScore returns the current score
//func (s *GameService) GetScore() int { }

// SkipSong skips to the next song
//func (s *GameService) SkipSong() error {}

// ClearQueue clears the current music queue
//func (s *GameService) ClearQueue() error { }
