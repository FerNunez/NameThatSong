-- +goose Up
CREATE TABLE users(
  id UUID PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  hashed_password TEXT DEFAULT 'unset' NOT NULL,
  email_verified BOOLEAN DEFAULT FALSE,
  display_name VARCHAR(100) NOT NULL,
  avatar_url TEXT,
  spotify_connected BOOLEAN DEFAULT FALSE,
  last_login_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

-- Create email verification token table
CREATE TABLE email_verification_tokens(
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token VARCHAR(255) NOT NULL UNIQUE,
  expires_at TIMESTAMP NOT NULL,
  used_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_email_verification_tokens_token ON email_verification_tokens(token);
CREATE INDEX idx_email_verification_tokens_user_id ON email_verification_tokens(user_id);

-- Create password reset tokens table
CREATE TABLE password_reset_tokens (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token VARCHAR(255) NOT NULL UNIQUE,
  expires_at TIMESTAMP NOT NULL,
  used_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL
);

-- Create index for fast token lookups
CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token);
CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);


-- Create user's sessions
CREATE TABLE user_sessions(
  id TEXT PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at TIMESTAMP NOT NULL,
  revoked_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

-- Create index for fast lookups
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_expired_at ON user_sessions(expires_at);

-- +goose Down
DROP TABLE users CASCADE;
DROP TABLE password_reset_tokens CASCADE;
DROP TABLE email_verification_tokens CASCADE;
DROP TABLE user_sessions;
-- +goose Up
CREATE TABLE spotify_tokens(
  user_id UUID PRIMARY KEY,
  refresh_token TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  access_token TEXT NOT NULL,
  token_type TEXT NOT NULL,
  scope TEXT NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE spotify_tokens;
-- +goose Up
-- Create game sessions table for tracking game history
CREATE TABLE game_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    game_mode VARCHAR(50) NOT NULL,
    total_questions INTEGER NOT NULL,
    correct_answers INTEGER NOT NULL,
    total_score INTEGER NOT NULL,
    duration_seconds INTEGER NOT NULL,
    tracks_played TEXT[], -- Array of Spotify track IDs
    completed_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for game sessions
CREATE INDEX idx_game_sessions_user_id ON game_sessions(user_id);
CREATE INDEX idx_game_sessions_completed_at ON game_sessions(completed_at);
CREATE INDEX idx_game_sessions_game_mode ON game_sessions(game_mode);

-- +goose Down
-- Drop tables in reverse order
DROP TABLE game_sessions CASCADE;
-- +goose Up
-- Spotify data caching and normalization schema
-- This implements the database layer of the 3-tier strategy: Cache -> Database -> API

-- Artists table: Core artist information from Spotify
CREATE TABLE spotify_artists (
    id TEXT PRIMARY KEY, -- Spotify artist ID
    name VARCHAR(255) NOT NULL,
    image_url TEXT,
    popularity INTEGER DEFAULT 0,
    followers_total INTEGER DEFAULT 0,
    genres TEXT[], -- PostgreSQL array for multiple genres
    cached_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Albums table: Album metadata with foreign key to primary artist
CREATE TABLE spotify_albums (
    id TEXT PRIMARY KEY, -- Spotify album ID
    name VARCHAR(255) NOT NULL,
    album_type VARCHAR(50) NOT NULL, -- album, single, compilation
    release_date DATE,
    release_date_precision VARCHAR(10), -- year, month, day
    total_tracks INTEGER DEFAULT 0,
    image_url TEXT,
    label VARCHAR(255),
    popularity INTEGER DEFAULT 0,
    cached_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Tracks table: Individual song data
CREATE TABLE spotify_tracks (
    id TEXT PRIMARY KEY, -- Spotify track ID
    name VARCHAR(255) NOT NULL,
    album_id TEXT REFERENCES spotify_albums(id) ON DELETE SET NULL,
    duration_ms INTEGER NOT NULL,
    disc_number INTEGER DEFAULT 1,
    track_number INTEGER DEFAULT 1,
    popularity INTEGER DEFAULT 0,
    explicit BOOLEAN DEFAULT FALSE,
    preview_url TEXT,
    is_local BOOLEAN DEFAULT FALSE,
    cached_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Spotify playlists: For caching Spotify playlist metadata (separate from user playlists)
-- This is pure cache data - not owned by any user in our system
CREATE TABLE spotify_playlists (
    id TEXT PRIMARY KEY, -- Spotify playlist ID
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id TEXT NOT NULL, -- Spotify user ID (not our user ID)
    owner_display_name VARCHAR(255),
    public BOOLEAN DEFAULT TRUE,
    collaborative BOOLEAN DEFAULT FALSE,
    followers_total INTEGER DEFAULT 0,
    total_tracks INTEGER DEFAULT 0,
    image_url TEXT,
    cached_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- RELATIONSHIPS
-- Many-to-many: Albums can have multiple artists (collaborations, compilations)
CREATE TABLE spotify_album_artists (
    album_id TEXT REFERENCES spotify_albums(id) ON DELETE CASCADE,
    artist_id TEXT REFERENCES spotify_artists(id) ON DELETE CASCADE,
    PRIMARY KEY (album_id, artist_id)
);

-- Many-to-many: Tracks can have multiple artists (features, collaborations)
CREATE TABLE spotify_track_artists (
    track_id TEXT REFERENCES spotify_tracks(id) ON DELETE CASCADE,
    artist_id TEXT REFERENCES spotify_artists(id) ON DELETE CASCADE,
    is_primary BOOLEAN DEFAULT FALSE, -- To identify main artist for display
    PRIMARY KEY (track_id, artist_id)
);

-- -- One-to-many: Alums can have multiple tracks
CREATE TABLE spotify_album_tracks (
    album_id TEXT REFERENCES spotify_albums(id) ON DELETE CASCADE,
    track_id TEXT REFERENCES spotify_tracks(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    PRIMARY KEY (album_id, track_id)
);


-- One-to-many: One spotify has many tracks
CREATE TABLE spotify_playlist_tracks(
    playlist_id TEXT REFERENCES spotify_playlists(id) ON DELETE CASCADE,
    track_id TEXT REFERENCES spotify_tracks(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (playlist_id, track_id),
    UNIQUE(playlist_id, position)
);


-- Performance indexes for common queries
CREATE INDEX idx_spotify_artists_name ON spotify_artists(name);
CREATE INDEX idx_spotify_artists_popularity ON spotify_artists(popularity DESC);
CREATE INDEX idx_spotify_albums_name ON spotify_albums(name);
CREATE INDEX idx_spotify_albums_release_date ON spotify_albums(release_date DESC);
CREATE INDEX idx_spotify_tracks_name ON spotify_tracks(name);
CREATE INDEX idx_spotify_tracks_album ON spotify_tracks(album_id);
CREATE INDEX idx_spotify_tracks_duration ON spotify_tracks(duration_ms);
CREATE INDEX idx_spotify_tracks_popularity ON spotify_tracks(popularity DESC);

-- Relationship indexes for efficient JOINs
CREATE INDEX idx_album_artists_artist ON spotify_album_artists(artist_id);
CREATE INDEX idx_track_artists_artist ON spotify_track_artists(artist_id);
CREATE INDEX idx_track_artists_primary ON spotify_track_artists(track_id, is_primary);
CREATE INDEX idx_album_tracks_track ON spotify_album_tracks(track_id);
CREATE INDEX idx_spotify_playlist_trackst_position ON spotify_playlist_tracks(playlist_id, position);

-- Cache management indexes for TTL-based cleanup
CREATE INDEX idx_spotify_artists_cached_at ON spotify_artists(cached_at);
CREATE INDEX idx_spotify_albums_cached_at ON spotify_albums(cached_at);
CREATE INDEX idx_spotify_tracks_cached_at ON spotify_tracks(cached_at);
CREATE INDEX idx_spotify_playlists_cached_at ON spotify_playlists(cached_at);

-- +goose Down
DROP TABLE spotify_track_artists CASCADE;
DROP TABLE spotify_album_artists CASCADE;
DROP TABLE spotify_playlists CASCADE;
DROP TABLE spotify_tracks CASCADE;
DROP TABLE spotify_albums CASCADE;
DROP TABLE spotify_artists CASCADE;
DROP TABLE spotify_playlist_tracks CASCADE;
DROP TABLE spotify_album_tracks CASCADE;
-- +goose Up
CREATE TABLE songs (
    spotify_track_id TEXT PRIMARY KEY,
    spotify_album_id TEXT NOT NULL,
    spotify_artist_id TEXT NOT NULL,
    track_name VARCHAR NOT NULL,
    album_name VARCHAR NOT NULL,
    artist_name TEXT NOT NULL,
    artist_image_url TEXT NOT NULL,
    album_image_url VARCHAR NOT NULL,
    duration_ms INT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Indexes for performance
CREATE INDEX idx_songs_spotify_album_id ON songs(spotify_album_id);
CREATE INDEX idx_songs_spotify_artist_id ON songs(spotify_artist_id);


-- +goose Down
DROP TABLE songs CASCADE;

-- +goose Up
CREATE TABLE local_playlists (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    spotify_playlist_id TEXT NULL, -- NULL for local playlists
    is_public BOOLEAN DEFAULT FALSE NOT NULL,
    sync_with_spotify BOOLEAN DEFAULT FALSE NOT NULL,
    last_synced_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- One-to-many: One playlist has many tracks
CREATE TABLE local_playlist_tracks(
    playlist_id UUID REFERENCES local_playlists(id) ON DELETE CASCADE,
    spotify_track_id TEXT REFERENCES songs(spotify_track_id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (playlist_id, spotify_track_id),
    UNIQUE(playlist_id, position)
);

-- Indexes for performance
CREATE INDEX idx_playlists_user_id ON local_playlists(user_id);
CREATE INDEX idx_playlists_spotify_id ON local_playlists(spotify_playlist_id);
CREATE INDEX idx_local_playlist_tracks_position ON local_playlist_tracks(playlist_id, position);
CREATE INDEX idx_local_playlist_tracks_spotify_track ON local_playlist_tracks(spotify_track_id);

-- +goose Down
DROP TABLE local_playlist_tracks CASCADE;
DROP TABLE local_playlists CASCADE;
