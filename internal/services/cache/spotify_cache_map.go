package cache

import (
	m "github.com/FerNunez/NameThatSong/internal/models"
	//"github.com/FerNunez/NameThatSong/internal/services/spotify"
)

type SpotifyCacheMap struct {
	// Core informations
	ArtistMap   map[string]m.ArtistData
	AlbumMap    map[string]m.AlbumData
	TrackMap    map[string]m.TrackData
	PlaylistMap map[string]m.PlaylistData

	// Many-to-many relationship maps
	ArtistToAlbumsMap   map[string][]string
	AlbumToTracksMap    map[string][]string
	PlaylistToTracksMap map[string][]string

	// Test: SearchToMap

}

func NewSpotifyCacheMap() *SpotifyCacheMap {
	return &SpotifyCacheMap{
		ArtistMap:           map[string]m.ArtistData{},
		ArtistToAlbumsMap:   map[string][]string{},
		AlbumMap:            map[string]m.AlbumData{},
		AlbumToTracksMap:    map[string][]string{},
		TrackMap:            map[string]m.TrackData{},
		PlaylistMap:         map[string]m.PlaylistData{},
		PlaylistToTracksMap: map[string][]string{},
	}
}

// func (c *SpotifyCacheMap) GetTrack(accessToken, trackId string) (m.TrackData, error) {
// 	if val, ok := c.TrackMap[trackId]; ok {
// 		return val, nil
// 	}
// 	fmt.Println("[SpotifyCacheMap] GetTrack miss chache trackId:", trackId)
// 	trackData, err := spotify.FetchTrack(accessToken, trackId)
// 	if err != nil {
// 		return m.TrackData{}, err
// 	}
// 	c.TrackMap[trackId] = trackData
// 	return trackData, nil
// }
//
// func (c *SpotifyCacheMap) GetAlbum(accessToken, albumId string) (m.AlbumData, error) {
// 	if val, ok := c.AlbumMap[albumId]; ok {
// 		return val, nil
// 	}
// 	fmt.Println("[SpotifyCacheMap] GetAlbum miss chache on AlbumId:", albumId)
// 	albumData, tracksData, err := spotify.FetchAlbum(accessToken, albumId)
// 	if err != nil {
// 		return m.AlbumData{}, err
// 	}
// 	c.AlbumMap[albumId] = albumData
// 	for _, track := range tracksData {
// 		c.TrackMap[track.ID] = track
// 	}
// 	return albumData, nil
// }
// func (c *SpotifyCacheMap) GetArtist(accessToken, artistId string) (m.ArtistData, error) {
// 	if val, ok := c.ArtistMap[artistId]; ok {
// 		return val, nil
// 	}
// 	fmt.Println("[SpotifyCacheMap] GetArtist miss chache on ArtistID:", artistId)
// 	artistData, err := spotify.FetchArtist(accessToken, artistId)
// 	if err != nil {
// 		return m.ArtistData{}, err
// 	}
// 	c.ArtistMap[artistId] = artistData
// 	return artistData, nil
// }
//
// func (c *SpotifyCacheMap) GetPlaylist(accessToken, playlistId string) (m.PlaylistData, error) {
// 	if val, ok := c.PlaylistMap[playlistId]; ok {
// 		return val, nil
// 	}
// 	fmt.Println("[SpotifyCacheMap] GetPlaylist miss chache on playlistID:", playlistId)
// 	playlistData, tracksData, err := spotify.FetchPlaylist(accessToken, playlistId)
// 	if err != nil {
// 		return m.PlaylistData{}, err
// 	}
//
// 	for _, track := range tracksData {
// 		c.TrackMap[track.ID] = track
// 	}
//
// 	c.PlaylistMap[playlistId] = playlistData
// 	return playlistData, nil
// }
//
// func (c *SpotifyCacheMap) GetAlbumsFromArtist(accessToken, artistId string) ([]string, error) {
// 	if val, ok := c.ArtistToAlbumsMap[artistId]; ok {
// 		return val, nil
// 	}
// 	fmt.Println("[SpotifyCacheMap] GetAlbumsFromArtist miss chache on artistId:", artistId)
// 	albumIDs, err := spotify.FetchAlbumsFromArtist(accessToken, artistId)
// 	if err != nil {
// 		return []string{}, err
// 	}
// 	c.ArtistToAlbumsMap[artistId] = albumIDs
// 	return albumIDs, nil
// }
//
// func (c *SpotifyCacheMap) GetTracksFromAlbum(accessToken, albumId string) ([]string, error) {
// 	if val, ok := c.AlbumToTracksMap[albumId]; ok {
// 		return val, nil
// 	}
// 	fmt.Println("[SpotifyCacheMap] GetTracksFromAlbum miss chache on albumId:", albumId)
// 	trackIDs, err := spotify.FetchTracksFromAlbum(accessToken, albumId)
// 	if err != nil {
// 		return []string{}, err
// 	}
// 	c.AlbumToTracksMap[albumId] = trackIDs
// 	return trackIDs, nil
// }
//
// func (c *SpotifyCacheMap) GetTracksFromPlaylist(accessToken, playlistId string) ([]string, error) {
// 	if val, ok := c.PlaylistToTracksMap[playlistId]; ok {
// 		return val, nil
// 	}
// 	fmt.Println("[SpotifyCacheMap] GetTracksFromPlaylist miss chache on playlistId:", playlistId)
// 	trackIDs, err := spotify.FetchTracksFromPlaylist(accessToken, playlistId)
// 	if err != nil {
// 		return []string{}, err
// 	}
// 	c.PlaylistToTracksMap[playlistId] = trackIDs
// 	return trackIDs, nil
// }
//
// func (c *SpotifyCacheMap) SearchTracks(accessToken, query string) ([]m.TrackSearch, error) {
// 	panic("Shouldnt happy")
// 	return []m.TrackSearch{}, nil
// }
// func (c *SpotifyCacheMap) SearchAlbums(accessToken, query string) ([]m.AlbumSearch, error) {
// 	panic("Shouldnt happy")
// 	return nil, nil
// }
// func (c *SpotifyCacheMap) SearchArtists(accessToken, query string) ([]m.ArtistSearch, error) {
// 	panic("Shouldnt happy")
// 	return nil, nil
// }
// func (c *SpotifyCacheMap) SearchPlaylists(accessToken, query string) ([]m.PlaylistSearch, error) {
// 	panic("Shouldnt happy")
// }
//
// /////////////////
// //
// // func (c *SpotifyCacheMap) GetArtistsByName(s *spotify.SpotifySongProvider, artistName string, limit int) ([]spotify_api.ArtistData, error) {
// //
// // 	return []m.ArtistData{}, nil
// // }
// // func (c *SpotifyCacheMap) GetAlbumsIdFromArtistId(s *spotify.SpotifySongProvider, artistId string) ([]string, error) {
// // 	return []string{}, nil
// // }
// // func (c *SpotifyCacheMap) GetArtistData(s *spotify.SpotifySongProvider, id string) (spotify_api.ArtistData, error) {
// // 	//s.Cache.ArtistMap[artist.Id] = artist
// // 	return m.ArtistData{}, nil
// // }
// // func (c *SpotifyCacheMap) GetArtistsAlbum(s *spotify.SpotifySongProvider, accessToken, artistId string) ([]spotify_api.AlbumData, error) {
// // 	// check if artist already known
// // 	albumsIds, exist := c.ArtistToAlbumsMap[artistId]
// //
// // 	if !exist {
// // 		// DEBUG
// // 		fmt.Println("[SpotifyCacheMap]GetArtistsAlbum: cache miss for ", artistId)
// // 		// get artist trop track
// // 		albumTopTrack, topTracks, err := s.CreateTopTracksAlbum(accessToken, artistId)
// // 		artist, ok := c.ArtistMap[artistId]
// // 		albumTopTrack.ImagesURL = artist.ImageUrl
// // 		if !ok {
// // 			panic("Track should always exist in cache")
// // 		}
// // 		if err != nil {
// // 			return nil, err
// // 		}
// //
// // 		albums, err := s.FetchAlbumByArtistID(accessToken, artistId)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// //
// // 		// update Artist to albumMaps
// // 		albumsIds = make([]string, 0, len(albums)+1)
// // 		albumsIds = append(albumsIds, albumTopTrack.ID)
// // 		c.AlbumMap[albumTopTrack.ID] = albumTopTrack
// // 		for _, album := range albums {
// // 			c.AlbumMap[album.ID] = album
// // 			albumsIds = append(albumsIds, album.ID)
// // 			c.AlbumIdToArtistId[album.ID] = artistId
// // 		}
// //
// // 		c.ArtistToAlbumsMap[artistId] = albumsIds
// //
// // 		// associate AlbumID for top tracks
// // 		tracksIds := make([]string, 0, len(topTracks))
// // 		for _, track := range topTracks {
// // 			c.TrackMap[track.ID] = track
// // 			tracksIds = append(tracksIds, track.ID)
// // 		}
// // 		c.AlbumToTracksMap[albumTopTrack.ID] = tracksIds
// // 	} else {
// // 		// DEBUG
// // 		fmt.Println("[SpotifyCacheMap]GetArtistsAlbum: cache hit for ", artistId)
// // 	}
// //
// // 	albums := make([]m.AlbumData, 0, len(albumsIds))
// // 	for _, albumId := range albumsIds {
// // 		album, _ := c.AlbumMap[albumId]
// // 		albums = append(albums, album)
// //
// // 	}
// //
// // 	return albums, nil
// // }
// //
// // func (c *SpotifyCacheMap) GetAlbumTracks(s *spotify.SpotifySongProvider, accessToken, albumId string) ([]spotify_api.TrackData, error) {
// // 	tracksIds, exist := c.AlbumToTracksMap[albumId]
// // 	if !exist {
// // 		fmt.Println("[SpotifyCacheMap]GetAlbumTracks: cache missed for", albumId)
// // 		tracks, err := s.FetchTracksByAlbumID(accessToken, albumId)
// // 		if err != nil {
// // 			return nil, err
// // 		}
// // 		tracksIds = make([]string, 0, len(tracks))
// // 		for _, track := range tracks {
// // 			c.TrackMap[track.ID] = track
// // 			tracksIds = append(tracksIds, track.ID)
// //
// // 			c.TrackIdToAlbumId[track.ID] = albumId
// // 		}
// // 		c.AlbumToTracksMap[albumId] = tracksIds
// // 	} else {
// // 		fmt.Println("[SpotifyCacheMap]GetAlbumTracks: cache hit for", albumId)
// // 	}
// //
// // 	tracks := make([]m.TrackData, 0, len(tracksIds))
// // 	for _, trackId := range tracksIds {
// // 		track, _ := c.TrackMap[trackId]
// // 		tracks = append(tracks, track)
// // 	}
// // 	return tracks, nil
// // }
