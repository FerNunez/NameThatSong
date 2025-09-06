package songs

import (
	"context"
	"fmt"
	"time"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
	"github.com/FerNunez/NameThatSong/internal/repository"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
)

// PlaylistService defines the contract for playlist service operations
type Service interface {
	StoreSong(ctx context.Context, song models.Song) error
	GetSongBySpotifyID(ctx context.Context, userID, spotifyTrackID string) (models.Song, error)
}

type SongProvider struct {
	SongStore      repository.SongsStore
	SpotifyService spotify.SpotifyService
}

func NewSongProvider(SongStore repository.SongsStore, SpotifyService spotify.SpotifyService) Service {
	return &SongProvider{
		SongStore,
		SpotifyService,
	}
}

func (s *SongProvider) StoreSong(ctx context.Context, song models.Song) error {
	return s.SongStore.UpsertSong(ctx, song)
}

func (s *SongProvider) GetSongBySpotifyID(ctx context.Context, userID, spotifyTrackID string) (models.Song, error) {
	song, err := s.SongStore.GetSongBySpotifyID(ctx, spotifyTrackID)
	if err == nil {
		logger.Debug(ctx, "song found in db")
		return song, nil
	}
	logger.Debug(ctx, "song not found in db, fetching from spotify")

	// add user ID
	trackData, err := s.SpotifyService.FetchTrack(ctx, userID, models.SpotifyID(spotifyTrackID))
	if err != nil {
		return models.Song{}, err
	}
	
	if len(trackData.ArtistIDs) == 0 {
		logger.Error(ctx, "Track has no artists!")
		return models.Song{}, fmt.Errorf("track has no artists")
	}

	// Fetch album data
	albumData, err := s.SpotifyService.FetchAlbum(ctx, userID, trackData.AlbumID)
	if err != nil {
		return models.Song{}, fmt.Errorf("failed to fetch album: %w", err)
	}

	// Fetch primary artist data (first artist)
	artistData, err := s.SpotifyService.FetchArtist(ctx, userID, trackData.ArtistIDs[0])
	if err != nil {
		return models.Song{}, fmt.Errorf("failed to fetch artist: %w", err)
	}

	song = models.Song{
		SpotifyTrackID:   spotifyTrackID,
		SpotifyAlbumID:   string(trackData.AlbumID),
		SpotifyArtistID:  string(trackData.ArtistIDs[0]),
		TrackName:        trackData.Name,
		ArtistName:       artistData.Name,
		AlbumName:        albumData.Name,
		SpotifyArtistURL: artistData.ImageURL,
		SpotifyAlbumURL:  albumData.ImageURL,
		DurationMs:       trackData.DurationMs,
		UpdatedAt:        time.Now(),
	}
	err = s.SongStore.UpsertSong(ctx, song)
	if err != nil {
		logger.Debug(ctx, "couldnt store data", logger.F("track id", song.SpotifyTrackID), logger.F("error", err))
	}

	return song, nil
}
