package game

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type GuessState struct {
	title *TitleGuessState
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

		if wNormalized == "" {
			continue
		}
		wordsCounts[wNormalized] += 1
	}

	return &TitleGuessState{
		RealTitle:       titleName,
		TitleAliveWords: wordsCounts,
	}
}

func (g *GuessState) Guess(text string) (string, bool) {

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

	return ProcessState(t.RealTitle, t.TitleAliveWords)
}

func CleanText(s string) string {
	// Split by spaces and symbols
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || (!unicode.IsLetter(r) && !unicode.IsNumber(r))
	})

	// Clean each part
	cleanedParts := make([]string, 0, len(parts))
	for _, part := range parts {
		// Normalize to decomposed form (NFD)
		t := norm.NFD.String(part)
		result := make([]rune, 0, len(t))

		for _, r := range t {
			// Skip diacritical marks
			if unicode.Is(unicode.Mn, r) {
				continue
			}
			result = append(result, r)
		}

		if len(result) > 0 {
			cleanedParts = append(cleanedParts, string(result))
		}
	}

	return strings.Join(cleanedParts, " ")
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
