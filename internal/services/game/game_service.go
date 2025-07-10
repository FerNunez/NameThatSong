package game

import (
	"context"
	"errors"
	"fmt"
	"time"

	m "github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/repository"
	"github.com/FerNunez/NameThatSong/internal/services/cache"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
	"github.com/FerNunez/NameThatSong/internal/utils"
	"github.com/google/uuid"
)

// GameService coordinates the song provider and music player
type GameService struct {
	MusicPlayer       *MusicPlayer
	SpotifyApi        *spotify.SpotifyService
	AlbumSelection    map[string]bool
	ArtistSelection   map[string]uint8
	TracksToPlayId    map[string]*Song
	GuessState        *GuessState
	Cache             cache.SpotifyCache
	UserId            uuid.UUID
	SpotifyToken      repository.SpotifyToken
	SpotifyTokenStore repository.SpotifyTokenStore
}

// NewGameService creates a new game service
func NewGameService(clientID, clientSecret, redirectURI string, userId uuid.UUID, spotifyTokenStore repository.SpotifyTokenStore) (*GameService, error) {

	// Generate a random state for OAuth
	_, err := utils.GenerateState(16)
	if err != nil {
		return nil, fmt.Errorf("error generating state: %v", err)
	}
	// Create song provider
	// TODO: implement this
	songProvider, err := spotify.NewSpotifyService(nil, nil, nil)
	// Create music service client
	musicPlayer := NewMusicPlayer()
	// Create game service
	guessState := NewGameState()
	return &GameService{
		MusicPlayer:       musicPlayer,
		SpotifyApi:        songProvider,
		AlbumSelection:    make(map[string]bool),
		ArtistSelection:   make(map[string]uint8),
		TracksToPlayId:    make(map[string]*Song),
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

func (s GameService) SearchArtists(ctx context.Context, artist string) ([]m.ArtistData, error) {

	err := s.EnsureAccessToken(ctx)
	if err != nil {
		return []m.ArtistData{}, fmt.Errorf("No token spotify available")
	}

	// artists, err := s.Cache.GetArtistsByName(s.SpotifyApi, artist)
	// // artists, err := s.SpotifyApi.SearchArtistsByName(s.SpotifyToken.AccessToken, artist)
	// // for _, artist := range artists {
	// // 	s.Cache.ArtistMap[artist.Id] = artist
	// // }
	// return artists, err

	return []m.ArtistData{}, fmt.Errorf("No token spotify available")
}

func (s GameService) GetArtistsAlbum(ctx context.Context, artistId string) ([]m.AlbumData, error) {

	err := s.EnsureAccessToken(ctx)
	if err != nil {
		return []m.AlbumData{}, fmt.Errorf("No token spotify available")
	}

	// return s.Cache.GetArtistsAlbum(s.SpotifyApi, s.SpotifyToken.AccessToken, artistId)
	return []m.AlbumData{}, fmt.Errorf("No token spotify available")
}

func (s GameService) GetAlbumTracks(ctx context.Context, albumId string) ([]m.TrackData, error) {
	err := s.EnsureAccessToken(ctx)
	if err != nil {
		return []m.TrackData{}, fmt.Errorf("No token spotify available")
	}
	// return s.Cache.GetAlbumTracks(s.SpotifyApi, s.SpotifyToken.AccessToken, albumId)
	return []m.TrackData{}, fmt.Errorf("No token spotify available")
}

// StartGame prepares the game with selected albums
func (s *GameService) StartGame(ctx context.Context) error {
	if len(s.AlbumSelection) <= 0 {
		return errors.New("Empty album selection")
	}

	for artistId := range s.ArtistSelection {
		// albumsId, err := s.Cache.GetAlbumsIdFromArtistId(s.SpotifyApi, artistId)
		albumsId := []string{}
		err := errors.New("asd")
		//albumsId, ok := s.Cache.ArtistToAlbumsMap[artistId]
		if err != nil {
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
					s.TracksToPlayId[track.ID] = NewSong(track.ID, albumId, artistId)
				}
			}
		}

	}

	for _, song := range s.TracksToPlayId {
		s.MusicPlayer.Queue = append(s.MusicPlayer.Queue, *song)
	}
	s.MusicPlayer.Shuffle()
	//song := s.MusicPlayer.Queue[s.MusicPlayer.CurrentIndex]

	// TODO: FIX THIS
	track := m.TrackData{}
	artist := m.ArtistData{}
	album := m.AlbumData{}

	// guessSong process:
	s.GuessState.SetTitle(track.Name, artist.Name, album.ImagesURL)
	s.MusicPlayer.Timer = time.Now()
	s.MusicPlayer.SongDuration = time.Duration(track.DurationMs) * time.Millisecond

	err := s.EnsureAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("Couldnt not ensure refresh token")
	}
	//s.SpotifyApi.PlaySong(s.SpotifyToken.AccessToken, song.TrackId)

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

	// _, err := s.MusicPlayer.NextInQueue()
	// if err != nil {
	// 	return err
	// }
	// track, ok := s.Cache.GetTrack(s.SpotifyToken.AccessToken, nextSong.TrackId)
	// // track, ok := s.Cache.TrackMap[nextSong.TrackId]
	// if ok != nil {
	// 	panic("Track should always exist in cache")
	// }
	// album, ok := s.Cache.GetAlbum(s.SpotifyToken.AccessToken, nextSong.AlbumId)
	// // album, ok := s.Cache.AlbumMap[nextSong.AlbumId]
	// if ok != nil {
	// 	panic("Track should always exist in cache")
	// }
	// artist, ok := s.Cache.GetArtist(s.SpotifyToken.AccessToken, nextSong.ArtistId)
	// //artist, ok := s.Cache.ArtistMap[nextSong.ArtistId]
	// if ok != nil {
	// 	panic("Track should always exist in cache")
	// }
	// TODO: FIX THIS
	track := m.TrackData{}
	artist := m.ArtistData{}
	album := m.AlbumData{}

	// guessSong process:
	s.GuessState.SetTitle(track.Name, artist.Name, album.ImagesURL)
	s.MusicPlayer.Timer = time.Now()
	s.MusicPlayer.SongDuration = time.Duration(track.DurationMs) * time.Millisecond

	// err = s.EnsureAccessToken(ctx)
	// if err != nil {
	// 	return fmt.Errorf("Couldnt not ensure refresh token")
	// }
	return nil //, s.SpotifyApi.PlaySong(s.SpotifyToken.AccessToken, nextSong.TrackId)
}

// ClearQueue clears the current music queue
func (s *GameService) ClearQueue() error {
	s.AlbumSelection = make(map[string]bool)
	s.ArtistSelection = make(map[string]uint8)
	s.GuessState = NewGameState()
	s.MusicPlayer.ClearQueue()
	//s.SpotifyApi.PausePlayback(s.SpotifyToken.AccessToken)
	return nil
}

func (s *GameService) RequestUserAuthoritazion() (string, error) {
	urlString := "to Implement"
	// urlString, err := s.SpotifyApi.AuthRequestURL()
	return urlString, nil
}

func (s *GameService) ExchangeToken(ctx context.Context, state, code string) error {
	fmt.Println("Hello!")
	// if code == "" || state == "" {
	// 	return fmt.Errorf("Error guetting code and state from spotify api")
	// }
	// err := s.SpotifyApi.ValidateState(state)
	// if err != nil {
	// 	return err
	// }
	//
	// spotifyTokenReponse, err := s.SpotifyApi.TokenExchange(code)
	// if err != nil {
	// 	return err
	// }

	// expires_at := time.Now().Add(time.Duration(spotifyTokenReponse.ExpiresIn) * time.Second)
	// s.SpotifyToken = repository.SpotifyToken{
	// 	RefreshToken: spotifyTokenReponse.RefreshToken,
	// 	AccessToken:  spotifyTokenReponse.AccessToken,
	// 	TokenType:    spotifyTokenReponse.TokenType,
	// 	Scope:        spotifyTokenReponse.Scope,
	// 	ExpiresAt:    expires_at,
	// }
	//
	// err = s.SpotifyTokenStore.Update(ctx, s.UserId, s.SpotifyToken.AccessToken, s.SpotifyToken.ExpiresAt)
	// if err != nil {
	// 	fmt.Println("Error updating accessToken in DB")
	// 	return err
	// }
	return nil
}
func (s *GameService) EnsureAccessToken(ctx context.Context) error {
	//Read from DB
	if s.SpotifyToken.RefreshToken == "" {
		fmt.Println("Empty Refresh token")
		storeSpotifyToken, err := s.SpotifyTokenStore.Get(ctx, s.UserId)
		if err != nil {
			fmt.Println("Failed reading exchange token from DB")
			return err
		}
		s.SpotifyToken.AccessToken = storeSpotifyToken.AccessToken
		s.SpotifyToken.RefreshToken = storeSpotifyToken.RefreshToken
		s.SpotifyToken.ExpiresAt = storeSpotifyToken.ExpiresAt
		fmt.Println("Read exchange token: ", storeSpotifyToken.AccessToken)
	}

	// Expired?
	if time.Now().After(s.SpotifyToken.ExpiresAt) {
		if s.SpotifyToken.RefreshToken == "" {
			return fmt.Errorf("Refresh token is empty")
		}

		// spotifyRefreshReponse :=
		// spotifyRefreshReponse, err := s.SpotifyApi.RegenerateToken()
		// if err != nil {
		// 	return err
		// }
		//
		// expires_at := time.Now().Add(time.Duration(spotifyRefreshReponse.ExpiresIn) * time.Second)
		s.SpotifyToken = repository.SpotifyToken{}

		// update token
		// err = s.SpotifyTokenStore.Update(ctx, s.UserId, s.SpotifyToken.AccessToken, s.SpotifyToken.ExpiresAt)
		// if err != nil {
		// 	return err
		// }
	}
	return nil
}

func (s *GameService) IsUserSpotifyConnected() bool {
	if s == nil {
		return false
	}

	return s.SpotifyToken.AccessToken != ""

}
