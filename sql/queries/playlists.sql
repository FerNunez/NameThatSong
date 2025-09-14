-- Playlist operations
-- name: UpsertPlaylist :one
INSERT INTO local_playlists (
    id, user_id, name, description, image_url,
    spotify_playlist_id, snapshot_id, is_public,
    created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
ON CONFLICT (id)
DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    image_url = EXCLUDED.image_url,
    spotify_playlist_id = EXCLUDED.spotify_playlist_id,
    snapshot_id = EXCLUDED.snapshot_id,
    is_public = EXCLUDED.is_public,
    updated_at = NOW()
RETURNING *;

-- name: GetPlaylistByID :one
SELECT * FROM local_playlists WHERE id = $1;

-- name: GetPlaylistByIDWithTracks :many
SELECT
    lp.id as playlist_id,
    lp.user_id,
    lp.name as playlist_name,
    lp.description,
    lp.image_url,
    lp.spotify_playlist_id,
    lp.snapshot_id,
    lp.is_public,
    lp.last_synced_at,
    lp.created_at,
    lp.updated_at,
    -- Track data (nullable for playlists with no tracks)
    lpt.spotify_track_id,
    lpt.position,
    lpt.updated_at as track_updated_at,
    st.name as track_name,
    st.duration_ms,
    st.disc_number,
    st.track_number,
    st.popularity,
    st.explicit,
    st.is_local,
    st.album_id,
    st.artist_ids,
    st.cached_at
FROM local_playlists lp
LEFT JOIN local_playlist_tracks lpt ON lp.id = lpt.playlist_id
LEFT JOIN spotify_tracks st ON lpt.spotify_track_id = st.id
WHERE lp.id = $1 AND lp.user_id = $2
ORDER BY lpt.position ASC;

-- name: GetPlaylistsByUserID :many
SELECT * FROM local_playlists WHERE user_id = $1 ORDER BY created_at DESC;

-- name: GetPlaylistByUserIDAndID :one
SELECT * FROM local_playlists WHERE id = $1 AND user_id = $2;

-- name: GetPlaylistBySpotifyIDAndUserID :one
SELECT * FROM local_playlists WHERE spotify_playlist_id = $1 AND user_id = $2 ORDER BY name DESC;

-- name: DeletePlaylist :exec
DELETE FROM local_playlists WHERE id = $1 AND user_id = $2;

-- Playlist songs operations
-- name: AddSongToPlaylist :one
INSERT INTO local_playlist_tracks (playlist_id, spotify_track_id, position, updated_at)
VALUES ($1, $2, $3, NOW())
RETURNING *;

--  ['track1', 'track2'] and [1, 2], it will create two rows with the corresponding track/position pairs
-- name: BulkInsertPlaylistTracks :exec
INSERT INTO local_playlist_tracks (playlist_id, spotify_track_id, position, updated_at)
SELECT $1, unnest($2::text[]), unnest($3::int[]), NOW();

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

-- name: GetPlaylistsByUserIDWithTracks :many
SELECT 
    lp.id as playlist_id,
    lp.user_id,
    lp.name as playlist_name,
    lp.description,
    lp.image_url,
    lp.spotify_playlist_id,
    lp.snapshot_id,
    lp.is_public,
    lp.last_synced_at,
    lp.created_at,
    lp.updated_at,
    -- Track data (nullable for playlists with no tracks)
    lpt.spotify_track_id,
    lpt.position,
    lpt.updated_at as track_updated_at,
    st.name as track_name,
    st.duration_ms,
    st.disc_number,
    st.track_number,
    st.popularity,
    st.explicit,
    st.is_local,
    st.album_id,
    st.artist_ids,
    st.cached_at
FROM local_playlists lp
LEFT JOIN local_playlist_tracks lpt ON lp.id = lpt.playlist_id
LEFT JOIN spotify_tracks st ON lpt.spotify_track_id = st.id
WHERE lp.user_id = $1
ORDER BY lp.created_at DESC, lpt.position ASC;
