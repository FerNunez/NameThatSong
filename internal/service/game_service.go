package service

import (
	"errors"
	"fmt"
	"goth/internal/cache"
	"goth/internal/game"
	"goth/internal/music_player"
	"goth/internal/spotify_api"
	"strings"
)

// GameService coordinates the song provider and music player
type GameService struct {
	MusicPlayer     *player.MusicPlayer
	SpotifyApi      *spotify_api.SpotifySongProvider
	AlbumSelection  map[string]bool
	ArtistSelection map[string]uint8
	TracksToPlayId  []string
	TracksOptions   []spotify_api.TrackData
	GuessState      game.GuessState
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

			// s.Cache.AlbumIdToArtistId[album.ID] = artistId
		}
		s.Cache.ArtistToAlbumsMap[artistId] = albumsIds

		// todo: add this album to artist
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

			//s.Cache.TrackIdToAlbumId[track.ID] = albumId
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
	if len(s.AlbumSelection) <= 0 {
		return errors.New("Empty album selection")
	}

	for artistId := range s.ArtistSelection {
		albumsId, ok := s.Cache.ArtistToAlbumsMap[artistId]
		if !ok {
			panic(fmt.Sprintf("artistId to albumId map should have artist: %v", artistId))
		}

		for _, albumId := range albumsId {
			tracksData, err := s.GetAlbumTracks(albumId)
			if err != nil {
				return err
			}
			s.TracksOptions = append(s.TracksOptions, tracksData...)

			_, exists := s.AlbumSelection[albumId]
			if exists {
				for _, track := range tracksData {
					s.TracksToPlayId = append(s.TracksToPlayId, track.ID)
				}
			}
		}

	}

	s.MusicPlayer.Queue = append(s.MusicPlayer.Queue, s.TracksToPlayId...)
	s.MusicPlayer.Shuffle()
	trackId := s.MusicPlayer.Queue[s.MusicPlayer.CurrentIndex]

	track, ok := s.Cache.TrackMap[trackId]
	if !ok {
		panic("Track should always exist in cache")
	}

	// guessSong process:
	s.GuessState = *game.NewGameState(track.Name)
	s.SpotifyApi.PlaySong(trackId)

	// Debug
	println("track Name:", track.Name)
	s.GuessState.ShowState()

	return nil
}

func (s *GameService) FilterTrackOptions(trackGuess string) ([]spotify_api.TrackData, error) {
	if len(s.TracksOptions) < 0 {
		return nil, errors.New("empty tracks options")
	}

	guessOptions := []spotify_api.TrackData{}
	for _, artistData := range s.TracksOptions {
		addTrack := true
		guessWord := strings.Split(strings.ToLower(trackGuess), " ")
		artistNameLowered := strings.ToLower(artistData.Name)

		for _, word := range guessWord {
			if !strings.Contains(artistNameLowered, word) {
				addTrack = false
				break
			}
		}

		if addTrack {
			guessOptions = append(guessOptions, artistData)
		}

	}

	return guessOptions, nil
}

// User tries to guess
func (s *GameService) UserGuess(guess string) (string, error) {

	guessResult, guessedCorrectly := s.GuessState.Guess(guess)

	if guessedCorrectly {

		nextTrackId, err := s.MusicPlayer.NextInQueue()
		if err != nil {
			return "", err
		}

		track, ok := s.Cache.TrackMap[nextTrackId]
		if !ok {
			panic("Track should always exist in cache")
		}

		// guessSong process:
		fmt.Println("new song to guess title", track.Name)
		s.GuessState = *game.NewGameState(track.Name)
		err = s.SpotifyApi.PlaySong(nextTrackId)
		if err != nil {
			return "", err
		}
	}

	return guessResult, nil
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
