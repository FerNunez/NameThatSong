-- =============================================================================
-- TRACK OPERATIONS
-- =============================================================================
-- name: GetSpotifyTrack :one
SELECT * FROM spotify_tracks WHERE id = $1;

-- name: UpsertSpotifyTrack :one
INSERT INTO spotify_tracks (id, name,  duration_ms, disc_number, track_number, popularity, explicit, is_local,album_id, artist_ids, cached_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    duration_ms = EXCLUDED.duration_ms,
    disc_number = EXCLUDED.disc_number,
    track_number = EXCLUDED.track_number,
    popularity = EXCLUDED.popularity,
    explicit = EXCLUDED.explicit,
    is_local = EXCLUDED.is_local,
    album_id = EXCLUDED.album_id,
    artist_ids = EXCLUDED.artist_ids,
    cached_at = EXCLUDED.cached_at
RETURNING *;

-- =============================================================================
-- ALBUM OPERATIONS  
-- =============================================================================
-- name: GetSpotifyAlbum :one
SELECT * FROM spotify_albums WHERE id = $1;

-- name: UpsertSpotifyAlbum :one
INSERT INTO spotify_albums (id, name, album_type, release_date, release_date_precision, total_tracks, image_url, label, popularity, track_ids, artist_ids, cached_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    album_type = EXCLUDED.album_type,
    release_date = EXCLUDED.release_date,
    release_date_precision = EXCLUDED.release_date_precision,
    total_tracks = EXCLUDED.total_tracks,
    image_url = EXCLUDED.image_url,
    label = EXCLUDED.label,
    popularity = EXCLUDED.popularity,
    track_ids = EXCLUDED.track_ids,
    artist_ids= EXCLUDED.artist_ids,
    cached_at = EXCLUDED.cached_at
RETURNING *;

-- =============================================================================
-- ARTIST OPERATIONS
-- =============================================================================
-- name: GetSpotifyArtist :one
SELECT * FROM spotify_artists WHERE id = $1;

-- name: UpsertSpotifyArtist :one
INSERT INTO spotify_artists (id, name, image_url, popularity, followers_total, genres, cached_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    image_url = EXCLUDED.image_url,
    popularity = EXCLUDED.popularity,
    followers_total = EXCLUDED.followers_total,
    genres = EXCLUDED.genres,
    cached_at = EXCLUDED.cached_at
RETURNING *;
-- =============================================================================
-- PLAYLIST OPERATIONS
-- =============================================================================

-- name: GetSpotifyPlaylist :one
SELECT * FROM spotify_playlists WHERE id = $1;

-- name: UpsertSpotifyPlaylist :one
INSERT INTO spotify_playlists (id, name, description, owner_id, owner_display_name, public, collaborative, followers_total, total_tracks, image_url, track_ids, cached_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
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
    track_ids = EXCLUDED.track_ids,
    cached_at = EXCLUDED.cached_at
RETURNING *;


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
    popularity, explicit, is_local, artist_ids, cached_at
)
SELECT 
    (track->>'id')::text,
    (track->>'name')::text,
    (track->>'album_id')::text,
    (track->>'duration_ms')::int,
    (track->>'disc_number')::int,
    (track->>'track_number')::int,
    (track->>'popularity')::int,
    (track->>'explicit')::boolean,
    (track->>'is_local')::boolean,
    ARRAY(SELECT jsonb_array_elements_text(track->'artist_ids'))::text[],
    (track->>'cached_at')::timestamp
FROM jsonb_array_elements($1::jsonb) AS track
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    album_id = EXCLUDED.album_id,
    duration_ms = EXCLUDED.duration_ms,
    disc_number = EXCLUDED.disc_number,
    track_number = EXCLUDED.track_number,
    popularity = EXCLUDED.popularity,
    explicit = EXCLUDED.explicit,
    is_local = EXCLUDED.is_local,
    artist_ids = EXCLUDED.artist_ids,
    cached_at = EXCLUDED.cached_at;

-- name: GetMultipleSpotifyAlbums :many
SELECT * FROM spotify_albums 
WHERE id = ANY($1::text[]);

-- name: UpsertMultipleSpotifyAlbumsFromJSON :exec
INSERT INTO spotify_albums (
    id,
    name,
    album_type,
    release_date,
    release_date_precision,
    total_tracks,
    image_url,
    label,
    popularity,
    artist_ids,
    track_ids,
    cached_at
)
SELECT 
    (album->>'id')::text,
    (album->>'name')::text,
    (album->>'album_type')::text,
    NULLIF(album->>'release_date', '')::date,
    (album->>'release_date_precision')::text,
    (album->>'total_tracks')::int,
    NULLIF(album->>'image_url','')::text,
    NULLIF(album->>'label', '')::text,
    (album->>'popularity')::int,
    ARRAY(SELECT jsonb_array_elements_text(album->'artist_ids'))::text[],
    ARRAY(SELECT jsonb_array_elements_text(album->'track_ids'))::text[],
    (album->>'cached_at')::timestamp
FROM jsonb_array_elements($1::jsonb) AS album
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    album_type = EXCLUDED.album_type,
    release_date = EXCLUDED.release_date,
    release_date_precision = EXCLUDED.release_date_precision,
    total_tracks = EXCLUDED.total_tracks,
    image_url = EXCLUDED.image_url,
    label = EXCLUDED.label,
    popularity = EXCLUDED.popularity,
    artist_ids = EXCLUDED.artist_ids,
    track_ids = EXCLUDED.track_ids,
    cached_at = EXCLUDED.cached_at;

-- name: GetMultipleSpotifyArtists :many
SELECT * FROM spotify_artists 
WHERE id = ANY($1::text[]);


-- name: UpsertMultipleSpotifyArtistsFromJSON :exec
INSERT INTO spotify_artists (
    id,
    name,
    image_url,
    popularity,
    followers_total,
    genres,
    cached_at
)
SELECT 
    (artist->>'id')::text,
    (artist->>'name')::text,
    NULLIF(artist->>'image_url', '')::text,
    (artist->>'popularity')::int,
    (artist->>'followers_total')::int,
    -- Handle genres as either JSON array or comma-separated string
    -- If JSON array: extract each element as text array
    -- If string: split by comma (fallback for legacy data)
    CASE 
        WHEN jsonb_typeof(artist->'genres') = 'array' 
        THEN ARRAY(SELECT jsonb_array_elements_text(artist->'genres'))
        ELSE string_to_array(NULLIF(artist->>'genres', ''), ',')::text[]
    END,
    (artist->>'cached_at')::timestamp
FROM jsonb_array_elements($1::jsonb) AS artist
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    image_url = EXCLUDED.image_url,
    popularity = EXCLUDED.popularity,
    followers_total = EXCLUDED.followers_total,
    genres = EXCLUDED.genres,
    cached_at = EXCLUDED.cached_at;

-- name: GetMultipleSpotifyPlaylists :many
SELECT * FROM spotify_playlists 
WHERE id = ANY($1::text[]);

-- name: UpsertMultipleSpotifyPlaylistsFromJSON :exec
INSERT INTO spotify_playlists (
    id,
    name,
    description,
    owner_id,
    owner_display_name,
    public,
    collaborative,
    followers_total,
    total_tracks,
    image_url,
    track_ids,
    cached_at
)
SELECT 
    (playlist->>'id')::text,
    (playlist->>'name')::text,
    NULLIF(playlist->>'description', '')::text,
    (playlist->>'owner_id')::text,
    NULLIF(playlist->>'owner_display_name', '')::text,
    (playlist->>'public')::boolean,
    (playlist->>'collaborative')::boolean,
    (playlist->>'followers_total')::int,
    (playlist->>'total_tracks')::int,
    NULLIF(playlist->>'image_url', '')::text,
    ARRAY(SELECT jsonb_array_elements_text(playlist->'track_ids'))::text[],
    (playlist->>'cached_at')::timestamp
FROM jsonb_array_elements($1::jsonb) AS playlist
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
    track_ids = EXCLUDED.track_ids,
    cached_at = EXCLUDED.cached_at;


