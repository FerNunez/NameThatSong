package player

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// SpotifyPlayer implements the MusicPlayer interface for Spotify
type SpotifyPlayer struct {
	Queue              []Song
	CurrentIndex       int
	Playing            bool
	MusicServiceClient MusicServiceClient
}

// NewSpotifyPlayer creates a new SpotifyPlayer
func NewSpotifyPlayer(musicServiceClient MusicServiceClient) *SpotifyPlayer {
	return &SpotifyPlayer{
		Queue:              []Song{},
		CurrentIndex:       0,
		Playing:            false,
		MusicServiceClient: musicServiceClient,
	}
}

// AddToQueue adds songs to the queue
func (p *SpotifyPlayer) AddToQueue(songs []Song) {
	p.Queue = append(p.Queue, songs...)
}

// GetQueue returns the current queue
func (p *SpotifyPlayer) GetQueue() []Song {
	return p.Queue
}

// ClearQueue clears the queue
func (p *SpotifyPlayer) ClearQueue() {
	p.Queue = []Song{}
	p.CurrentIndex = 0
	p.Playing = false
}

// Shuffle shuffles the queue
func (p *SpotifyPlayer) Shuffle() {
	if len(p.Queue) <= 1 {
		return
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(p.Queue), func(i, j int) {
		p.Queue[i], p.Queue[j] = p.Queue[j], p.Queue[i]
	})

	p.CurrentIndex = 0
}

// Play starts playing the current song in the queue
func (p *SpotifyPlayer) Play() error {
	if len(p.Queue) == 0 {
		return errors.New("queue is empty")
	}

	currentSong := p.Queue[p.CurrentIndex]
	fmt.Printf("Playing song: %s\n", currentSong.Name)

	err := p.MusicServiceClient.PlaySong(currentSong.ID)
	if err != nil {
		return fmt.Errorf("failed to play song: %v", err)
	}

	p.Playing = true
	return nil
}

// Pause pauses the current playback
func (p *SpotifyPlayer) Pause() error {
	if !p.Playing {
		return errors.New("not currently playing")
	}

	err := p.MusicServiceClient.PausePlayback()
	if err != nil {
		return fmt.Errorf("failed to pause playback: %v", err)
	}

	p.Playing = false
	return nil
}

// Resume resumes the current playback
func (p *SpotifyPlayer) Resume() error {
	if p.Playing {
		return errors.New("already playing")
	}

	if len(p.Queue) == 0 {
		return errors.New("queue is empty")
	}

	err := p.MusicServiceClient.ResumePlayback()
	if err != nil {
		return fmt.Errorf("failed to resume playback: %v", err)
	}

	p.Playing = true
	return nil
}

// Skip skips to the next song
func (p *SpotifyPlayer) Skip() error {
	if len(p.Queue) == 0 {
		return errors.New("queue is empty")
	}

	p.CurrentIndex = (p.CurrentIndex + 1) % len(p.Queue)

	if p.Playing {
		return p.Play()
	}

	return nil
}

// GetCurrentSong returns the current song
func (p *SpotifyPlayer) GetCurrentSong() *Song {
	if len(p.Queue) == 0 {
		return nil
	}

	return &p.Queue[p.CurrentIndex]
}

// IsPlaying returns whether playback is active
func (p *SpotifyPlayer) IsPlaying() bool {
	return p.Playing
}
