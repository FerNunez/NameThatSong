package player

// Song represents a playable song entity
type Song struct {
	ID       string
	Name     string
	URI      string
	Duration int
	AlbumID  string
	Artists  []string
}

// MusicPlayer defines the interface for playing and managing songs
type MusicPlayer interface {
	// Queue management
	AddToQueue(songs []Song)
	GetQueue() []Song
	ClearQueue()
	Shuffle()

	// Playback control
	Play() error
	Pause() error
	Resume() error
	Skip() error

	// State information
	GetCurrentSong() *Song
	IsPlaying() bool
}

// SongProvider defines the interface for retrieving songs
type SongProvider interface {
	// Get songs by IDs
	GetSongsByIDs(ids []string) ([]Song, error)

	// Get songs for an album
	GetSongsByAlbumID(albumID string) ([]Song, error)

	// Get all songs for multiple albums
	GetSongsByAlbumIDs(albumIDs []string) (map[string][]Song, error)
}

// MusicServiceClient defines the interface for interacting with a music service API
type MusicServiceClient interface {
	// Playback control
	PlaySong(songID string) error
	PausePlayback() error
	ResumePlayback() error
	SkipToNext() error

	// Get the currently active device ID
	GetActiveDevice() (string, error)
}
