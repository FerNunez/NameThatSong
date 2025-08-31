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

