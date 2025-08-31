# Game & Playlist Service Design - MVP

## Overview
Single-player music guessing game with playlist management. Two main services: GameService and PlaylistService.

## Service Interfaces

### GameService Interface
```go
type GameService interface {
    // Game Lifecycle
    StartGame(userID, playlistID, difficulty string, totalRounds int) (*GameSession, error)
    // Creates new game session, generates first round, returns game state
    
    GetGameState(gameID string) (*GameSession, error)  
    // Returns current game session with all rounds and progress
    
    SubmitAnswer(gameID, answer string) (*GameRound, error)
    // Process user answer, calculate points, advance to next round or finish game
    
    GetCurrentRound(gameID string) (*GameRound, error)
    // Returns active round with question data (song preview, choices, etc.)
    
    FinishGame(gameID string) (*GameSession, error)
    // End game early, calculate final score, update user stats
    
    // Statistics
    GetUserStats(userID string) (*UserGameStats, error)
    // Return user's game history, best scores, accuracy rates
}
```

### PlaylistService Interface
```go
type PlaylistService interface {
    // Playlist Management
    ImportUserPlaylists(userID string) ([]Playlist, error)
    // Fetch all user playlists from Spotify, store locally with metadata
    
    GetUserPlaylists(userID string) ([]Playlist, error)
    // Return user's playlists (for game playlist selection)
    
    GetPlaylistSongs(playlistID string) ([]Song, error)
    // Return all songs in a playlist (for game song selection)
    
    AnalyzePlaylist(playlistID string) (*PlaylistAnalysis, error)
    // Analyze genres, duplicates, audio features, generate insights
    
    SyncPlaylist(playlistID string) error
    // Update local playlist with latest Spotify changes
    
    // Playlist Operations
    GetPlaylistDetails(playlistID string) (*Playlist, error)
    // Get single playlist with metadata and stats
}
```

## Data Models

### GameSession
```go
type GameSession struct {
    // Identification
    ID          string    `json:"id"`           // Unique game identifier
    UserID      string    `json:"user_id"`      // Player who started the game
    
    // Game Configuration
    SourcePlaylistID string `json:"source_playlist_id"` // Which playlist to pull songs from
    Difficulty       string `json:"difficulty"`         // easy/medium/hard
    TotalRounds      int    `json:"total_rounds"`       // How many questions (e.g., 10)
    QuestionType     string `json:"question_type"`      // "guess_song", "guess_artist", "multiple_choice"
    
    // Game State
    Status       string `json:"status"`        // "active", "finished", "paused"
    CurrentRound int    `json:"current_round"` // Which question we're on (1-10)
    Score        int    `json:"score"`         // Total points accumulated
    
    // Timestamps
    CreatedAt  time.Time `json:"created_at"`
    StartedAt  *time.Time `json:"started_at"`  // When first question was shown
    FinishedAt *time.Time `json:"finished_at"` // When game completed
}
```

### GameRound
```go
type GameRound struct {
    // Identification
    GameID      string `json:"game_id"`      // Links to parent game session
    RoundNumber int    `json:"round_number"` // 1, 2, 3... up to total_rounds
    
    // Question Data
    SongID         string   `json:"song_id"`         // Song being asked about
    CorrectAnswer  string   `json:"correct_answer"`  // The right answer
    QuestionText   string   `json:"question_text"`   // "What song is this?"
    AnswerChoices  []string `json:"answer_choices"`  // For multiple choice [A, B, C, D]
    
    // User Response
    UserAnswer   *string    `json:"user_answer"`   // What user typed/selected
    IsCorrect    *bool      `json:"is_correct"`    // true/false/null if not answered
    ResponseTime *int       `json:"response_time"` // Milliseconds to answer
    PointsEarned int        `json:"points_earned"` // Points for this round
    
    // Timestamps
    StartedAt  time.Time  `json:"started_at"`  // When question was presented
    AnsweredAt *time.Time `json:"answered_at"` // When user submitted answer
}
```

### Playlist Models
```go
type Playlist struct {
    ID               string    `json:"id"`
    UserID           string    `json:"user_id"`
    SpotifyPlaylistID string   `json:"spotify_playlist_id"`
    Name             string    `json:"name"`
    Description      string    `json:"description"`
    SongCount        int       `json:"song_count"`
    TotalDuration    int       `json:"total_duration"` // in seconds
    IsAnalyzed       bool      `json:"is_analyzed"`
    LastSync         time.Time `json:"last_sync"`
    CreatedAt        time.Time `json:"created_at"`
}

type PlaylistAnalysis struct {
    PlaylistID       string            `json:"playlist_id"`
    Genres           map[string]int    `json:"genres"`          // genre -> count
    AvgEnergy        float64           `json:"avg_energy"`
    AvgDanceability  float64           `json:"avg_danceability"`
    DuplicateSongs   []string          `json:"duplicate_songs"` // song IDs
    Recommendations  []string          `json:"recommendations"` // suggested song IDs
}

type UserGameStats struct {
    UserID           string  `json:"user_id"`
    GamesPlayed      int     `json:"games_played"`
    BestScore        int     `json:"best_score"`
    AverageScore     float64 `json:"average_score"`
    CorrectAnswers   int     `json:"correct_answers"`
    TotalQuestions   int     `json:"total_questions"`
    AccuracyRate     float64 `json:"accuracy_rate"`
}
```

## User Flow Diagram

```
1. User Login
   ↓
2. PlaylistService.ImportUserPlaylists(userID)
   → Fetches playlists from Spotify
   → Stores locally: [Work Playlist, Chill Vibes, 90s Hits, etc.]
   ↓
3. Game Start Screen
   → User sees list from PlaylistService.GetUserPlaylists()
   → User selects: "90s Hits" + "Medium" difficulty + 10 rounds
   ↓
4. GameService.StartGame(userID, "90s-hits-id", "medium", 10)
   → Calls PlaylistService.GetPlaylistSongs("90s-hits-id")
   → Gets: [Song1, Song2, Song3... Song50]
   → Randomly picks 10 songs for game
   → Creates GameSession + first GameRound
   ↓
5. Game Loop:
   → GameService.GetCurrentRound() → Shows question
   → User answers → GameService.SubmitAnswer()
   → Repeat 10 times
   ↓
6. Game End
   → Final score, update user stats
```

## Service Communication Flow

- **GameService** ↔ **PlaylistService**: Get playlist songs for game generation
- **GameService** ↔ **SpotifyService**: Fetch song metadata, preview URLs
- **PlaylistService** ↔ **SpotifyService**: Import playlists, sync changes
- **Both Services** ↔ **UserService**: User authentication, profile data

## Implementation Priority

1. **Phase 1**: Basic PlaylistService (import, get playlists/songs)
2. **Phase 2**: Basic GameService (start game, submit answer, single question type)
3. **Phase 3**: Enhanced features (multiple question types, analytics, stats)
4. **Phase 4**: Advanced playlist features (analysis, duplicates, recommendations)

## Technical Notes

- Single-player MVP first (no multiplayer complexity)
- Use existing UserService and SpotifyService
- Cache service already implemented for performance
- Audio playback handled by frontend using Spotify Web Playback SDK
- Question types: "guess_song", "guess_artist", "multiple_choice"
- Difficulty affects: song popularity, snippet length, answer choices