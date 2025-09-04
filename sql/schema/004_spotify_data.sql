-- +goose Up
-- Tracks table: Individual song data
CREATE TABLE spotify_tracks (
    id TEXT PRIMARY KEY, -- Spotify track ID
    name VARCHAR(255) NOT NULL,
    duration_ms INTEGER NOT NULL,
    disc_number INTEGER DEFAULT 1,
    track_number INTEGER DEFAULT 1,
    popularity INTEGER DEFAULT 0,
    explicit BOOLEAN DEFAULT FALSE,
    is_local BOOLEAN DEFAULT FALSE,
    album_id TEXT NOT NULL,
    artist_ids TEXT[] NOT NULL,
    cached_at TIMESTAMP NOT NULL,
);


-- Albums table: Album metadata with foreign key to primary artist
CREATE TABLE spotify_albums (
    id TEXT PRIMARY KEY, -- Spotify album ID
    name VARCHAR(255) NOT NULL,
    album_type VARCHAR(50) NOT NULL, -- album, single, compilation
    release_date DATE,
    release_date_precision VARCHAR(10) NOT NULL, -- year, month, day
    total_tracks INTEGER NOT NULL DEFAULT 0,
    image_url TEXT,
    label VARCHAR(255),
    popularity INTEGER NOT NULL DEFAULT 0 ,
    artist_ids TEXT[] NOT NULL,
    track_ids TEXT[] NOT NULL,
    cached_at TIMESTAMP NOT NULL
);

-- Artists table: Core artist information from Spotify
CREATE TABLE spotify_artists (
    id TEXT PRIMARY KEY, -- Spotify artist ID
    name VARCHAR(255) NOT NULL,
    image_url TEXT,
    popularity INTEGER DEFAULT 0,
    followers_total INTEGER NOT NULL DEFAULT 0,
    genres TEXT[], -- PostgreSQL array for multiple genres
    cached_at TIMESTAMP NOT NULL
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
    track_ids TEXT[] NOT NULL,
    cached_at TIMESTAMP NOT NULL
);

-- Cache management indexes for TTL-based cleanup
CREATE INDEX idx_spotify_artists_cached_at ON spotify_artists(cached_at);
CREATE INDEX idx_spotify_albums_cached_at ON spotify_albums(cached_at);
CREATE INDEX idx_spotify_tracks_cached_at ON spotify_tracks(cached_at);
CREATE INDEX idx_spotify_playlists_cached_at ON spotify_playlists(cached_at);

-- +goose Down
DROP TABLE spotify_playlists CASCADE;
DROP TABLE spotify_artists CASCADE;
DROP TABLE spotify_albums CASCADE;
DROP TABLE spotify_tracks CASCADE;
