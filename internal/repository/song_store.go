package repository

import (
	"context"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/repository/database"
)

type SongsStore interface {
	UpsertSong(ctx context.Context, song models.Song) error
	GetSongBySpotifyID(ctx context.Context, spotify_track_id string) (models.Song, error)
}

type SQLSongStore struct {
	db *database.Queries
}

func NewSQLSongStore(db *database.Queries) SongsStore {
	return &SQLSongStore{db}
}

func (s *SQLSongStore) UpsertSong(ctx context.Context, song models.Song) error {
	_, err := s.db.UpsertSong(ctx, database.UpsertSongParams{
		SpotifyTrackID:  song.SpotifyTrackID,
		SpotifyAlbumID:  song.SpotifyAlbumID,
		SpotifyArtistID: song.SpotifyArtistID,
		TrackName:       song.TrackName,
		AlbumName:       song.AlbumName,
		ArtistName:      song.ArtistName,
		ArtistImageUrl:  song.SpotifyArtistURL,
		AlbumImageUrl:   song.SpotifyAlbumURL,
		DurationMs:      int32(song.DurationMs),
	})
	return err
}

func (s *SQLSongStore) GetSongBySpotifyID(ctx context.Context, spotify_track_id string) (models.Song, error) {
	dbSong, err := s.db.GetSongBySpotifyID(ctx, spotify_track_id)
	if err != nil {
		return models.Song{}, err
	}
	return models.Song{
		SpotifyTrackID:   dbSong.SpotifyTrackID,
		SpotifyAlbumID:   dbSong.SpotifyAlbumID,
		SpotifyArtistID:  dbSong.SpotifyArtistID,
		TrackName:        dbSong.TrackName,
		ArtistName:       dbSong.ArtistName,
		AlbumName:        dbSong.AlbumName,
		SpotifyAlbumURL:  dbSong.AlbumImageUrl,
		SpotifyArtistURL: dbSong.ArtistImageUrl,
		DurationMs:       int(dbSong.DurationMs),
		UpdatedAt:        dbSong.UpdatedAt,
	}, err

}
