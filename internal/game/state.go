package game

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type GuessState struct {
	title *TitleGuessState
}

func (g GuessState) ShowState() {
	fmt.Println("Current Guess state title RealTitle: ", g.title.RealTitle)
	fmt.Println("Current Guess state map: ", g.title.TitleAliveWords)
}

func NewGameState(trackName string) *GuessState {
	return &GuessState{
		title: NewTitleGuessState(trackName),
	}

}

type TitleGuessState struct {
	RealTitle       string
	TitleAliveWords map[string]uint8
}

func NewTitleGuessState(titleName string) *TitleGuessState {
	// array of words
	words := strings.Split((strings.ToLower(titleName)), " ")

	// count ords
	wordsCounts := make(map[string]uint8, len(words))

	for _, w := range words {
		wNormalized := CleanText(w)
		fmt.Printf("w: %v, wNorm: %v", w, wNormalized)
		wordsCounts[wNormalized] += 1
	}

	return &TitleGuessState{
		RealTitle:       titleName,
		TitleAliveWords: wordsCounts,
	}
}

func (g *GuessState) Guess(text string) (string, bool) {

	fmt.Println("Guessing... ", text)
	fmt.Println("Current Guess state title RealTitle: ", g.title.RealTitle)
	fmt.Println("Current Guess state map: ", g.title.TitleAliveWords)

	// update Guess
	g.title.updateGuessState(text)

	return g.title.showGuessState(), len(g.title.TitleAliveWords) == 0
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

func (t TitleGuessState) showGuessState() string {

	if len(t.TitleAliveWords) == 0 {
		return t.RealTitle
	}

	output := ""

	wordsInTitle := strings.Split(t.RealTitle, " ")
	if len(wordsInTitle) <= 0 {
		panic("title hsould have words?")
	}
	for _, w := range wordsInTitle {
		wLow := CleanText(strings.ToLower(w))
		_, exits := t.TitleAliveWords[wLow]
		if !exits {
			output += w + " "
		} else {
			for range len(w) {
				output += "_ "
			}
			output += "    "
		}
	}

	return output
}

func CleanText(s string) string {
	// Normalize to decomposed form (NFD)
	t := norm.NFD.String(s)
	result := make([]rune, 0, len(t))

	for _, r := range t {
		// Skip diacritical marks
		if unicode.Is(unicode.Mn, r) || (!unicode.IsLetter(r) && !unicode.IsNumber(r)) {
			continue
		}
		result = append(result, r)
	}

	return string(result)
}
