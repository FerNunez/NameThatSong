package validation

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode"
)

// Email validation
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	// Trim whitespace
	email = strings.TrimSpace(email)

	// Check if empty after trimming
	if email == "" {
		return fmt.Errorf("email is required")
	}

	// Check length after trimming
	if len(email) > 254 {
		return fmt.Errorf("email is too long (max 254 characters)")
	}

	// Use Go's built-in email validation
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// Password validation
func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}

	// Check minimum length
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Check maximum length to prevent DoS
	if len(password) > 128 {
		return fmt.Errorf("password is too long (max 128 characters)")
	}

	// Check for at least one uppercase letter
	hasUpper := false
	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	// Check for at least one lowercase letter
	hasLower := false
	for _, char := range password {
		if unicode.IsLower(char) {
			hasLower = true
			break
		}
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	// Check for at least one digit
	hasDigit := false
	for _, char := range password {
		if unicode.IsDigit(char) {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}

	return nil
}

// Display name validation
func ValidateDisplayName(displayName string) error {
	// Trim whitespace
	displayName = strings.TrimSpace(displayName)

	// Check if empty after trimming
	if displayName == "" {
		return fmt.Errorf("display name is required")
	}

	// Check length after trimming
	if len(displayName) > 100 {
		return fmt.Errorf("display name is too long (max 100 characters)")
	}

	// Check for valid characters (alphanumeric, spaces, basic punctuation)
	validName := regexp.MustCompile(`^[a-zA-Z0-9\s\-_\.]+$`)
	if !validName.MatchString(displayName) {
		return fmt.Errorf("display name contains invalid characters")
	}

	return nil
}

// Avatar URL validation
func ValidateAvatarURL(avatarURL string) error {
	// Trim whitespace
	avatarURL = strings.TrimSpace(avatarURL)

	// Check if empty after trimming
	if avatarURL == "" {
		return fmt.Errorf("avatar URL is required")
	}

	// Check length after trimming
	if len(avatarURL) > 500 {
		return fmt.Errorf("avatar URL is too long (max 500 characters)")
	}

	// Basic URL validation
	if !strings.HasPrefix(avatarURL, "http://") && !strings.HasPrefix(avatarURL, "https://") {
		return fmt.Errorf("avatar URL must start with http:// or https://")
	}

	return nil
}

// Token validation
func ValidateToken(token string) error {
	// Trim whitespace
	token = strings.TrimSpace(token)

	// Check if empty after trimming
	if token == "" {
		return fmt.Errorf("token is required")
	}

	// Check length constraints after trimming
	if len(token) < 16 {
		return fmt.Errorf("token is too short")
	}

	if len(token) > 128 {
		return fmt.Errorf("token is too long")
	}

	// Check for valid characters (alphanumeric)
	validToken := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !validToken.MatchString(token) {
		return fmt.Errorf("token contains invalid characters")
	}

	return nil
}

// Playlist name validation
func ValidatePlaylistName(name string) error {
	// Trim whitespace
	name = strings.TrimSpace(name)

	// Check if empty after trimming
	if name == "" {
		return fmt.Errorf("playlist name is required")
	}

	// Check length after trimming
	if len(name) > 100 {
		return fmt.Errorf("playlist name is too long (max 100 characters)")
	}

	// Check for valid characters (alphanumeric, spaces, basic punctuation)
	validName := regexp.MustCompile(`^[a-zA-Z0-9\s\-_\.\(\)\[\]\{\}]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("playlist name contains invalid characters")
	}

	return nil
}

// Playlist description validation
func ValidatePlaylistDescription(description string) error {
	// Description is optional, so empty is valid
	if description == "" {
		return nil
	}

	// Trim whitespace
	description = strings.TrimSpace(description)

	// Check length after trimming
	if len(description) > 500 {
		return fmt.Errorf("playlist description is too long (max 500 characters)")
	}

	// Allow more flexible characters for descriptions
	validDescription := regexp.MustCompile(`^[a-zA-Z0-9\s\-_\.\(\)\[\]\{\}!@#$%^&*+=:;",\?]+$`)
	if !validDescription.MatchString(description) {
		return fmt.Errorf("playlist description contains invalid characters")
	}

	return nil
}

// Spotify ID validation (for playlists and tracks)
func ValidateSpotifyID(spotifyID string) error {
	// Trim whitespace
	spotifyID = strings.TrimSpace(spotifyID)

	// Check if empty after trimming
	if spotifyID == "" {
		return fmt.Errorf("spotify ID is required")
	}

	// Spotify IDs are base64-like strings, typically 22 characters
	if len(spotifyID) < 15 || len(spotifyID) > 30 {
		return fmt.Errorf("invalid spotify ID length")
	}

	// Check for valid Spotify ID characters (base64 characters)
	validSpotifyID := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validSpotifyID.MatchString(spotifyID) {
		return fmt.Errorf("invalid spotify ID format")
	}

	return nil
}
