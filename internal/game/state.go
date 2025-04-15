package game

import (
	"fmt"
	"strings"
)

type GuessState struct {
	Title          *TitleGuessState
	points         int
	correctGuesses int
}

func NewGameState() *GuessState {
	return &GuessState{
		Title:          NewTitleGuessState(""),
		points:         0,
		correctGuesses: 0,
	}
}

func (g *GuessState) SetTitle(trackName string) {
	g.Title = NewTitleGuessState(trackName)
}

type TitleGuessState struct {
	RealTitle       string
	TitleAliveWords map[string]uint8
}

func NewTitleGuessState(titleName string) *TitleGuessState {
	// array of words
	words := strings.Split(CleanText(titleName), " ")

	// count ords
	wordsCounts := make(map[string]uint8, len(words))

	for _, w := range words {
		if w == "" {
			continue
		}
		wordsCounts[w] += 1
	}

	fmt.Printf("wc: %v\n", wordsCounts)

	return &TitleGuessState{
		RealTitle:       titleName,
		TitleAliveWords: wordsCounts,
	}
}

func (g *GuessState) Guess(text string) (string, bool) {
	// update Guess
	g.Title.updateGuessState(text)

	// Check if all words are guessed
	allGuessed := len(g.Title.TitleAliveWords) == 0
	if allGuessed {
		g.points += 100 // Award 100 points for a correct guess
		g.correctGuesses++
	}

	return g.Title.ShowGuessState(), allGuessed
}

func (g *GuessState) GetPoints() int {
	return g.points
}

func (g *GuessState) GetCorrectGuesses() int {
	return g.correctGuesses
}

func (t *TitleGuessState) updateGuessState(text string) {
	words := strings.Split(text, " ")

	for _, w := range words {
		wLow := CleanText(strings.ToLower(w))
		remaining, ok := t.TitleAliveWords[wLow]
		if ok {
			if remaining == 1 {
				delete(t.TitleAliveWords, wLow)
				continue
			}
			t.TitleAliveWords[wLow] -= 1
		}
	}
}

func (t TitleGuessState) ShowGuessState() string {

	return ProcessState(t.RealTitle, t.TitleAliveWords)
}
