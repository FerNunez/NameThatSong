package utils

import (
	"goth/internal/base"
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
