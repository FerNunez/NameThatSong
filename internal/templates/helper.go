package templates

import (
	"sort"
	"strings"
)

func TruncateString(nbShowChars int, input string) string {
	return input[0:nbShowChars] + ".." + input[len(input)-nbShowChars:]
}

func TruncateAndJoin(maxNumbChars int, artistNameList []string) string {
	if len(artistNameList) == 0 {
		return ""
	}

	truncateStringList(maxNumbChars, &artistNameList)
	// Join
	result := strings.Join(artistNameList, ", ")
	return result
}

func truncateStringList(maxNumbChars int, stringList *[]string) {
	if stringList == nil || len(*stringList) == 0 {
		return
	}

	// Create pairs with index and string
	type pair struct {
		idx  int
		word string
	}

	pairs := make([]pair, len(*stringList))
	for i, v := range *stringList {
		pairs[i] = pair{i, v}
	}

	// Sort by length in descending order
	sort.Slice(pairs, func(i, j int) bool {
		return len(pairs[i].word) > len(pairs[j].word)
	})

	// Calculate total size including separators
	totalSize := 0
	for i, artist := range *stringList {
		totalSize += len(artist)
		if i < len(*stringList)-1 {
			totalSize += 2 // for ", "
		}
	}

	// Truncate longest strings first until we're under the limit
	idxLastChanged := 0
	for totalSize > maxNumbChars && idxLastChanged < len(pairs) {
		originalLen := len(pairs[idxLastChanged].word)
		idxLongestWord := pairs[idxLastChanged].idx

		// Truncate the string in place
		(*stringList)[idxLongestWord] = TruncateString(3, (*stringList)[idxLongestWord])
		newLen := len((*stringList)[idxLongestWord])

		// Update total size
		totalSize += newLen - originalLen

		// Update the pair's word for future comparisons
		pairs[idxLastChanged].word = (*stringList)[idxLongestWord]

		idxLastChanged++
	}
}

// func truncatestringlist(maxNumbChars int, stringList []string) []string {
// 	// Pair each string with its original index
// 	// and order decreasing
// 	type pair struct {
// 		idx  int
// 		word string
// 	}
// 	pairs := make([]pair, len(stringList))
// 	for i, v := range stringList {
// 		pairs[i] = pair{i, v}
// 	}
// 	sort.Slice(pairs, func(i, j int) bool {
// 		return len(pairs[i].word) > len(pairs[i].word)
// 	})
//
// 	totalSize := 0
// 	for i, artist := range stringList {
// 		totalSize += len(artist)
// 		if i < len(stringList)-1 && len(pairs) > 1 {
// 			totalSize += 2 // for ", "
// 		}
// 	}
//
// 	idxLastChanged := 0
// 	for totalSize > maxNumbChars && idxLastChanged < len(pairs) {
// 		totalSize += 8 - len(pairs[idxLastChanged].word)
// 		idxLongestWord := pairs[idxLastChanged].idx
// 		TruncateString(3, stringList[idxLongestWord])
// 		idxLastChanged += 1
// 	}
//
// }

// if len(artistNameList) == 1{
//
// 		if len(artistNameList[0]<maxNumbChars ){
// 			return
// 		}
//
//
// 	}
//
// 	if totalSize <= maxNumbChars &&  {
// 		return artistNameList[0]
// 	} else if totalSize <= maxNumbChars {
// 		return strings.Join(artistNameList, ", ")
// 	}
//
// }
