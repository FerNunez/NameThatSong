package spotify_api

// // GetSongsByIDs retrieves songs by their IDs
// func (p *SpotifySongProvider) GetSongsByIDs(ids []string) ([]player.Song, error) {
// 	result := make([]player.Song, 0, len(ids))
// 	missingIDs := make([]string, 0)
//
// 	// If all songs were in cache, return immediately
// 	if len(missingIDs) == 0 {
// 		return result, nil
// 	}
//
// 	// Otherwise, fetch the missing songs from the API
// 	// We can fetch up to 50 songs at once
// 	for i := 0; i < len(missingIDs); i += 50 {
// 		end := min(i+50, len(missingIDs))
//
// 		batchIDs := missingIDs[i:end]
// 		songs, err := p.fetchSongs(batchIDs)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		result = append(result, songs...)
// 	}
//
// 	return result, nil
// }
//
// // GetSongsByAlbumID retrieves all songs from a specific album
// func (p *SpotifySongProvider) GetSongsByAlbumID(albumID string) ([]player.Song, error) {
// 	// If not cached, fetch from API
// 	requestURL := fmt.Sprintf("https://api.spotify.com/v1/albums/%s/tracks", albumID)
// 	req, err := http.NewRequest("GET", requestURL, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("could not create request: %v", err)
// 	}
//
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", p.AccessToken))
//
// 	client := http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, fmt.Errorf("could not execute request: %v", err)
// 	}
// 	defer resp.Body.Close()
//
// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
// 	}
//
// 	var albumTracksResponse struct {
// 		Items []struct {
// 			ID          string `json:"id"`
// 			Name        string `json:"name"`
// 			URI         string `json:"uri"`
// 			DurationMs  int    `json:"duration_ms"`
// 			DiscNumber  int    `json:"disc_number"`
// 			TrackNumber int    `json:"track_number"`
// 			Artists     []struct {
// 				ID   string `json:"id"`
// 				Name string `json:"name"`
// 			} `json:"artists"`
// 		} `json:"items"`
// 	}
//
// 	if err := json.NewDecoder(resp.Body).Decode(&albumTracksResponse); err != nil {
// 		return nil, fmt.Errorf("could not decode response: %v", err)
// 	}
//
// 	// Convert to our Song type
// 	songs := make([]player.Song, 0, len(albumTracksResponse.Items))
// 	songIDs := make([]string, 0, len(albumTracksResponse.Items))
//
// 	for _, item := range albumTracksResponse.Items {
// 		artistIDs := make([]string, len(item.Artists))
// 		for i, artist := range item.Artists {
// 			artistIDs[i] = artist.ID
// 		}
//
// 		song := player.Song{
// 			ID:       item.ID,
// 			Name:     item.Name,
// 			URI:      item.URI,
// 			Duration: item.DurationMs,
// 			AlbumID:  albumID,
// 			Artists:  artistIDs,
// 		}
//
// 		songs = append(songs, song)
// 		songIDs = append(songIDs, item.ID)
//
// 	}
//
// 	return songs, nil
// }
//
// // GetSongsByAlbumIDs retrieves songs for multiple albums
// func (p *SpotifySongProvider) GetSongsByAlbumIDs(albumIDs []string) (map[string][]player.Song, error) {
// 	result := make(map[string][]player.Song, len(albumIDs))
//
// 	// We can parallelize the album requests for better performance
// 	var wg sync.WaitGroup
// 	var mu sync.Mutex
// 	var firstError error
//
// 	for _, albumID := range albumIDs {
// 		wg.Add(1)
// 		go func(id string) {
// 			defer wg.Done()
//
// 			songs, err := p.GetSongsByAlbumID(id)
// 			if err != nil {
// 				mu.Lock()
// 				if firstError == nil {
// 					firstError = err
// 				}
// 				mu.Unlock()
// 				return
// 			}
//
// 			mu.Lock()
// 			result[id] = songs
// 			mu.Unlock()
// 		}(albumID)
// 	}
//
// 	wg.Wait()
//
// 	if firstError != nil {
// 		return nil, firstError
// 	}
//
// 	return result, nil
// }
//
// // fetchSongs fetches song details from Spotify API
// func (p *SpotifySongProvider) fetchSong(id string) (player.Song, error) {
//
// 	requestURL := fmt.Sprintf("https://api.spotify.com/v1/tracks?ids=%s", id)
//
// 	req, err := http.NewRequest("GET", requestURL, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("could not create request: %v", err)
// 	}
//
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", p.AccessToken))
//
// 	client := http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return nil, fmt.Errorf("could not execute request: %v", err)
// 	}
// 	defer resp.Body.Close()
//
// 	if resp.StatusCode != http.StatusOK {
// 		return nil, fmt.Errorf("unexpected status code: %v", resp.StatusCode)
// 	}
//
// 	var tracksResponse struct {
// 		Tracks []struct {
// 			ID         string `json:"id"`
// 			Name       string `json:"name"`
// 			URI        string `json:"uri"`
// 			DurationMs int    `json:"duration_ms"`
// 			Album      struct {
// 				ID string `json:"id"`
// 			} `json:"album"`
// 			Artists []struct {
// 				ID string `json:"id"`
// 			} `json:"artists"`
// 		} `json:"tracks"`
// 	}
//
// 	if err := json.NewDecoder(resp.Body).Decode(&tracksResponse); err != nil {
// 		return nil, fmt.Errorf("could not decode response: %v", err)
// 	}
//
// 	songs := make([]player.Song, 0, len(tracksResponse.Tracks))
//
// 	for _, track := range tracksResponse.Tracks {
// 		artistIDs := make([]string, len(track.Artists))
// 		for i, artist := range track.Artists {
// 			artistIDs[i] = artist.ID
// 		}
//
// 		song := player.Song{
// 			ID:       track.ID,
// 			Name:     track.Name,
// 			URI:      track.URI,
// 			Duration: track.DurationMs,
// 			AlbumID:  track.Album.ID,
// 			Artists:  artistIDs,
// 		}
//
// 		songs = append(songs, song)
// 	}
//
// 	return songs, nil
// }
