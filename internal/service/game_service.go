package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/FerNunez/NameThatSong/internal/cache"
	"github.com/FerNunez/NameThatSong/internal/game"
	"github.com/FerNunez/NameThatSong/internal/music_player"
	"github.com/FerNunez/NameThatSong/internal/spotify_api"
	"github.com/FerNunez/NameThatSong/internal/store"
	"github.com/FerNunez/NameThatSong/internal/utils"
	"github.com/google/uuid"
)

// GameService coordinates the song provider and music player
type GameService struct {
	MusicPlayer       *player.MusicPlayer
	SpotifyApi        *spotify_api.SpotifySongProvider
	AlbumSelection    map[string]bool
	ArtistSelection   map[string]uint8
	TracksToPlayId    map[string]*player.Song
	GuessState        *game.GuessState
	Cache             *cache.SpotifyCache
	UserId            uuid.UUID
	SpotifyToken      store.SpotifyToken
	SpotifyTokenStore store.SpotifyTokenStore
}

// NewGameService creates a new game service
func NewGameService(clientID, clientSecret, redirectURI string, userId uuid.UUID, spotifyTokenStore store.SpotifyTokenStore) (*GameService, error) {

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
		MusicPlayer:       musicPlayer,
		SpotifyApi:        songProvider,
		AlbumSelection:    make(map[string]bool),
		ArtistSelection:   make(map[string]uint8),
		TracksToPlayId:    make(map[string]*player.Song),
		Cache:             cache.NewSpotifyCache(),
		GuessState:        guessState,
		UserId:            userId,
		SpotifyTokenStore: spotifyTokenStore,
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

func (s GameService) SearchArtists(ctx context.Context, artist string) ([]spotify_api.ArtistData, error) {

	err := s.EnsureAccessToken(ctx)
	if err != nil {
		return []spotify_api.ArtistData{}, fmt.Errorf("No token spotify available")
	}

	artists, err := s.SpotifyApi.SearchArtistsByName(s.SpotifyToken.AccessToken, artist)
	for _, artist := range artists {
		s.Cache.ArtistMap[artist.Id] = artist
	}
	return artists, err
}

func (s GameService) GetArtistsAlbum(ctx context.Context, artistId string) ([]spotify_api.AlbumData, error) {

	err := s.EnsureAccessToken(ctx)
	if err != nil {
		return []spotify_api.AlbumData{}, fmt.Errorf("No token spotify available")
	}

	return s.Cache.GetArtistsAlbum(s.SpotifyApi, s.SpotifyToken.AccessToken, artistId)
}

func (s GameService) GetAlbumTracks(ctx context.Context, albumId string) ([]spotify_api.TrackData, error) {
	err := s.EnsureAccessToken(ctx)
	if err != nil {
		return []spotify_api.TrackData{}, fmt.Errorf("No token spotify available")
	}
	return s.Cache.GetAlbumTracks(s.SpotifyApi, s.SpotifyToken.AccessToken, albumId)
}

// StartGame prepares the game with selected albums
func (s *GameService) StartGame(ctx context.Context) error {
	if len(s.AlbumSelection) <= 0 {
		return errors.New("Empty album selection")
	}

	for artistId := range s.ArtistSelection {
		albumsId, ok := s.Cache.ArtistToAlbumsMap[artistId]
		if !ok {
			panic(fmt.Sprintf("artistId to albumId map should have artist: %v", artistId))
		}

		for _, albumId := range albumsId {
			tracksData, err := s.GetAlbumTracks(ctx, albumId)
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

	err := s.EnsureAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("Couldnt not ensure refresh token")
	}
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
func (s *GameService) SkipSong(ctx context.Context) error {

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

	err = s.EnsureAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("Couldnt not ensure refresh token")
	}
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
func (s *GameService) EnsureAccessToken(ctx context.Context) error {
	//Read from DB
	if s.SpotifyToken.RefreshToken == "" {
		storeSpotifyToken, err := s.SpotifyTokenStore.Get(ctx, s.UserId)
		if err != nil {
			return err
		}
		s.SpotifyToken.AccessToken = storeSpotifyToken.AccessToken
		s.SpotifyToken.RefreshToken = storeSpotifyToken.RefreshToken
		s.SpotifyToken.ExpiresAt = storeSpotifyToken.ExpiresAt
	}

	// Expired?
	if time.Now().After(s.SpotifyToken.ExpiresAt) {
		if s.SpotifyToken.RefreshToken == "" {
			return fmt.Errorf("Refresh token is empty")
		}

		spotifyRefreshReponse, err := s.SpotifyApi.RegenerateToken()
		if err != nil {
			return err
		}

		expires_at := time.Now().Add(time.Duration(spotifyRefreshReponse.ExpiresIn) * time.Second)
		s.SpotifyToken = store.SpotifyToken{
			RefreshToken: spotifyRefreshReponse.RefreshToken,
			AccessToken:  spotifyRefreshReponse.AccessToken,
			TokenType:    spotifyRefreshReponse.TokenType,
			Scope:        spotifyRefreshReponse.Scope,
			ExpiresAt:    expires_at,
		}

		// update token
		err = s.SpotifyTokenStore.Update(ctx, s.UserId, s.SpotifyToken.AccessToken, s.SpotifyToken.ExpiresAt)
		if err != nil {
			return err
		}
	}
	return nil
}
