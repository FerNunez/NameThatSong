package utils

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	// MaxCacheKeyLength defines the maximum length for a cache key before hashing
	MaxCacheKeyLength = 250
)

var (
	// multipleSpacesRegex matches multiple consecutive whitespace characters
	multipleSpacesRegex = regexp.MustCompile(`\s+`)

	// punctuationRegex matches ALL punctuation and symbols (keeps only letters, numbers, spaces)
	punctuationRegex = regexp.MustCompile(`[^\w\s]+`)
)

// NormalizeSearchQuery normalizes a search query for consistent caching
// This function:
// 1. Trims whitespace
// 2. Converts to lowercase
// 3. Normalizes Unicode characters (removes accents)
// 4. Removes ALL punctuation and symbols (keeps only letters, numbers, spaces)
// 5. Collapses multiple spaces to single space
func NormalizeSearchQuery(query string) string {
	if query == "" {
		return ""
	}

	// Step 1: Trim whitespace
	normalized := strings.TrimSpace(query)
	if normalized == "" {
		return ""
	}

	// Step 2: Convert to lowercase
	normalized = strings.ToLower(normalized)

	// Step 3: Normalize Unicode characters (remove accents)
	normalized = removeAccents(normalized)

	// Step 4: Remove ALL punctuation and symbols (keep only letters, numbers, spaces)
	normalized = punctuationRegex.ReplaceAllString(normalized, " ")

	// Step 5: Collapse multiple spaces to single space
	normalized = multipleSpacesRegex.ReplaceAllString(normalized, " ")

	// Final trim after processing
	normalized = strings.TrimSpace(normalized)

	return normalized
}

// SanitizeForCacheKey makes a normalized query safe for use as a Redis cache key
// This function:
// 1. URL encodes the query for Redis key safety
// 2. Hashes very long queries to prevent oversized keys
func SanitizeForCacheKey(normalizedQuery string) string {
	if normalizedQuery == "" {
		return ""
	}

	// URL encode for Redis key safety
	encoded := url.QueryEscape(normalizedQuery)

	// If the encoded query is too long, hash it to a fixed length
	if len(encoded) > MaxCacheKeyLength {
		hash := sha256.Sum256([]byte(normalizedQuery))
		return fmt.Sprintf("hash_%x", hash[:16]) // Use first 16 bytes (32 hex chars)
	}

	return encoded
}

// NormalizeAndSanitizeQuery combines normalization and sanitization for cache keys
func NormalizeAndSanitizeQuery(query string) string {
	normalized := NormalizeSearchQuery(query)
	return SanitizeForCacheKey(normalized)
}

// removeAccents removes accent marks from Unicode characters
// Example: "Café" -> "Cafe", "Beyoncé" -> "Beyonce"
func removeAccents(s string) string {
	// Create a transformer that decomposes Unicode characters and removes combining marks
	transformer := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(transformer, s)
	return result
}

// IsValidSearchQuery checks if a query is valid for searching after normalization
func IsValidSearchQuery(query string) bool {
	normalized := NormalizeSearchQuery(query)

	// Consider invalid if empty after normalization
	if normalized == "" {
		return false
	}

	// Consider invalid if only whitespace or punctuation
	if strings.TrimSpace(punctuationRegex.ReplaceAllString(normalized, "")) == "" {
		return false
	}

	return true
}
