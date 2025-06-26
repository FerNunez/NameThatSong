package templates

import "strings"

func ArtistsListString(maxNumbChars int, artistList []string) string {
	if len(artistList) == 0 {
		return ""
	}

	totalSize := 0
	for i, artist := range artistList {
		totalSize += len(artist)
		if i < len(artistList)-1 && len(artistList) > 1 {
			totalSize += 2 // for ", "
		}
	}

	if totalSize <= maxNumbChars && len(artistList) == 1 {
		return artistList[0]
	} else if totalSize <= maxNumbChars {
		return strings.Join(artistList, ", ")
	}

	// Otherwise, truncate long names
	result := ""
	for i := 0; i < len(artistList)-1; i++ {
		artist := artistList[i]
		if len(artist) > 9 {
			result = result + artist[0:3] + ".." + artist[len(artist)-3:] + ", "
		} else {
			result = result + artist + ", "
		}
	}

	// Handle last artist (no trailing comma)
	artist := artistList[len(artistList)-1]
	if len(artist) > 9 {
		result = result + artist[0:3] + ".." + artist[len(artist)-3:]
	} else {
		result = result + artist
	}
	return result
}
