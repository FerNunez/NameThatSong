-- =============================================================================
-- ARTIST OPERATIONS
-- =============================================================================

-- name: GetSpotifyArtist :one
SELECT * FROM spotify_artists WHERE id = $1;

-- name: UpsertSpotifyArtist :one
INSERT INTO spotify_artists (id, name, image_url, popularity, followers_total, genres, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    image_url = EXCLUDED.image_url,
    popularity = EXCLUDED.popularity,
    followers_total = EXCLUDED.followers_total,
    genres = EXCLUDED.genres,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- =============================================================================
-- ALBUM OPERATIONS  
-- =============================================================================

-- name: GetSpotifyAlbum :one
SELECT * FROM spotify_albums WHERE id = $1;

-- name: UpsertSpotifyAlbum :one
INSERT INTO spotify_albums (id, name, album_type, release_date, release_date_precision, total_tracks, image_url, label, popularity, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    album_type = EXCLUDED.album_type,
    release_date = EXCLUDED.release_date,
    release_date_precision = EXCLUDED.release_date_precision,
    total_tracks = EXCLUDED.total_tracks,
    image_url = EXCLUDED.image_url,
    label = EXCLUDED.label,
    popularity = EXCLUDED.popularity,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- =============================================================================
-- TRACK OPERATIONS
-- =============================================================================

-- name: GetSpotifyTrack :one
SELECT * FROM spotify_tracks WHERE id = $1;

-- name: UpsertSpotifyTrack :one
INSERT INTO spotify_tracks (id, name, album_id, duration_ms, disc_number, track_number, popularity, explicit, preview_url, is_local, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    album_id = EXCLUDED.album_id,
    duration_ms = EXCLUDED.duration_ms,
    disc_number = EXCLUDED.disc_number,
    track_number = EXCLUDED.track_number,
    popularity = EXCLUDED.popularity,
    explicit = EXCLUDED.explicit,
    preview_url = EXCLUDED.preview_url,
    is_local = EXCLUDED.is_local,
    updated_at = EXCLUDED.updated_at
RETURNING *;


-- =============================================================================
-- PLAYLIST OPERATIONS
-- =============================================================================

-- name: GetSpotifyPlaylist :one
SELECT * FROM spotify_playlists WHERE id = $1;

-- name: UpsertSpotifyPlaylist :one
INSERT INTO spotify_playlists (id, name, description, owner_id, owner_display_name, public, collaborative, followers_total, total_tracks, image_url, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    owner_id = EXCLUDED.owner_id,
    owner_display_name = EXCLUDED.owner_display_name,
    public = EXCLUDED.public,
    collaborative = EXCLUDED.collaborative,
    followers_total = EXCLUDED.followers_total,
    total_tracks = EXCLUDED.total_tracks,
    image_url = EXCLUDED.image_url,
    updated_at = EXCLUDED.updated_at
RETURNING *;


-- =============================================================================
-- RELATIONSHIP OPERATIONS
-- =============================================================================

-- name: UpsertAlbumArtist :exec
INSERT INTO spotify_album_artists (album_id, artist_id)
VALUES ($1, $2)
ON CONFLICT (album_id, artist_id) DO NOTHING;

-- name: ClearAlbumArtists :exec
DELETE FROM spotify_album_artists WHERE album_id = $1;

-- name: UpsertTrackArtist :exec
INSERT INTO spotify_track_artists (track_id, artist_id, is_primary)
VALUES ($1, $2, $3)
ON CONFLICT (track_id, artist_id) DO UPDATE SET
    is_primary = EXCLUDED.is_primary;

-- name: ClearTrackArtists :exec
DELETE FROM spotify_track_artists WHERE track_id = $1;

-- ALBUMS
-- name: UpsertAlbumTrack :exec
INSERT INTO spotify_album_tracks (album_id, track_id, position)
VALUES ($1, $2, $3)
ON CONFLICT (album_id, track_id) DO UPDATE SET
    position  = EXCLUDED.position;
-- name: GetAlbumTracks :many
SELECT * FROM spotify_album_tracks WHERE album_id = $1;
-- name: GetAlbumByTrackID :one
SELECT album_id FROM spotify_album_tracks WHERE track_id = $1;
-- name: ClearAlbumTracks :exec
DELETE FROM spotify_album_tracks WHERE album_id = $1;

-- PLAYLISTS 
-- name: UpsertPlaylistTracks :exec
INSERT INTO spotify_playlist_tracks (
    playlist_id, track_id, position, updated_at
)
VALUES ($1, $2, $3, $4)
ON CONFLICT (playlist_id, track_id) DO UPDATE SET
    position  = EXCLUDED.position;
-- name: GetPlaylistTracks :many
SELECT * FROM spotify_playlist_tracks WHERE playlist_id = $1;
-- name: DeletePlaylistTrack :exec
DELETE FROM spotify_playlist_tracks WHERE track_id = $1;
-- name: ClearPlaylistTracks :exec
DELETE FROM spotify_playlist_tracks WHERE playlist_id = $1;

-- =============================================================================
-- CACHE MANAGEMENT OPERATIONS
-- =============================================================================

-- name: CleanupOldSpotifyArtists :exec
DELETE FROM spotify_artists WHERE cached_at < $1;

-- name: CleanupOldSpotifyAlbums :exec
DELETE FROM spotify_albums WHERE cached_at < $1;

-- name: CleanupOldSpotifyTracks :exec
DELETE FROM spotify_tracks WHERE cached_at < $1;

-- name: CleanupOldSpotifyPlaylists :exec
DELETE FROM spotify_playlists WHERE cached_at < $1;

-- name: GetCacheStats :one
SELECT 
    (SELECT COUNT(*) FROM spotify_artists) as artists_count,
    (SELECT COUNT(*) FROM spotify_albums) as albums_count,
    (SELECT COUNT(*) FROM spotify_tracks) as tracks_count,
    (SELECT COUNT(*) FROM spotify_playlists) as playlists_count;

-- =============================================================================
-- BATCH OPERATIONS
-- =============================================================================

-- name: GetMultipleSpotifyTracks :many
SELECT * FROM spotify_tracks 
WHERE id = ANY($1::text[]);

-- name: UpsertMultipleSpotifyTracksFromJSON :exec
INSERT INTO spotify_tracks (
    id, name, album_id, duration_ms, disc_number, track_number,
    popularity, explicit, preview_url, is_local, updated_at
)
SELECT 
    (track->>'id')::text,
    (track->>'name')::text,
    NULLIF(track->>'album_id', '')::text,
    (track->>'duration_ms')::int,
    (track->>'disc_number')::int,
    (track->>'track_number')::int,
    (track->>'popularity')::int,
    (track->>'explicit')::boolean,
    NULLIF(track->>'preview_url', ''),
    (track->>'is_local')::boolean,
    NOW()
FROM jsonb_array_elements($1::jsonb) AS track
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    album_id = EXCLUDED.album_id,
    duration_ms = EXCLUDED.duration_ms,
    disc_number = EXCLUDED.disc_number,
    track_number = EXCLUDED.track_number,
    popularity = EXCLUDED.popularity,
    explicit = EXCLUDED.explicit,
    preview_url = EXCLUDED.preview_url,
    is_local = EXCLUDED.is_local,
    updated_at = EXCLUDED.updated_at;

-- =============================================================================
-- EFFICIENT SEARCH OPERATIONS
-- =============================================================================

-- name: SearchSpotifyArtistsByName :many
SELECT id, name, image_url, popularity, followers_total, genres
FROM spotify_artists 
WHERE name ILIKE '%' || $1 || '%'
ORDER BY popularity DESC, name
LIMIT $2;

-- name: SearchSpotifyAlbumsByName :many
SELECT a.id, a.name, a.album_type, a.release_date, a.total_tracks, a.image_url,
       string_agg(ar.name, ', ') as artist_names
FROM spotify_albums a
LEFT JOIN spotify_album_artists aa ON a.id = aa.album_id
LEFT JOIN spotify_artists ar ON aa.artist_id = ar.id
WHERE a.name ILIKE '%' || $1 || '%'
GROUP BY a.id, a.name, a.album_type, a.release_date, a.total_tracks, a.image_url
ORDER BY a.popularity DESC, a.name
LIMIT $2;

-- name: SearchSpotifyTracksByName :many
SELECT t.id, t.name, t.duration_ms, t.popularity, t.explicit,
       a.name as album_name, a.image_url as album_image_url,
       string_agg(ar.name, ', ') as artist_names
FROM spotify_tracks t
LEFT JOIN spotify_albums a ON t.album_id = a.id
LEFT JOIN spotify_track_artists ta ON t.id = ta.track_id
LEFT JOIN spotify_artists ar ON ta.artist_id = ar.id
WHERE t.name ILIKE '%' || $1 || '%'
GROUP BY t.id, t.name, t.duration_ms, t.popularity, t.explicit, a.name, a.image_url
ORDER BY t.popularity DESC, t.name
LIMIT $2;
