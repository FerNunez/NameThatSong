package cache

import (
	"fmt"

	"github.com/FerNunez/NameThatSong/internal/spotify_api"
)

type SpotifyCacheMap struct {
	ArtistMap         map[string]spotify_api.ArtistData
	ArtistToAlbumsMap map[string][]string
	AlbumMap          map[string]spotify_api.AlbumData
	AlbumToTracksMap  map[string][]string
	TrackMap          map[string]spotify_api.TrackData
	TrackIdToAlbumId  map[string]string
	AlbumIdToArtistId map[string]string
}

func NewSpotifyCacheMap() *SpotifyCacheMap {
	return &SpotifyCacheMap{
		ArtistMap:         map[string]spotify_api.ArtistData{},
		ArtistToAlbumsMap: map[string][]string{},
		AlbumMap:          map[string]spotify_api.AlbumData{},
		AlbumToTracksMap:  map[string][]string{},
		TrackMap:          map[string]spotify_api.TrackData{},
		TrackIdToAlbumId:  make(map[string]string),
		AlbumIdToArtistId: make(map[string]string),
	}
}

func (c *SpotifyCacheMap) GetArtistsByName(s *spotify_api.SpotifySongProvider, artistName string) ([]spotify_api.ArtistData, error) {

	return []spotify_api.ArtistData{}, nil
}

func (c *SpotifyCacheMap) GetTrack(s *spotify_api.SpotifySongProvider, accessToken, trackId string) (spotify_api.TrackData, error) {

	val, ok := c.TrackMap[trackId]
	if !ok {
		fmt.Println("[SpotifyCacheMap] GetTrack miss chache")

		// TODO: Fetching
		trackData, err := s.FetchTrack(accessToken, trackId)
		if err != nil {
			return spotify_api.TrackData{}, nil
		}
		c.TrackMap[trackId] = trackData
	}
	return val, nil
}
func (c *SpotifyCacheMap) GetAlbum(s *spotify_api.SpotifySongProvider, accessToken, albumId string) (spotify_api.AlbumData, error) {
	val, ok := c.AlbumMap[albumId]
	if !ok {
		fmt.Println("[SpotifyCacheMap] GetAlbum miss chache")

		albumData, err := s.FetchAlbum(accessToken, albumId)
		if err != nil {
			return spotify_api.AlbumData{}, nil
		}
		c.AlbumMap[albumId] = albumData
	}
	return val, nil
}
func (c *SpotifyCacheMap) GetArtist(s *spotify_api.SpotifySongProvider, accessToken, artistId string) (spotify_api.ArtistData, error) {
	val, ok := c.ArtistMap[artistId]
	if !ok {
		fmt.Println("[SpotifyCacheMap] GetArtist miss chache")

		artistData, err := s.FetchArtist(accessToken, artistId)
		if err != nil {
			return spotify_api.ArtistData{}, nil
		}
		c.ArtistMap[artistId] = artistData
	}
	return val, nil
}

func (c *SpotifyCacheMap) GetAlbumsIdFromArtistId(s *spotify_api.SpotifySongProvider, artistId string) ([]string, error) {
	return []string{}, nil
}
func (c *SpotifyCacheMap) GetArtistData(s *spotify_api.SpotifySongProvider, id string) (spotify_api.ArtistData, error) {
	//s.Cache.ArtistMap[artist.Id] = artist
	return spotify_api.ArtistData{}, nil
}
func (c *SpotifyCacheMap) GetArtistsAlbum(s *spotify_api.SpotifySongProvider, accessToken, artistId string) ([]spotify_api.AlbumData, error) {
	// check if artist already known
	albumsIds, exist := c.ArtistToAlbumsMap[artistId]

	if !exist {
		// DEBUG
		fmt.Println("[SpotifyCacheMap]GetArtistsAlbum: cache miss for ", artistId)
		// get artist trop track
		albumTopTrack, topTracks, err := s.CreateTopTracksAlbum(accessToken, artistId)
		artist, ok := c.ArtistMap[artistId]
		albumTopTrack.ImagesURL = artist.ImageUrl
		if !ok {
			panic("Track should always exist in cache")
		}
		if err != nil {
			return nil, err
		}

		albums, err := s.FetchAlbumByArtistID(accessToken, artistId)
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
			c.AlbumIdToArtistId[album.ID] = artistId
		}

		c.ArtistToAlbumsMap[artistId] = albumsIds

		// associate AlbumID for top tracks
		tracksIds := make([]string, 0, len(topTracks))
		for _, track := range topTracks {
			c.TrackMap[track.ID] = track
			tracksIds = append(tracksIds, track.ID)
		}
		c.AlbumToTracksMap[albumTopTrack.ID] = tracksIds
	} else {
		// DEBUG
		fmt.Println("[SpotifyCacheMap]GetArtistsAlbum: cache hit for ", artistId)
	}

	albums := make([]spotify_api.AlbumData, 0, len(albumsIds))
	for _, albumId := range albumsIds {
		album, _ := c.AlbumMap[albumId]
		albums = append(albums, album)

	}

	return albums, nil
}

func (c *SpotifyCacheMap) GetAlbumTracks(s *spotify_api.SpotifySongProvider, accessToken, albumId string) ([]spotify_api.TrackData, error) {
	tracksIds, exist := c.AlbumToTracksMap[albumId]
	if !exist {
		fmt.Println("[SpotifyCacheMap]GetAlbumTracks: cache missed for", albumId)
		tracks, err := s.FetchTracksByAlbumID(accessToken, albumId)
		if err != nil {
			return nil, err
		}
		tracksIds = make([]string, 0, len(tracks))
		for _, track := range tracks {
			c.TrackMap[track.ID] = track
			tracksIds = append(tracksIds, track.ID)

			c.TrackIdToAlbumId[track.ID] = albumId
		}
		c.AlbumToTracksMap[albumId] = tracksIds
	} else {
		fmt.Println("[SpotifyCacheMap]GetAlbumTracks: cache hit for", albumId)
	}

	tracks := make([]spotify_api.TrackData, 0, len(tracksIds))
	for _, trackId := range tracksIds {
		track, _ := c.TrackMap[trackId]
		tracks = append(tracks, track)
	}
	return tracks, nil
}
