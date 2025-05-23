package player

import (
	"errors"
	"fmt"
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
	Timer        time.Time
	SongDuration time.Duration
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
func (p *MusicPlayer) SongOver() bool {
	return time.Since(p.Timer) >= p.SongDuration
}

func (p *MusicPlayer) GetTimerAsString() string {
	timerElapsed := time.Since(p.Timer)
	return DurationToString(timerElapsed)
}

func (p *MusicPlayer) GetSongDurationAsString() string {
	return DurationToString(p.SongDuration)
}
func DurationToString(d time.Duration) string {
	// Get minutes and seconds
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60

	// Format as "m:ss"
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
