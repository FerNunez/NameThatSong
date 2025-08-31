-- name: UpsertSong :one
INSERT INTO songs (spotify_track_id,
    spotify_album_id,
    spotify_artist_id,
    track_name,
    album_name,
    artist_name,
    album_image_url,
    artist_image_url,
    duration_ms,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
ON CONFLICT (spotify_track_id) DO UPDATE SET
    spotify_album_id = EXCLUDED.spotify_album_id,
    spotify_artist_id = EXCLUDED.spotify_artist_id,
    track_name = EXCLUDED.track_name,
    album_name = EXCLUDED.album_name,
    artist_name = EXCLUDED.artist_name,
    album_image_url = EXCLUDED.album_image_url,
    artist_image_url = EXCLUDED.artist_image_url,
    duration_ms = EXCLUDED.duration_ms,
    created_at = EXCLUDED.created_at,
    updated_at = NOW()
RETURNING *;

-- name: GetSongBySpotifyID :one
SELECT * FROM songs WHERE spotify_track_id = $1;
