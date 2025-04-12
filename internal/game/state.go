package game

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type GuessState struct {
	title          *TitleGuessState
	points         int
	correctGuesses int
}

func NewGameState(trackName string) *GuessState {
	return &GuessState{
		title:          NewTitleGuessState(trackName),
		points:         0,
		correctGuesses: 0,
	}
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

	return &TitleGuessState{
		RealTitle:       titleName,
		TitleAliveWords: wordsCounts,
	}
}

func (g *GuessState) Guess(text string) (string, bool) {
	// update Guess
	g.title.updateGuessState(text)

	// Check if all words are guessed
	allGuessed := len(g.title.TitleAliveWords) == 0
	if allGuessed {
		g.points += 100 // Award 100 points for a correct guess
		g.correctGuesses++
	}

	return g.title.showGuessState(), allGuessed
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

func (t TitleGuessState) showGuessState() string {

	return ProcessState(t.RealTitle, t.TitleAliveWords)
}

func CleanText(s string) string {
	cleanText := RemoveParenthesis(strings.ToLower(s))
	cleanText = RemoveAfterWord(cleanText, "-")
	cleanText = RemoveAfterWord(cleanText, "feat.")
	cleanText = RemoveAfterWord(cleanText, "feature")
	cleanText = RemoveAccent(cleanText)

	return cleanText
}

func RemoveAccent(s string) string {
	// Normalize to decomposed form (NFD)
	t := norm.NFD.String(s)
	result := make([]rune, 0, len(t))
	for _, r := range t {
		// Skip diacritical marks
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		result = append(result, r)
	}
	return string(result)
}

func RemoveParenthesis(s string) string {

	splitted := strings.SplitN(s, "(", 2)

	if len(splitted) > 1 {
		if !strings.Contains(splitted[1], ")") {
			return strings.Join(splitted, "(")
		}
	}
	return splitted[0]

}

func RemoveAfterWord(s string, w string) string {
	if s == "" {
		return ""
	}
	splitted := strings.Split(s, w)

	return splitted[0]
}

func ProcessState(original string, aliveWords map[string]uint8) string {

	if len(aliveWords) == 0 {
		return original
	}

	solution := ""
	words := strings.Split(original, " ")
	for _, w := range words {

		// remove symbol
		noSymbol := ""
		for _, r := range strings.ToLower(w) {
			// remove symbol
			if unicode.IsLetter(r) || unicode.IsNumber(r) {
				noSymbol += string(r)
			}
		}

		_, ok := aliveWords[noSymbol]
		if !ok {
			solution += w
		} else {

			for _, r := range w {
				// is symbol
				if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
					solution += string(r)
				} else {
					solution += "_ "
				}
			}
		}
		solution += " "
	}

	return strings.Trim(solution, " ")
}
