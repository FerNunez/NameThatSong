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

-- Spotify playlists cache: For caching Spotify playlist metadata (separate from user playlists)
-- This is pure cache data - not owned by any user in our system
CREATE TABLE spotify_playlists_cache (
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
CREATE INDEX idx_album_artists_album ON spotify_album_artists(album_id);
CREATE INDEX idx_album_artists_artist ON spotify_album_artists(artist_id);
CREATE INDEX idx_track_artists_track ON spotify_track_artists(track_id);
CREATE INDEX idx_track_artists_artist ON spotify_track_artists(artist_id);
CREATE INDEX idx_track_artists_primary ON spotify_track_artists(track_id, is_primary);

-- Cache management indexes for TTL-based cleanup
CREATE INDEX idx_spotify_artists_cached_at ON spotify_artists(cached_at);
CREATE INDEX idx_spotify_albums_cached_at ON spotify_albums(cached_at);
CREATE INDEX idx_spotify_tracks_cached_at ON spotify_tracks(cached_at);
CREATE INDEX idx_spotify_playlists_cached_at ON spotify_playlists_cache(cached_at);

-- Search optimization indexes
CREATE INDEX idx_spotify_artists_name_trgm ON spotify_artists USING gin(name gin_trgm_ops);
CREATE INDEX idx_spotify_albums_name_trgm ON spotify_albums USING gin(name gin_trgm_ops);
CREATE INDEX idx_spotify_tracks_name_trgm ON spotify_tracks USING gin(name gin_trgm_ops);

-- +goose Down
DROP TABLE spotify_track_artists CASCADE;
DROP TABLE spotify_album_artists CASCADE;
DROP TABLE spotify_playlists_cache CASCADE;
DROP TABLE spotify_tracks CASCADE;
DROP TABLE spotify_albums CASCADE;
DROP TABLE spotify_artists CASCADE;