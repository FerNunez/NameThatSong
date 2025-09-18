package utils

import (
	"strings"
	"testing"
)

func TestNormalizeSearchQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic normalization",
			input:    "The Beatles",
			expected: "the beatles",
		},
		{
			name:     "whitespace trimming",
			input:    "  The Beatles  ",
			expected: "the beatles",
		},
		{
			name:     "multiple spaces",
			input:    "The    Beatles",
			expected: "the beatles",
		},
		{
			name:     "mixed whitespace",
			input:    "The\t\nBeatles\r",
			expected: "the beatles",
		},
		{
			name:     "punctuation removal",
			input:    "AC/DC - Thunderstruck!!!",
			expected: "ac dc thunderstruck",
		},
		{
			name:     "remove apostrophes",
			input:    "Don't Stop Me Now",
			expected: "don t stop me now",
		},
		{
			name:     "remove hyphens",
			input:    "Twenty-One Pilots",
			expected: "twenty one pilots",
		},
		{
			name:     "unicode normalization",
			input:    "Café Tacvba",
			expected: "cafe tacvba",
		},
		{
			name:     "accented characters",
			input:    "Beyoncé",
			expected: "beyonce",
		},
		{
			name:     "complex unicode",
			input:    "Sigur Rós",
			expected: "sigur ros",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
		},
		{
			name:     "punctuation only",
			input:    "!!!",
			expected: "",
		},
		{
			name:     "mixed case with numbers",
			input:    "Blink-182",
			expected: "blink 182",
		},
		{
			name:     "special characters",
			input:    "Guns N' Roses",
			expected: "guns n roses",
		},
		{
			name:     "complex query",
			input:    "  The Rolling Stones - Paint It Black (Live)  ",
			expected: "the rolling stones paint it black live",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeSearchQuery(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeSearchQuery(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeForCacheKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic sanitization",
			input:    "the beatles",
			expected: "the+beatles",
		},
		{
			name:     "special characters",
			input:    "ac/dc thunderstruck",
			expected: "ac%2Fdc+thunderstruck",
		},
		{
			name:     "removed apostrophe encoding",
			input:    "don t stop me now",
			expected: "don+t+stop+me+now",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "spaces to plus",
			input:    "hello world",
			expected: "hello+world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeForCacheKey(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeForCacheKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeForCacheKey_LongQuery(t *testing.T) {
	// Create a query longer than MaxCacheKeyLength
	longQuery := strings.Repeat("very long query ", 20) // Should be > 250 chars

	result := SanitizeForCacheKey(longQuery)

	// Should be hashed and start with "hash_"
	if !strings.HasPrefix(result, "hash_") {
		t.Errorf("Expected long query to be hashed with 'hash_' prefix, got: %s", result)
	}

	// Should be a reasonable length (32 hex chars + "hash_" = 37 chars)
	if len(result) != 37 {
		t.Errorf("Expected hashed query to be 37 characters, got: %d", len(result))
	}
}

func TestNormalizeAndSanitizeQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "complete normalization and sanitization",
			input:    "  The Beatles - Hey Jude!  ",
			expected: "the+beatles+hey+jude",
		},
		{
			name:     "unicode and special chars",
			input:    "Café Tacvba / Rock",
			expected: "cafe+tacvba+rock",
		},
		{
			name:     "apostrophes removed",
			input:    "Don't Stop Believin'",
			expected: "don+t+stop+believin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeAndSanitizeQuery(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeAndSanitizeQuery(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsValidSearchQuery(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid query",
			input:    "The Beatles",
			expected: true,
		},
		{
			name:     "valid with numbers",
			input:    "Blink-182",
			expected: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: false,
		},
		{
			name:     "punctuation only",
			input:    "!!!???",
			expected: false,
		},
		{
			name:     "valid after normalization",
			input:    "  Beatles!!!  ",
			expected: true,
		},
		{
			name:     "single character",
			input:    "a",
			expected: true,
		},
		{
			name:     "unicode characters",
			input:    "Café",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidSearchQuery(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidSearchQuery(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRemoveAccents(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "cafe with accent",
			input:    "Café",
			expected: "Cafe",
		},
		{
			name:     "beyonce with accent",
			input:    "Beyoncé",
			expected: "Beyonce",
		},
		{
			name:     "sigur ros",
			input:    "Sigur Rós",
			expected: "Sigur Ros",
		},
		{
			name:     "no accents",
			input:    "Beatles",
			expected: "Beatles",
		},
		{
			name:     "multiple accents",
			input:    "Mötörhead",
			expected: "Motorhead",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeAccents(tt.input)
			if result != tt.expected {
				t.Errorf("removeAccents(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkNormalizeSearchQuery(b *testing.B) {
	query := "  The Rolling Stones - Paint It Black (Live)!!!  "
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		NormalizeSearchQuery(query)
	}
}

func BenchmarkSanitizeForCacheKey(b *testing.B) {
	query := "the rolling stones paint it black live"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		SanitizeForCacheKey(query)
	}
}

func BenchmarkNormalizeAndSanitizeQuery(b *testing.B) {
	query := "  The Rolling Stones - Paint It Black (Live)!!!  "
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		NormalizeAndSanitizeQuery(query)
	}
}
