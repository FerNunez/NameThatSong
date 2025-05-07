package store

import (
	"context"
)

type GameStore interface {
	DeleteAlbumSelection(ctx context.Context, albumId string) error
	DecrementArtistSelection(ctx context.Context, artistId string) error
	InsertAlbumSelection(ctx context.Context, albumId, artistId string) error
	AddTrackToQueue(ctx context.Context, trackId string) error
	ClearMusicQueue(ctx context.Context) error
}
