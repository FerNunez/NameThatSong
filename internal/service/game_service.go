package service

import (
	"errors"
	"fmt"
	"goth/internal/cache"
	"goth/internal/game"
	player "goth/internal/music_player"
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
	GuessState      *game.GuessState
	Cache           *cache.SpotifyCache
}

// NewGameService creates a new game service
func NewGameService(player *player.MusicPlayer, provider *spotify_api.SpotifySongProvider) *GameService {
	guessState := game.NewGameState()
	return &GameService{
		MusicPlayer:     player,
		SpotifyApi:      provider,
		AlbumSelection:  make(map[string]bool),
		ArtistSelection: make(map[string]uint8),
		Cache:           cache.NewSpotifyCache(),
		GuessState:      guessState,
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
	return s.Cache.GetArtistsAlbum(s.SpotifyApi, artistId)
}

func (s GameService) GetAlbumTracks(albumId string) ([]spotify_api.TrackData, error) {
	return s.Cache.GetAlbumTracks(s.SpotifyApi, albumId)
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
	s.GuessState.SetTitle(track.Name)
	s.SpotifyApi.PlaySong(trackId)

	// Debug
	println("track Name:", track.Name)

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
		s.GuessState.SetTitle(track.Name)
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
func (s *GameService) SkipSong() error {

	nextTrackId, err := s.MusicPlayer.NextInQueue()
	if err != nil {
		return err
	}
	track, ok := s.Cache.TrackMap[nextTrackId]
	if !ok {
		panic("Track should always exist in cache")
	}

	// guessSong process:
	s.GuessState.SetTitle(track.Name)
	return s.SpotifyApi.PlaySong(nextTrackId)
}

// ClearQueue clears the current music queue
func (s *GameService) ClearQueue() error {
	s.AlbumSelection = make(map[string]bool)
	s.ArtistSelection = make(map[string]uint8)
	s.GuessState = game.NewGameState()
	s.MusicPlayer.ClearQueue()
	s.SpotifyApi.PausePlayback()
	return nil

}
