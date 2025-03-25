package service

import (
	"errors"
	"goth/internal/player"
	"math/rand"
	"strings"
	"time"
)

// GameService coordinates the song provider and music player
type GameService struct {
	Player         player.MusicPlayer
	SongProvider   player.SongProvider
	SelectedAlbums map[string]bool
	CurrentGuess   string
	Score          int
}

// NewGameService creates a new game service
func NewGameService(player player.MusicPlayer, provider player.SongProvider) *GameService {
	return &GameService{
		Player:         player,
		SongProvider:   provider,
		SelectedAlbums: make(map[string]bool),
		CurrentGuess:   "",
		Score:          0,
	}
}

// SelectAlbum selects or deselects an album
func (s *GameService) SelectAlbum(albumID string) error {
	if _, exists := s.SelectedAlbums[albumID]; exists {
		delete(s.SelectedAlbums, albumID)
	} else {
		s.SelectedAlbums[albumID] = true
	}
	return nil
}

// GetSelectedAlbums returns the currently selected albums
func (s *GameService) GetSelectedAlbums() []string {
	albums := make([]string, 0, len(s.SelectedAlbums))
	for id := range s.SelectedAlbums {
		albums = append(albums, id)
	}
	return albums
}

// StartGame prepares the game with selected albums
func (s *GameService) StartGame() error {
	if len(s.SelectedAlbums) == 0 {
		return errors.New("no albums selected")
	}

	// Get all album IDs
	albumIDs := make([]string, 0, len(s.SelectedAlbums))
	for id := range s.SelectedAlbums {
		albumIDs = append(albumIDs, id)
	}

	// Get all songs for selected albums
	songsByAlbum, err := s.SongProvider.GetSongsByAlbumIDs(albumIDs)
	if err != nil {
		return err
	}

	// Flatten the songs
	allSongs := make([]player.Song, 0)
	for _, songs := range songsByAlbum {
		allSongs = append(allSongs, songs...)
	}

	// Shuffle the songs
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(allSongs), func(i, j int) {
		allSongs[i], allSongs[j] = allSongs[j], allSongs[i]
	})

	// Cap at 10 songs if we have more
	if len(allSongs) > 10 {
		allSongs = allSongs[:10]
	}

	// Add to queue
	s.Player.ClearQueue()
	s.Player.AddToQueue(allSongs)

	return nil
}

// PlayGame starts playback
func (s *GameService) PlayGame() error {
	return s.Player.Play()
}

// MakeGuess checks if the user's guess matches the current song
func (s *GameService) MakeGuess(guess string) (bool, string, error) {
	currentSong := s.Player.GetCurrentSong()
	if currentSong == nil {
		return false, "", errors.New("no song is currently playing")
	}

	// Simple string comparison for now
	// Could be improved with fuzzy matching, ignoring case, etc.
	cleanGuess := strings.ToLower(strings.TrimSpace(guess))
	cleanSongName := strings.ToLower(strings.TrimSpace(currentSong.Name))

	correct := cleanGuess == cleanSongName

	if correct {
		s.Score++
	}

	return correct, currentSong.Name, nil
}

// GetScore returns the current score
func (s *GameService) GetScore() int {
	return s.Score
}

// SkipSong skips to the next song
func (s *GameService) SkipSong() error {
	return s.Player.Skip()
}
