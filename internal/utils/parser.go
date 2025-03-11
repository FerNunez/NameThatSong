package utils

import (
	"fmt"
	"goth/internal/base"
	"time"
)

func ParseArtistsJson(data base.ArtistsJson) ([]base.Artist, error) {

	var artists []base.Artist
	for _, item := range data.Items {
		artist := base.Artist{
			TotalFollowers: item.Followers.Total,
			Genres:         item.Genres,
			Href:           item.Href,
			ID:             item.ID,
			Images:         make([]base.Image, len(item.Images)),
			Name:           item.Name,
			Popularity:     item.Popularity,
			Type:           item.Type,
			URI:            item.URI,
			AlbumsID:       []string{}, // Initialize empty slice for album IDs
		}

		// Convert nested images to Image struct
		for i, img := range item.Images {
			artist.Images[i] = base.Image{
				URL:    img.URL,
				Height: img.Height,
				Width:  img.Width,
			}
		}

		artists = append(artists, artist)
	}
	return artists, nil
}

func ParseAlbumsJson(data base.AlbumsJson) ([]base.Album, error) {
	var albums []base.Album

	for _, item := range data.Items {

		// Parse release date with precision handling
		releaseDate, err := parseReleaseDate(item.ReleaseDate, item.ReleaseDatePrecision)
		if err != nil {
			return nil, fmt.Errorf("failed to parse release date for album %s: %v", item.ID, err)
		}

		// Convert JSON images to Image structs
		images := make([]base.Image, len(item.Images))
		for i, img := range item.Images {
			images[i] = base.Image{
				URL:    img.URL,
				Height: img.Height,
				Width:  img.Width,
			}
		}

		// Extract artist IDs
		artistIDs := make([]string, len(item.Artists))
		for i, artist := range item.Artists {
			artistIDs[i] = artist.ID
		}

		album := base.Album{
			AlbumType:        item.AlbumType,
			TotalTracks:      item.TotalTracks,
			AvailableMarkets: item.AvailableMarkets,
			Href:             item.Href,
			ID:               item.ID,
			Images:           images,
			Name:             item.Name,
			ReleaseDate:      releaseDate,
			URI:              item.URI,
			ArtistsFeatureID: artistIDs,
			Selected:         false,
		}

		albums = append(albums, album)
	}

	return albums, nil

}

func parseReleaseDate(dateStr, precision string) (time.Time, error) {
	var layout string
	switch precision {
	case "year":
		layout = "2006"
	case "month":
		layout = "2006-01"
	case "day":
		layout = "2006-01-02"
	default:
		return time.Time{}, fmt.Errorf("unknown precision format: %s", precision)
	}

	return time.Parse(layout, dateStr)
}
