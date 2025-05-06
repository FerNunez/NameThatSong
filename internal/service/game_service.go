package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/FerNunez/NameThatSong/internal/cache"
	"github.com/FerNunez/NameThatSong/internal/game"
	"github.com/FerNunez/NameThatSong/internal/music_player"
	"github.com/FerNunez/NameThatSong/internal/spotify_api"
	"github.com/FerNunez/NameThatSong/internal/store"
	"github.com/FerNunez/NameThatSong/internal/utils"
)

// GameService coordinates the song provider and music player
type GameService struct {
	MusicPlayer     *player.MusicPlayer
	SpotifyApi      *spotify_api.SpotifySongProvider
	AlbumSelection  map[string]bool
	ArtistSelection map[string]uint8
	TracksToPlayId  map[string]*player.Song
	GuessState      *game.GuessState
	Cache           *cache.SpotifyCache
	UserId          string
	SpotifyToken    store.SpotifyToken
}

// NewGameService creates a new game service
func NewGameService(clientID, clientSecret, redirectURI, userId string) (*GameService, error) {

	// Generate a random state for OAuth
	state, err := utils.GenerateState(16)
	if err != nil {
		return nil, fmt.Errorf("error generating state: %v", err)
	}

	// Create song provider
	songProvider := spotify_api.NewSpotifySongProvider(clientID, clientSecret, redirectURI, state)
	// Create music service client
	musicPlayer := player.NewMusicPlayer()
	// Create game service
	guessState := game.NewGameState()
	return &GameService{
		MusicPlayer:     musicPlayer,
		SpotifyApi:      songProvider,
		AlbumSelection:  make(map[string]bool),
		ArtistSelection: make(map[string]uint8),
		TracksToPlayId:  make(map[string]*player.Song),
		Cache:           cache.NewSpotifyCache(),
		GuessState:      guessState,
		UserId:          userId,
	}, nil
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
	artists, err := s.SpotifyApi.SearchArtistsByName(s.SpotifyToken.AccessToken, artist)

	for _, artist := range artists {
		s.Cache.ArtistMap[artist.Id] = artist
	}
	return artists, err
}

func (s GameService) GetArtistsAlbum(artistId string) ([]spotify_api.AlbumData, error) {
	return s.Cache.GetArtistsAlbum(s.SpotifyApi, s.SpotifyToken.AccessToken, artistId)
}

func (s GameService) GetAlbumTracks(albumId string) ([]spotify_api.TrackData, error) {
	return s.Cache.GetAlbumTracks(s.SpotifyApi, s.SpotifyToken.AccessToken, albumId)
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

			_, exists := s.AlbumSelection[albumId]
			if exists {
				for _, track := range tracksData {
					s.TracksToPlayId[track.ID] = player.NewSong(track.ID, albumId, artistId)
				}
			}
		}

	}

	for _, song := range s.TracksToPlayId {
		s.MusicPlayer.Queue = append(s.MusicPlayer.Queue, *song)
	}
	s.MusicPlayer.Shuffle()
	song := s.MusicPlayer.Queue[s.MusicPlayer.CurrentIndex]

	track, ok := s.Cache.TrackMap[song.TrackId]
	if !ok {
		panic("Track should always exist in cache")
	}
	album, ok := s.Cache.AlbumMap[song.AlbumId]
	if !ok {
		panic("Track should always exist in cache")
	}
	artist, ok := s.Cache.ArtistMap[song.ArtistId]
	if !ok {
		panic("Track should always exist in cache")
	}

	// guessSong process:
	s.GuessState.SetTitle(track.Name, artist.Name, album.ImagesURL)
	s.MusicPlayer.Timer = time.Now()
	s.MusicPlayer.SongDuration = time.Duration(track.DurationMs) * time.Millisecond
	fmt.Println("song duration:", track.DurationMs)
	s.SpotifyApi.PlaySong(s.SpotifyToken.AccessToken, song.TrackId)

	// Debug
	println("track Name:", track.Name)

	return nil
}

// User tries to guess
func (s *GameService) UserGuess(guess string) (bool, error) {

	_, guessedCorrectly := s.GuessState.Guess(guess)

	return guessedCorrectly, nil
}

// SkipSong skips to the next song
func (s *GameService) SkipSong() error {

	nextSong, err := s.MusicPlayer.NextInQueue()
	if err != nil {
		return err
	}
	track, ok := s.Cache.TrackMap[nextSong.TrackId]
	if !ok {
		panic("Track should always exist in cache")
	}

	album, ok := s.Cache.AlbumMap[nextSong.AlbumId]
	if !ok {
		panic("Track should always exist in cache")
	}
	artist, ok := s.Cache.ArtistMap[nextSong.ArtistId]
	if !ok {
		panic("Track should always exist in cache")
	}

	// guessSong process:
	s.GuessState.SetTitle(track.Name, artist.Name, album.ImagesURL)
	s.MusicPlayer.Timer = time.Now()
	s.MusicPlayer.SongDuration = time.Duration(track.DurationMs) * time.Millisecond
	return s.SpotifyApi.PlaySong(s.SpotifyToken.AccessToken, nextSong.TrackId)
}

// ClearQueue clears the current music queue
func (s *GameService) ClearQueue() error {
	s.AlbumSelection = make(map[string]bool)
	s.ArtistSelection = make(map[string]uint8)
	s.GuessState = game.NewGameState()
	s.MusicPlayer.ClearQueue()
	s.SpotifyApi.PausePlayback(s.SpotifyToken.AccessToken)
	return nil
}

func (s GameService) spotifyTokenIsValid() bool {
	if time.Now().After(s.SpotifyToken.ExpiresAt) {
		return false
	}
	return true
}

func (s *GameService) RequestUserAuthoritazion() (string, error) {
	urlString, err := s.SpotifyApi.AuthRequestURL()
	return urlString, err
}

func (s *GameService) ExchangeToken(state, code string) error {
	if code == "" || state == "" {
		return fmt.Errorf("Error guetting code and state from spotify api")
	}
	err := s.SpotifyApi.ValidateState(state)
	if err != nil {
		return err
	}
	spotiufyTokenReponse, err := s.SpotifyApi.TokenExchange(code)
	if err != nil {
		return err
	}

	expires_at := time.Now().Add(time.Duration(spotiufyTokenReponse.ExpiresIn) * time.Second)
	s.SpotifyToken = store.SpotifyToken{
		RefreshToken: spotiufyTokenReponse.RefreshToken,
		AccessToken:  spotiufyTokenReponse.AccessToken,
		TokenType:    spotiufyTokenReponse.TokenType,
		Scope:        spotiufyTokenReponse.Scope,
		ExpiresAt:    expires_at,
	}
	return nil
}
