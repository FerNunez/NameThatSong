package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePlaylistName(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid playlist name",
			input:       "My Awesome Playlist",
			expectError: false,
		},
		{
			name:        "valid playlist name with numbers",
			input:       "Playlist 123",
			expectError: false,
		},
		{
			name:        "valid playlist name with punctuation",
			input:       "My-Playlist_2024.mp3",
			expectError: false,
		},
		{
			name:        "valid playlist name with brackets",
			input:       "Playlist [Rock] (2024) {Best}",
			expectError: false,
		},
		{
			name:        "empty playlist name",
			input:       "",
			expectError: true,
			errorMsg:    "playlist name is required",
		},
		{
			name:        "whitespace only playlist name",
			input:       "   ",
			expectError: true,
			errorMsg:    "playlist name is required",
		},
		{
			name:        "playlist name too long",
			input:       string(make([]byte, 101)), // 101 characters
			expectError: true,
			errorMsg:    "playlist name is too long (max 100 characters)",
		},
		{
			name:        "playlist name with invalid characters",
			input:       "Playlist with émojis! 🎵",
			expectError: true,
			errorMsg:    "playlist name contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePlaylistName(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePlaylistDescription(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid description",
			input:       "This is a great playlist for working out",
			expectError: false,
		},
		{
			name:        "empty description (should be valid)",
			input:       "",
			expectError: false,
		},
		{
			name:        "description with punctuation",
			input:       "My playlist! Contains: rock, pop & jazz (2024)",
			expectError: false,
		},
		{
			name:        "description too long",
			input:       string(make([]byte, 501)), // 501 characters
			expectError: true,
			errorMsg:    "playlist description is too long (max 500 characters)",
		},
		{
			name:        "description with invalid characters",
			input:       "Description with émojis 🎵",
			expectError: true,
			errorMsg:    "playlist description contains invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePlaylistDescription(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateSpotifyID(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid spotify track ID",
			input:       "4iV5W9uYEdYUVa79Axb7Rh",
			expectError: false,
		},
		{
			name:        "valid spotify playlist ID",
			input:       "37i9dQZF1DXcBWIGoYBM5M",
			expectError: false,
		},
		{
			name:        "valid spotify ID with underscores and hyphens",
			input:       "4iV5W9uY_dYUVa79-xb7Rh",
			expectError: false,
		},
		{
			name:        "empty spotify ID",
			input:       "",
			expectError: true,
			errorMsg:    "spotify ID is required",
		},
		{
			name:        "spotify ID too short",
			input:       "short",
			expectError: true,
			errorMsg:    "invalid spotify ID length",
		},
		{
			name:        "spotify ID too long",
			input:       "this_is_a_very_long_spotify_id_that_exceeds_the_maximum_length",
			expectError: true,
			errorMsg:    "invalid spotify ID length",
		},
		{
			name:        "spotify ID with invalid characters",
			input:       "4iV5W9uYEdYUVa79Axb7Rh!",
			expectError: true,
			errorMsg:    "invalid spotify ID format",
		},
		{
			name:        "spotify ID with spaces",
			input:       "4iV5W9uY EdYUVa79Axb7Rh",
			expectError: true,
			errorMsg:    "invalid spotify ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSpotifyID(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePlaylistNameTrimming(t *testing.T) {
	// Test that whitespace is properly trimmed
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"leading whitespace", "  Valid Playlist", true},
		{"trailing whitespace", "Valid Playlist  ", true},
		{"both whitespace", "  Valid Playlist  ", true},
		{"only whitespace", "   ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePlaylistName(tt.input)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestValidateSpotifyIDTrimming(t *testing.T) {
	// Test that whitespace is properly trimmed
	validID := "4iV5W9uYEdYUVa79Axb7Rh"
	
	tests := []struct {
		name  string
		input string
	}{
		{"leading whitespace", "  " + validID},
		{"trailing whitespace", validID + "  "},
		{"both whitespace", "  " + validID + "  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSpotifyID(tt.input)
			assert.NoError(t, err)
		})
	}
}