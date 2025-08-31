package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/FerNunez/NameThatSong/internal/api/middleware"
	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
	"github.com/FerNunez/NameThatSong/internal/services/playlist"
	"github.com/FerNunez/NameThatSong/web/templates"
)

// Temporary game state storage (in production, this would be in database/redis)
var activeGames = make(map[string]*GameSession)

type GameSession struct {
	ID           string        `json:"id"`
	UserID       string        `json:"user_id"`
	PlaylistID   string        `json:"playlist_id"`
	Difficulty   string        `json:"difficulty"`
	TotalRounds  int           `json:"total_rounds"`
	CurrentRound int           `json:"current_round"`
	Score        int           `json:"score"`
	Status       string        `json:"status"`
	TimeLeft     int           `json:"time_left"`
	Songs        []models.Song `json:"songs"`
	CurrentSong  *models.Song  `json:"current_song"`
	Answers      []GameAnswer  `json:"answers"`
}

type GameAnswer struct {
	Round         int    `json:"round"`
	UserAnswer    string `json:"user_answer"`
	CorrectAnswer string `json:"correct_answer"`
	IsCorrect     bool   `json:"is_correct"`
	Points        int    `json:"points"`
	ResponseTime  int    `json:"response_time"`
}

type GameHandler struct {
	playlistService playlist.Service
}

func NewGameHandler(ps playlist.Service) *GameHandler {
	return &GameHandler{
		playlistService: ps,
	}
}

// GameSetupPage method removed - game setup is now integrated into modern UI
// Use /api/game/setup endpoint in playlist_handler.go instead

// GET /api/playlists - Get user playlists for game selection
func (h *GameHandler) GetUserPlaylists(w http.ResponseWriter, r *http.Request) {
	logger.Info(r.Context(), "fetching user playlists for game")

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's playlists
	playlists, err := h.playlistService.GetUserPlaylists(r.Context(), user.ID)
	if err != nil {
		logger.Error(r.Context(), "failed to get user playlists",
			logger.F("user_id", user.ID.String()),
			logger.F("error", err))
		http.Error(w, "failed to get playlists", http.StatusInternalServerError)
		return
	}

	logger.Info(r.Context(), "got user playlists", logger.F("playlist_size", len(playlists)))
	// Convert []*models.Playlist to []models.Playlist
	playlistsSlice := make([]models.Playlist, len(playlists))
	for i, p := range playlists {
		playlistsSlice[i] = *p
	}

	// NOTE: PlaylistGrid template removed - using GamePlaylistSelection instead
	// This method is deprecated, use playlist_handler.GetGamePlaylists instead
	http.Error(w, "deprecated endpoint - use /api/game/playlists instead", http.StatusGone)
}

// POST /game/start - Start a new game
func (h *GameHandler) StartGame(w http.ResponseWriter, r *http.Request) {
	logger.Info(r.Context(), "starting new game")

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	playlistID := r.FormValue("selected_playlist")
	difficulty := r.FormValue("difficulty")
	roundsStr := r.FormValue("rounds")

	rounds, err := strconv.Atoi(roundsStr)
	if err != nil {
		rounds = 10 // default
	}

	logger.Info(r.Context(), "creating new game session",
		logger.F("user_id", user.ID.String()),
		logger.F("playlist_id", playlistID),
		logger.F("difficulty", difficulty),
		logger.F("rounds", rounds))

	// Get playlist with songs
	// playlistWithSongs, err := h.playlistService.GetPlaylistWithSongs(r.Context(), uuid.MustParse(playlistID), user.ID)
	// if err != nil {
	// 	logger.Error(r.Context(), "failed to get playlist songs",
	// 		logger.F("error", err))
	// 	http.Error(w, "failed to get playlist songs", http.StatusInternalServerError)
	// 	return
	// }
	//
	// if len(playlistWithSongs.Songs) < rounds {
	// 	http.Error(w, "playlist doesn't have enough songs", http.StatusBadRequest)
	// 	return
	// }
	//
	// // Create game session
	// gameID := generateGameID()
	// game := &GameSession{
	// 	ID:           gameID,
	// 	UserID:       user.ID.String(),
	// 	PlaylistID:   playlistID,
	// 	Difficulty:   difficulty,
	// 	TotalRounds:  rounds,
	// 	CurrentRound: 1,
	// 	Score:        0,
	// 	Status:       "active",
	// 	TimeLeft:     getTimeLimit(difficulty),
	// 	Songs:        selectRandomSongs(playlistWithSongs.Songs, rounds),
	// 	Answers:      make([]GameAnswer, 0),
	// }
	//
	// // Set current song
	// if len(game.Songs) > 0 {
	// 	game.CurrentSong = &game.Songs[0]
	// }
	//
	// // Store game session
	// activeGames[gameID] = game
	//
	// logger.Info(r.Context(), "game session created",
	// 	logger.F("game_id", gameID),
	// 	logger.F("user_id", user.ID.String()))
	//
	// // Render game interface
	// gameState := templates.GameState{
	// 	ID:           game.ID,
	// 	CurrentRound: game.CurrentRound,
	// 	TotalRounds:  game.TotalRounds,
	// 	Score:        game.Score,
	// 	TimeLeft:     game.TimeLeft,
	// 	Status:       game.Status,
	// 	Difficulty:   game.Difficulty,
	// }
	//
	// currentSong := templates.CurrentSong{
	// 	ID:         game.CurrentSong.SpotifyTrackID,
	// 	Title:      game.CurrentSong.TrackName,
	// 	Artist:     game.CurrentSong.ArtistName,
	// 	Album:      "", // TODO: Add album info to PlaylistSong
	// 	AlbumArt:   "", // TODO: Add album art to PlaylistSong
	// 	PreviewURL: "", // TODO: Add preview URL to PlaylistSong or fetch from Spotify
	// 	IsPlaying:  false,
	// }
	//
	// component := templates.GameActivePage(gameState, currentSong)
	// if err := component.Render(r.Context(), w); err != nil {
	// 	logger.Error(r.Context(), "failed to render game page",
	// 		logger.F("error", err))
	// 	http.Error(w, "failed to render game", http.StatusInternalServerError)
	// 	return
	// }
	http.Error(w, "failed to render game", http.StatusInternalServerError)
}

// POST /game/submit-answer - Submit an answer
func (h *GameHandler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("id")
	game, ok := activeGames[gameID]
	if !ok {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	userAnswer := r.FormValue("answer")
	correctAnswer := game.CurrentSong.TrackName

	// Check if answer is correct (simple string matching for now)
	isCorrect := checkAnswer(userAnswer, correctAnswer)
	points := calculatePoints(isCorrect, game.Difficulty, game.TimeLeft)

	// Record answer
	answer := GameAnswer{
		Round:         game.CurrentRound,
		UserAnswer:    userAnswer,
		CorrectAnswer: correctAnswer,
		IsCorrect:     isCorrect,
		Points:        points,
		ResponseTime:  getTimeLimit(game.Difficulty) - game.TimeLeft,
	}

	game.Answers = append(game.Answers, answer)
	if isCorrect {
		game.Score += points
	}

	// Move to next round or finish game
	if game.CurrentRound >= game.TotalRounds {
		// Game finished
		game.Status = "completed"
		h.renderGameResults(w, r, game)
		return
	}

	// Next round
	game.CurrentRound++
	game.TimeLeft = getTimeLimit(game.Difficulty)

	if game.CurrentRound <= len(game.Songs) {
		game.CurrentSong = &game.Songs[game.CurrentRound-1]
	}

	// Show answer feedback first, then next round
	h.showAnswerFeedback(w, r, game, answer)
}

// GET /game/timer - Get current timer value
func (h *GameHandler) GetTimer(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("id")
	game, ok := activeGames[gameID]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Decrease timer (in production, this would be more sophisticated)
	if game.TimeLeft > 0 && game.Status == "active" {
		game.TimeLeft--
	}

	// Auto-submit if time runs out
	if game.TimeLeft <= 0 && game.Status == "active" {
		// Auto-submit empty answer
		answer := GameAnswer{
			Round:         game.CurrentRound,
			UserAnswer:    "",
			CorrectAnswer: game.CurrentSong.TrackName,
			IsCorrect:     false,
			Points:        0,
			ResponseTime:  getTimeLimit(game.Difficulty),
		}

		game.Answers = append(game.Answers, answer)

		if game.CurrentRound >= game.TotalRounds {
			game.Status = "completed"
			// Redirect to results
			w.Header().Set("HX-Redirect", "/game/results?id="+gameID)
			return
		}

		// Next round
		game.CurrentRound++
		game.TimeLeft = getTimeLimit(game.Difficulty)
		if game.CurrentRound <= len(game.Songs) {
			game.CurrentSong = &game.Songs[game.CurrentRound-1]
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(strconv.Itoa(game.TimeLeft) + "s"))
}

// GET /game/results - Show game results
func (h *GameHandler) GameResults(w http.ResponseWriter, r *http.Request) {
	gameID := r.URL.Query().Get("id")
	game, ok := activeGames[gameID]
	if !ok {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}

	h.renderGameResults(w, r, game)
}

// Helper functions
func (h *GameHandler) renderGameResults(w http.ResponseWriter, r *http.Request, game *GameSession) {
	correctAnswers := 0
	totalTime := 0
	pointsBreakdown := make([]templates.PointEntry, len(game.Answers))

	for i, answer := range game.Answers {
		if answer.IsCorrect {
			correctAnswers++
		}
		totalTime += answer.ResponseTime

		// Get song info
		var songTitle, artist string
		if answer.Round <= len(game.Songs) {
			song := game.Songs[answer.Round-1]
			songTitle = song.TrackName
			artist = song.ArtistName
		}

		pointsBreakdown[i] = templates.PointEntry{
			Round:        answer.Round,
			SongTitle:    songTitle,
			Artist:       artist,
			UserAnswer:   answer.UserAnswer,
			IsCorrect:    answer.IsCorrect,
			Points:       answer.Points,
			ResponseTime: answer.ResponseTime,
		}
	}

	accuracyRate := float64(correctAnswers) / float64(game.TotalRounds) * 100

	results := templates.GameResults{
		ID:              game.ID,
		TotalScore:      game.Score,
		TotalQuestions:  game.TotalRounds,
		CorrectAnswers:  correctAnswers,
		AccuracyRate:    accuracyRate,
		TotalTime:       totalTime,
		Difficulty:      game.Difficulty,
		PlaylistName:    "Selected Playlist", // TODO: get actual playlist name
		NewBestScore:    false,               // TODO: check against user's best scores
		PointsBreakdown: pointsBreakdown,
	}

	component := templates.GameResultsPage(results)
	if err := component.Render(r.Context(), w); err != nil {
		logger.Error(r.Context(), "failed to render game results",
			logger.F("error", err))
		http.Error(w, "failed to render results", http.StatusInternalServerError)
		return
	}
}

func (h *GameHandler) showAnswerFeedback(w http.ResponseWriter, r *http.Request, game *GameSession, answer GameAnswer) {
	// Return JSON for HTMX to handle with JavaScript
	feedback := map[string]interface{}{
		"correct":        answer.IsCorrect,
		"points":         answer.Points,
		"correct_answer": answer.CorrectAnswer,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feedback)
}

// Utility functions
func generateGameID() string {
	// Simple ID generation - in production use proper UUID
	return "game_" + strconv.FormatInt(int64(len(activeGames)+1), 10)
}

func getTimeLimit(difficulty string) int {
	switch difficulty {
	case "easy":
		return 30
	case "medium":
		return 20
	case "hard":
		return 15
	default:
		return 20
	}
}

func selectRandomSongs(songs []models.Song, count int) []models.Song {
	// Simple random selection - in production use proper shuffling
	if len(songs) <= count {
		return songs
	}
	return songs[:count]
}

func checkAnswer(userAnswer, correctAnswer string) bool {
	// Simple case-insensitive matching - in production use fuzzy matching
	return len(userAnswer) > 0 &&
		(userAnswer == correctAnswer ||
			len(userAnswer) > 3 &&
				contains(correctAnswer, userAnswer))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr
}

func calculatePoints(isCorrect bool, difficulty string, timeLeft int) int {
	if !isCorrect {
		return 0
	}

	basePoints := 100

	// Difficulty multiplier
	switch difficulty {
	case "easy":
		basePoints = 50
	case "medium":
		basePoints = 100
	case "hard":
		basePoints = 150
	}

	// Time bonus
	timeBonusPercent := timeLeft * 2 // 2% per second remaining
	timeBonus := basePoints * timeBonusPercent / 100

	return basePoints + timeBonus
}

// Helper functions for future Spotify API integration
// TODO: Implement these when we add full track metadata
