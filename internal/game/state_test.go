package game_test

import (
	"testing"

	"github.com/FerNunez/NameThatSong/internal/game"
)

func TestRemoveAccent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single accented character",
			input:    "é",
			expected: "e",
		},
		{"word with é", "café", "cafe"},
		{"word with à", "à la carte", "a la carte"},
		{"word with ü", "über", "uber"},
		{"word with ñ", "jalapeño", "jalapeno"},
		{"word with ç", "façade", "facade"},
		{"word with ö", "doppelgänger", "doppelganger"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := game.RemoveAccent(tt.input)
			if result != tt.expected {
				t.Errorf("RemoveAccent(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRemoveParenthesis(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no parenthesis", "Hello world", "Hello world"},
		{"simple parentheses", "Hello (world)", "Hello "},
		{"parenthesis without closing", "Hello (world", "Hello (world"},
		{"multiple parentheses", "Remove (this) text (and) (keep)", "Remove "},
		{"parenthesis at end", "Test (case)", "Test "},
		{"parenthesis inside text", "Keep (this) part (only)", "Keep "},
		{"only parentheses", "(only)", ""},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := game.RemoveParenthesis(tt.input)
			if got != tt.expected {
				t.Errorf("RemoveParenthesis(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
func TestRemoveWord(t *testing.T) {
	tests := []struct {
		name       string
		input_text string
		input_word string
		expected   string
	}{
		{"remove feat.", "Hello feat. tu mama", "feat.", "Hello "},
		{"remove feature", "Hello feature tu mama", "feature", "Hello "},
		{"remove featurastico", "Hello featurastico tu mama", "feat.", "Hello featurastico tu mama"},
		{"remove - ", "Hello - tu mama", "-", "Hello "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := game.RemoveAfterWord(tt.input_text, tt.input_word)
			if got != tt.expected {
				t.Errorf("RemoveAfterWord(%q) = %q & %q, want %q", tt.input_text, tt.input_word, got, tt.expected)
			}
		})
	}
}
func TestProcessState(t *testing.T) {
	// Define the test cases
	testCases := []struct {
		original   string
		aliveWords map[string]uint8
		expected   string
	}{
		{
			"Hello world!",
			map[string]uint8{"hello": 1},
			"_ _ _ _ _  world!", // Replace Hello with underscore, keep 'world!' as it is
		},
		{
			"Go programming is fun!",
			map[string]uint8{"go": 1, "fun": 1},
			"_ _  programming is _ _ _ !", // Replace 'Go' and 'fun' with underscores
		},
		{
			"Hello, how are you?",
			map[string]uint8{"how": 1, "you": 1},
			"Hello, _ _ _  are _ _ _ ?", // Replace 'how' and 'you' with underscores
		},
		{
			"Test123 is here!",
			map[string]uint8{"test123": 1},
			"_ _ _ _ _ _ _  is here!", // Replace 'Test123' with underscore, keep 'is here!' intact
		},
		{
			"Start testing symbols: @#%$!",
			map[string]uint8{"testing": 1},
			"Start _ _ _ _ _ _ _  symbols: @#%$!", // Replace 'testing' with underscores
		},
		{
			"Test symbols and spaces.",
			map[string]uint8{"symbols": 1},
			"Test _ _ _ _ _ _ _  and spaces.", // Replace 'symbols' with underscores
		},
	}

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.original, func(t *testing.T) {
			result := game.ProcessState(tc.original, tc.aliveWords)
			if result != tc.expected {
				t.Errorf("FAIL: Original: '%s', Expected: '%s', Got: '%s'", tc.original, tc.expected, result)
			}
		})
	}
}
