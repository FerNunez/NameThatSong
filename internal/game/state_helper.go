package game

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

func CleanText(s string) string {
	cleanText := RemoveParenthesis(strings.ToLower(s))
	cleanText = RemoveAfterWord(cleanText, "-")
	cleanText = RemoveAfterWord(cleanText, "feat.")
	cleanText = RemoveAfterWord(cleanText, "feature")
	cleanText = RemoveSymbols(cleanText)
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

func RemoveSymbols(s string) string {

	// remove symbol
	noSymbol := ""
	for _, r := range strings.ToLower(s) {
		// remove symbol
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			noSymbol += string(r)
		} else {
			noSymbol += " "
		}
	}

	return noSymbol
}
func ProcessState(original string, aliveWords map[string]uint8) string {

	if len(aliveWords) == 0 {
		return original
	}

	original = RemoveSymbols(original)
	original = RemoveAccent(original)

	solution := ""
	words := strings.Split(original, " ")
	for _, w := range words {

		// remove symbol
		// noSymbol := ""
		// for _, r := range strings.ToLower(w) {
		// 	// remove symbol
		// 	if unicode.IsLetter(r) || unicode.IsNumber(r) {
		// 		noSymbol += string(r)
		// 	}
		// }
		//
		_, ok := aliveWords[w]
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
