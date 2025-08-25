-- Playlist operations
-- name: CreatePlaylist :one
INSERT INTO local_playlists (id, user_id, name, description, spotify_playlist_id, is_public, sync_with_spotify, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
RETURNING *;

-- name: GetPlaylistByID :one
SELECT * FROM local_playlists WHERE id = $1;

-- name: GetPlaylistsByUserID :many
SELECT * FROM local_playlists WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetPlaylistByUserIDAndID :one
SELECT * FROM local_playlists WHERE id = $1 AND user_id = $2;

-- name: UpdatePlaylist :exec
UPDATE local_playlists 
SET name = $2, description = $3, is_public = $4, sync_with_spotify = $5, updated_at = NOW()
WHERE id = $1 AND user_id = $6;

-- name: UpdatePlaylistSyncTime :exec
UPDATE local_playlists 
SET last_synced_at = NOW(), updated_at = NOW()
WHERE id = $1 AND user_id = $2;

-- name: DeletePlaylist :exec
DELETE FROM local_playlists WHERE id = $1 AND user_id = $2;

-- Playlist songs operations
-- name: AddSongToPlaylist :one
INSERT INTO local_playlist_tracks (playlist_id, spotify_track_id, position, updated_at)
VALUES ($1, $2, $3, NOW())
RETURNING *;


-- name: GetPlaylistSongs :many
SELECT * FROM local_playlist_tracks WHERE playlist_id = $1 ORDER BY position;

-- name: GetPlaylistSongByID :one
SELECT * FROM local_playlist_tracks WHERE playlist_id = $1 AND spotify_track_id = $2;

-- name: RemoveSongFromPlaylist :exec
DELETE FROM local_playlist_tracks WHERE playlist_id = $1 AND spotify_track_id = $2;

-- name: UpdateSongPosition :exec
UPDATE local_playlist_tracks SET position = $2 WHERE spotify_track_id = $1;

-- name: GetMaxSongPosition :one
SELECT COALESCE(MAX(position), 0)::INT AS last_position FROM local_playlist_tracks WHERE playlist_id = $1;

-- name: ClearPlaylistSongs :exec
DELETE FROM local_playlist_tracks WHERE playlist_id = $1;
