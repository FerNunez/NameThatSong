package player

import (
	"errors"
	"math/rand"
	"time"
)

type Song struct {
	TrackId  string
	AlbumId  string
	ArtistId string
}

func NewSong(trackId string, albumId string, artistId string) *Song {
	return &Song{
		TrackId:  trackId,
		AlbumId:  albumId,
		ArtistId: artistId,
	}
}

type MusicPlayer struct {
	Queue        []Song
	CurrentIndex int
}

func NewMusicPlayer() *MusicPlayer {
	return &MusicPlayer{
		Queue:        []Song{},
		CurrentIndex: 0,
	}
}

func (p *MusicPlayer) AddToQueue(songs []Song) {
	p.Queue = append(p.Queue, songs...)
}

func (p *MusicPlayer) GetQueue() []Song {
	return p.Queue
}

func (p *MusicPlayer) ClearQueue() {
	p.Queue = []Song{}
	p.CurrentIndex = 0
}

// Shuffle shuffles the queue
func (p *MusicPlayer) Shuffle() {
	if len(p.Queue) <= 1 {
		return
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(p.Queue), func(i, j int) {
		p.Queue[i], p.Queue[j] = p.Queue[j], p.Queue[i]
	})

	p.CurrentIndex = 0
}

func (p *MusicPlayer) NextInQueue() (Song, error) {
	if p.CurrentIndex >= len(p.Queue) {
		return Song{}, errors.New("cannot next song as last song in the queue")
	}

	p.CurrentIndex += 1
	return p.Queue[p.CurrentIndex], nil

}
