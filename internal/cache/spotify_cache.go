package cache

import (
	"goth/internal/spotify_api"
)

type SpotifyCache struct {
	ArtistToAlbumsMap map[string][]string
	AlbumMap          map[string]spotify_api.AlbumData
	AlbumToTracksMap  map[string][]string
	TrackMap          map[string]spotify_api.TrackData
	TrackIdToAlbumId  map[string]string
	AlbumIdToArtistId map[string]string
}

func NewSpotifyCache() *SpotifyCache {
	return &SpotifyCache{
		ArtistToAlbumsMap: map[string][]string{},
		AlbumMap:          map[string]spotify_api.AlbumData{},
		AlbumToTracksMap:  map[string][]string{},
		TrackMap:          map[string]spotify_api.TrackData{},
	}
}

func (c *SpotifyCache) GetArtistData(s *spotify_api.SpotifySongProvider, id string) (spotify_api.ArtistData, error) {
	return spotify_api.ArtistData{}, nil
}

func (c *SpotifyCache) GetArtistsAlbum(s *spotify_api.SpotifySongProvider, artistId string) ([]spotify_api.AlbumData, error) {
	// check if artist already known
	albumsIds, exist := c.ArtistToAlbumsMap[artistId]

	if !exist {
		// get artist trop track
		albumTopTrack, topTracks, err := s.CreateAlbumFromTopTracks(artistId)
		if err != nil {
			return nil, err
		}

		albums, err := s.FetchAlbumByArtistID(artistId)
		if err != nil {
			return nil, err
		}

		// update Artist to albumMaps
		albumsIds = make([]string, 0, len(albums)+1)
		albumsIds = append(albumsIds, albumTopTrack.ID)
		c.AlbumMap[albumTopTrack.ID] = albumTopTrack
		for _, album := range albums {
			c.AlbumMap[album.ID] = album
			albumsIds = append(albumsIds, album.ID)
			// c.AlbumIdToArtistId[album.ID] = artistId
		}
		c.ArtistToAlbumsMap[artistId] = albumsIds

		// associate AlbumID for top tracks
		tracksIds := make([]string, 0, len(topTracks))
		for _, track := range topTracks {
			c.TrackMap[track.ID] = track
			tracksIds = append(tracksIds, track.ID)
		}
		c.AlbumToTracksMap[albumTopTrack.ID] = tracksIds
	}

	albums := make([]spotify_api.AlbumData, 0, len(albumsIds))
	for _, albumId := range albumsIds {
		album, _ := c.AlbumMap[albumId]
		albums = append(albums, album)

	}

	return albums, nil
}

func (c *SpotifyCache) GetAlbumTracks(s *spotify_api.SpotifySongProvider, albumId string) ([]spotify_api.TrackData, error) {
	tracksIds, exist := c.AlbumToTracksMap[albumId]
	if !exist {
		tracks, err := s.FetchTracksByAlbumID(albumId)
		if err != nil {
			return nil, err
		}
		tracksIds = make([]string, 0, len(tracks))
		for _, track := range tracks {
			c.TrackMap[track.ID] = track
			tracksIds = append(tracksIds, track.ID)

			//s.Cache.TrackIdToAlbumId[track.ID] = albumId
		}
		c.AlbumToTracksMap[albumId] = tracksIds
	}

	tracks := make([]spotify_api.TrackData, 0, len(tracksIds))
	for _, trackId := range tracksIds {
		track, _ := c.TrackMap[trackId]
		tracks = append(tracks, track)
	}
	return tracks, nil
}
