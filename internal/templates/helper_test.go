package templates

import (
	"testing"
)

func TestArtistsListString(t *testing.T) {
	// Empty list
	result := ArtistsListString(20, []string{})
	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}

	// Single short name
	result = ArtistsListString(20, []string{"Bob"})
	if result != "Bob" {
		t.Errorf("Expected 'Bob', got %q", result)
	}

	// Limit long name
	result = ArtistsListString(20, []string{"VeryyLongArtistName"})
	if result != "VeryyLongArtistName" {
		t.Errorf("Expected 'VeryyLongArtistName', got %q", result)
	}
	// Long name
	result = ArtistsListString(17, []string{"VeryLongArtistName"})
	if result != "Ver..ame" {
		t.Errorf("Expected 'Ver..ame', got %q", result)
	}

	// Two names under limit
	result = ArtistsListString(20, []string{"Bob", "Alice"})
	if result != "Bob, Alice" {
		t.Errorf("Expected 'Bob, Alice', got %q", result)
	}

	// Mixed short and long names
	result = ArtistsListString(10, []string{"LongName", "AnotherLong"})
	if result != "LongName, Ano..ong" {
		t.Errorf("Expected 'LongName, Ano..ong', got %q", result)
	}
}
