package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

func GenerateState(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func ParseGuessText(input string) string {

	words := strings.Split(strings.ToLower(input), " ")

	artists := []string{}
	albums := []string{}
	// todo: Add year, genre etc
	trackName := ""

	for i := 0; i < len(words); i++ {
		w := words[i]

		fmt.Println(w)
		// has a prefix /
		if strings.HasPrefix(w, "/") {
			fmt.Println("YES")
			// has good conditions
			if len(w) >= 2 && i+1 <= len(words)-1 {
				cat := w[1:]
				println("goten cat: ", cat)
				switch cat {
				case "artist":
				case "ar":
				case "a":
					artists = append(artists, words[i+1])
					i += 1
				case "album":
				case "al":
				case "b":
					albums = append(albums, words[i+1])
					i += 1
				}
			}
			continue
		}
		trackName += w + " "
	}

	cleanedQuery := "track:" + trackName
	for _, artist := range artists {
		cleanedQuery += " artist:" + artist
	}
	for _, album := range albums {
		cleanedQuery += " album:" + album
	}

	return cleanedQuery
}
