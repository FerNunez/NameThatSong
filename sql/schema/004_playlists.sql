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

-- Playlist songs table
CREATE TABLE playlist_songs (
    id UUID PRIMARY KEY,
    playlist_id UUID NOT NULL REFERENCES local_playlists(id) ON DELETE CASCADE,
    spotify_track_id TEXT NOT NULL,
    spotify_album_id TEXT NOT NULL,
    spotify_artist_id TEXT NOT NULL,       -- list of artists ID
    track_name VARCHAR NOT NULL,
    album_name VARCHAR NOT NULL,
    artist_name TEXT NOT NULL,     -- list of artists names
    position INT NOT NULL,
    duration_ms INT NOT NULL,
    added_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(playlist_id, spotify_track_id)
);

-- Indexes for performance
CREATE INDEX idx_playlists_user_id ON local_playlists(user_id);
CREATE INDEX idx_playlists_spotify_id ON local_playlists(spotify_playlist_id);
CREATE INDEX idx_playlist_songs_playlist_id ON playlist_songs(playlist_id);
CREATE INDEX idx_playlist_songs_position ON playlist_songs(playlist_id, position);
CREATE INDEX idx_playlist_songs_spotify_track ON playlist_songs(spotify_track_id);

-- +goose Down
DROP TABLE playlist_songs CASCADE;
DROP TABLE local_playlists CASCADE;
