package player

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// SpotifyPlayer implements the MusicPlayer interface for Spotify
type SpotifyPlayer struct {
	Queue        []Song
	CurrentIndex int
	Playing      bool
	DeviceID     string
	AccessToken  string
}

// NewSpotifyPlayer creates a new SpotifyPlayer
func NewSpotifyPlayer(deviceID, accessToken string) *SpotifyPlayer {
	return &SpotifyPlayer{
		Queue:        []Song{},
		CurrentIndex: 0,
		Playing:      false,
		DeviceID:     deviceID,
		AccessToken:  accessToken,
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

	// Here you would implement the actual Spotify API call to play the song
	// Example implementation:
	// Call Spotify API to play the song at p.Queue[p.CurrentIndex].URI on device p.DeviceID
	fmt.Printf("Playing song: %s on device: %s\n", p.Queue[p.CurrentIndex].Name, p.DeviceID)

	p.Playing = true
	return nil
}

// Pause pauses the current playback
func (p *SpotifyPlayer) Pause() error {
	if !p.Playing {
		return errors.New("not currently playing")
	}

	// Call Spotify API to pause playback
	fmt.Println("Pausing playback")

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

	// Call Spotify API to resume playback
	fmt.Println("Resuming playback")

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
