package player

import (
	"errors"
	"math/rand"
	"time"
)

type MusicPlayer struct {
	Queue        []string
	CurrentIndex int
}

func NewMusicPlayer() *MusicPlayer {
	return &MusicPlayer{
		Queue:        []string{},
		CurrentIndex: 0,
	}
}

func (p *MusicPlayer) AddToQueue(songsId []string) {
	p.Queue = append(p.Queue, songsId...)
}

func (p *MusicPlayer) GetQueue() []string {
	return p.Queue
}

func (p *MusicPlayer) ClearQueue() {
	p.Queue = []string{}
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

func (p *MusicPlayer) NextInQueue() (string, error) {
	if p.CurrentIndex >= len(p.Queue) {
		return "", errors.New("cannot next song as last song in the queue")
	}

	p.CurrentIndex += 1
	return p.Queue[p.CurrentIndex], nil

}
