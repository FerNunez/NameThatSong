package game_test

import (
	"testing"
)

func TestCleanText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"é", "e"},
		{"Crème brûlée", "Creme brulee"},
		{"naïve", "naive"},
		{"façade", "facade"},
		{"élève", "eleve"},
		{"piñata", "pinata"},
		{"über", "uber"},
		{"Ångström", "Angstrom"},
	}

	for _, tt := range tests {
		result := CleanText(tt.input)
		if result != tt.expected {
			t.Errorf("removeAccents(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}
