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
