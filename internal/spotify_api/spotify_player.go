package spotify_api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// PlaySong starts playing a specific song on the user's active device
func (p *SpotifySongProvider) PlaySong(accessToken, songID string) error {
	type PlaySongRequest struct {
		Uris       []string `json:"uris"`
		PositionMs int      `json:"position_ms"`
	}

	psr := PlaySongRequest{
		Uris:       []string{fmt.Sprintf("spotify:track:%v", songID)},
		PositionMs: 0,
	}

	dat, err := json.Marshal(psr)
	if err != nil {
		return fmt.Errorf("could not marshal request: %v", err)
	}

	// Set request
	url := "https://api.spotify.com/v1/me/player/play"
	//if c.DeviceID != "" {
	//	url = fmt.Sprintf("%s?device_id=%s", url, c.DeviceID)
	//}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(dat))
	if err != nil {
		return fmt.Errorf("could not create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	return nil
}

// PausePlayback pauses the current playback on the user's active device
func (p *SpotifySongProvider) PausePlayback(accessToken string) error {
	url := "https://api.spotify.com/v1/me/player/pause"
	//if c.DeviceID != "" {
	//	url = fmt.Sprintf("%s?device_id=%s", url, c.DeviceID)
	//}

	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return fmt.Errorf("could not create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	return nil
}

// ResumePlayback resumes the current playback on the user's active device
func (p *SpotifySongProvider) ResumePlayback(accessToken string) error {
	url := "https://api.spotify.com/v1/me/player/play"
	// if c.DeviceID != "" {
	// 	url = fmt.Sprintf("%s?device_id=%s", url, c.DeviceID)
	// }
	//
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return fmt.Errorf("could not create request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", accessToken))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
	}

	return nil
}

// SkipToNext skips to the next track in the user's queue
// func (p *SpotifySongProvider) SkipToNext() error {
// 	url := "https://api.spotify.com/v1/me/player/next"
// 	if c.DeviceID != "" {
// 		url = fmt.Sprintf("%s?device_id=%s", url, c.DeviceID)
// 	}
//
// 	req, err := http.NewRequest("POST", url, nil)
// 	if err != nil {
// 		return fmt.Errorf("could not create request: %v", err)
// 	}
//
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.AccessToken))
//
// 	client := http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return fmt.Errorf("could not execute request: %v", err)
// 	}
// 	defer resp.Body.Close()
//
// 	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
// 		return fmt.Errorf("unexpected status code: %v", resp.StatusCode)
// 	}
//
// 	return nil
// }

// GetActiveDevice gets the user's active Spotify device
//func (c *SpotifyMusicServiceClient) GetActiveDevice() (string, error) {
//	// If a device ID is already set, return it
//	if c.DeviceID != "" {
//		return c.DeviceID, nil
//	}
//
//	// Otherwise, query the Spotify API for the user's devices
//	req, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/player/devices", nil)
//	if err != nil {
//		return "", fmt.Errorf("could not create request: %v", err)
//	}
//
//	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", c.AccessToken))
//
//	client := http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		return "", fmt.Errorf("could not execute request: %v", err)
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK {
//		return "", fmt.Errorf("unexpected status code: %v", resp.StatusCode)
//	}
//
//	var devicesResponse struct {
//		Devices []struct {
//			ID            string `json:"id"`
//			IsActive      bool   `json:"is_active"`
//			Name          string `json:"name"`
//			VolumePercent int    `json:"volume_percent"`
//		} `json:"devices"`
//	}
//
//	if err := json.NewDecoder(resp.Body).Decode(&devicesResponse); err != nil {
//		return "", fmt.Errorf("could not decode response: %v", err)
//	}
//
//	// Look for an active device
//	for _, device := range devicesResponse.Devices {
//		if device.IsActive {
//			c.DeviceID = device.ID
//			return device.ID, nil
//		}
//	}
//
//	// If no active device is found but there's at least one device, use the first one
//	if len(devicesResponse.Devices) > 0 {
//		c.DeviceID = devicesResponse.Devices[0].ID
//		return devicesResponse.Devices[0].ID, nil
//	}
//
//	return "", errors.New("no active device found")
//}
