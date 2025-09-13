-- Playlist operations
-- TODO: To add UPSERT LocalPlaylist
-- name: CreatePlaylist :one
INSERT INTO local_playlists (id, user_id, name, description, image_url, spotify_playlist_id, snapshot_id, is_public, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
RETURNING *;

-- name: GetPlaylistByID :one
SELECT * FROM local_playlists WHERE id = $1;

-- name: GetPlaylistsByUserID :many
SELECT * FROM local_playlists WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetPlaylistByUserIDAndID :one
SELECT * FROM local_playlists WHERE id = $1 AND user_id = $2;

-- name: GetPlaylistBySpotifyIDAndUserID :one
SELECT * FROM local_playlists WHERE spotify_playlist_id = $1 AND user_id = $2 ORDER BY name DESC;

-- name: UpdatePlaylist :exec
UPDATE local_playlists 
SET name = $2, description = $3, image_url=$4, is_public = $5, updated_at = NOW()
WHERE id = $1 AND user_id = $6;

-- name: DeletePlaylist :exec
DELETE FROM local_playlists WHERE id = $1 AND user_id = $2;

-- Playlist songs operations
-- name: AddSongToPlaylist :one
INSERT INTO local_playlist_tracks (playlist_id, spotify_track_id, position, updated_at)
VALUES ($1, $2, $3, NOW())
RETURNING *;


-- name: GetPlaylistSongs :many
SELECT * FROM local_playlist_tracks WHERE playlist_id = $1 ORDER BY position;

-- name: GetPlaylistSongsWithTrackData :many
SELECT 
    lpt.spotify_track_id,
    lpt.position,
    lpt.updated_at,
    st.name AS track_name,
    st.duration_ms AS track_duration_ms,
    st.album_id AS track_album_id,
    st.artist_ids AS track_artist_ids
FROM local_playlist_tracks lpt
JOIN spotify_tracks st ON lpt.spotify_track_id = st.id
WHERE lpt.playlist_id = $1 
ORDER BY lpt.position;

-- name: GetPlaylistSongByID :one
SELECT * FROM local_playlist_tracks WHERE playlist_id = $1 AND spotify_track_id = $2;

-- name: RemoveSongFromPlaylist :exec
DELETE FROM local_playlist_tracks WHERE playlist_id = $1 AND spotify_track_id = $2;

-- name: UpdateSongPosition :exec
UPDATE local_playlist_tracks SET position = $3 WHERE spotify_track_id = $2 AND playlist_id = $1;

-- name: GetMaxSongPosition :one
SELECT COALESCE(MAX(position), 0)::INT AS last_position FROM local_playlist_tracks WHERE playlist_id = $1;

-- name: ClearPlaylistSongs :exec
DELETE FROM local_playlist_tracks WHERE playlist_id = $1;
