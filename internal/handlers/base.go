package handlers

import (
	"fmt"
	"goth/internal/base"
	"goth/internal/guesser"
	"goth/internal/utils"
	"os"

	"github.com/joho/godotenv"
)

type ApiConfig struct {
	ClientID          string
	ClientSecret      string
	State             string
	AuthorizationCode string
	AccessToken       string
	RefreshToken      string
	DeviceId          string
}

type SpotifyApi struct {
	Config                      ApiConfig
	Cache                       base.ApiCache
	MusicPlayer                 guesser.MusicPlayer
	SelectedAlbumsIdsToTracksId map[string][]string
}

func NewSpotifyApi() SpotifyApi {
	godotenv.Load()
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")

	apiConfig := ApiConfig{
		ClientID:          clientID,
		ClientSecret:      clientSecret,
		State:             "",
		AuthorizationCode: "",
		AccessToken:       "",
		RefreshToken:      "",
		DeviceId:          "",
	}

	state, err := utils.GenerateState(16)
	if err != nil {
		errmsg := fmt.Sprintf("could not generate random state: %v", state)
		fmt.Println(errmsg)
	}

	return SpotifyApi{
		Config:                      apiConfig,
		Cache:                       base.NewApiCache(),
		MusicPlayer:                 guesser.MusicPlayer{},
		SelectedAlbumsIdsToTracksId: make(map[string][]string),
	}
}

func (cfg *SpotifyApi) ParserTracksResponse(searchTrackResponse base.SearchTrackResponse) ([]base.Track, []string, []string) {
	var tracks []base.Track
	var albumUris []string
	var artistsNames []string

	for _, item := range searchTrackResponse.Tracks.Items {

		// Album
		albumJson := item.Album
		images := make([]base.Image, len(albumJson.Images))
		for _, img := range albumJson.Images {
			images = append(images, base.NewImage(img.URL, img.Height, img.Width))
		}

		releaseDate, _ := utils.ParseReleaseDate(albumJson.ReleaseDate, albumJson.ReleaseDatePrecision)
		//todo error
		artistIDs := make([]string, len(item.Artists))
		for i, artist := range item.Artists {
			artistIDs[i] = artist.ID
		}

		album := base.Album{
			AlbumType:        albumJson.AlbumType,
			TotalTracks:      0,
			AvailableMarkets: albumJson.AvailableMarkets,
			Href:             albumJson.Href,
			ID:               albumJson.ID,
			Images:           images,
			Name:             albumJson.Name,
			ReleaseDate:      releaseDate,
			URI:              albumJson.URI,
			ArtistsFeatureID: artistIDs,
			Selected:         false,
		}
		// update
		cfg.Cache.InsertAlbum(album.ID, album)

		// Artist
		for _, artistJson := range item.Artists {

			artist := base.Artist{
				TotalFollowers: 0,
				Genres:         []string{"Unknown"},
				Href:           artistJson.Href,
				ID:             artistJson.ID,
				Images:         nil,
				Name:           artistJson.Name,
				Popularity:     0,
				Type:           artistJson.Type,
				URI:            artistJson.URI,
			}
			cfg.Cache.InsertArtist(artistJson.ID, artist)
		}

		// Collect artist IDs
		artistIDs = make([]string, 0, len(item.Artists))
		artistsName := ""
		for _, artist := range item.Artists {
			artistIDs = append(artistIDs, artist.ID)
			artistsName += artist.Name
		}

		if len(item.Album.Images) <= 0 {
			panic("No images found for album")
		}
		albumUris = append(albumUris, item.Album.Images[0].URL)
		artistsNames = append(artistsNames, artistsName)

		// Create track with all fields
		track := base.Track{
			DiscNumber:  item.DiscNumber,
			DurationMs:  item.DurationMs,
			Href:        item.Href,
			ID:          item.ID,
			IsPlayable:  item.IsPlayable,
			Name:        item.Name,
			Popularity:  item.Popularity,
			TrackNumber: item.TrackNumber,
			URI:         item.URI,
			AlbumID:     item.Album.ID,
			ArtistsID:   artistIDs,
			Selected:    false,
		}
		cfg.Cache.InsertTrack(track.ID, track)

		tracks = append(tracks, track)
	}

	return tracks, artistsNames, albumUris
}
