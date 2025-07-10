package handlers

type GetSpotifyApi struct {
}

func NewGetSpotifyApi() *GetSpotifyApi {
	return &GetSpotifyApi{}
}

// func (h *GetSpotifyApi) ServeHttp(w http.ResponseWriter, r *http.Request) {
// 	component := templates.InternalSpotifyApi()
// 	component.Render(r.Context(), w)
//
// }
//
// type GetSpotifyApiTrack struct {
// 	accessToken string
// }
//
// func NewGetSpotifyApiTrack(accessToken string) *GetSpotifyApiTrack {
// 	return &GetSpotifyApiTrack{accessToken}
// }
// func (h *GetSpotifyApiTrack) ServeHttp(w http.ResponseWriter, r *http.Request) {
// 	trackID := "11dFghVXANMlKmJXsNCbNl"
// 	fetchTrack, err := spotify.FetchTrack(h.accessToken, trackID)
// 	if err != nil {
// 		fmt.Println("couldnt fetch tracks", err)
// 		return
// 	}
// 	fmt.Printf("[fetchTrack] for %+v\n ", trackID)
// 	w.Write(fmt.Appendf(nil, "fetchTrack %#v", fetchTrack))
// }
//
// type GetSpotifyApiAlbum struct {
// 	accessToken string
// }
//
// func NewGetSpotifyApiAlbum(accessToken string) *GetSpotifyApiAlbum {
// 	return &GetSpotifyApiAlbum{accessToken}
// }
// func (h *GetSpotifyApiAlbum) ServeHttp(w http.ResponseWriter, r *http.Request) {
// 	albumID := "4aawyAB9vmqN3uQ7FjRGTy"
// 	fetchAlbum, _, err := spotify.FetchAlbum(h.accessToken, albumID)
// 	if err != nil {
// 		fmt.Println("couldnt fetch albums", err)
// 		return
// 	}
// 	trackList, err := spotify.FetchTracksFromAlbum(h.accessToken, albumID)
// 	if err != nil {
// 		fmt.Println("couldnt fetch tracks from albums: ", err)
// 		return
// 	}
//
// 	w.Write(fmt.Appendf(nil, "fetchAlbum %#v with tracks: %#v ", fetchAlbum, strings.Join(trackList, ", ")))
// }
//
// type GetSpotifyApiArtist struct {
// 	accessToken string
// }
//
// func NewGetSpotifyApiArtist(accessToken string) *GetSpotifyApiArtist {
// 	return &GetSpotifyApiArtist{accessToken}
// }
// func (h *GetSpotifyApiArtist) ServeHttp(w http.ResponseWriter, r *http.Request) {
// 	artistID := "0TnOYISbd1XYRBk9myaseg"
// 	fetchArtist, err := spotify.FetchArtist(h.accessToken, artistID)
// 	if err != nil {
// 		fmt.Println("couldnt fetch artists", err)
// 		return
// 	}
//
// 	albumList, err := spotify.FetchAlbumsFromArtist(h.accessToken, artistID)
// 	if err != nil {
// 		fmt.Println("couldnt fetch tracks from albums: ", err)
// 		return
// 	}
// 	w.Write(fmt.Appendf(nil, "fetchArtist %#v, it has albums: %#v", fetchArtist, strings.Join(albumList, ", ")))
// }
//
// type GetSpotifyApiPlaylist struct {
// 	accessToken string
// }
//
// func NewGetSpotifyApiPlaylist(accessToken string) *GetSpotifyApiPlaylist {
// 	return &GetSpotifyApiPlaylist{accessToken}
// }
// func (h *GetSpotifyApiPlaylist) ServeHttp(w http.ResponseWriter, r *http.Request) {
// 	playlistID := "3cEYpjA9oz9GiPac4AsH4n"
// 	fetchPlaylist, _, err := spotify.FetchPlaylist(h.accessToken, playlistID)
// 	if err != nil {
// 		fmt.Println("couldnt fetch playlists", err)
// 		return
// 	}
// 	trackList, err := spotify.FetchTracksFromPlaylist(h.accessToken, playlistID)
// 	if err != nil {
// 		fmt.Println("couldnt fetch tracks from playlist: ", err)
// 		return
// 	}
// 	w.Write(fmt.Appendf(nil, "fetchplaylist %#v, with tracks: %#v", fetchPlaylist, strings.Join(trackList, ", ")))
// }
//
// type GetSpotifyApiTrackName struct {
// 	accessToken string
// }
//
// func NewGetSpotifyApiTrackName(accessToken string) *GetSpotifyApiTrackName {
// 	return &GetSpotifyApiTrackName{accessToken}
// }
// func (h *GetSpotifyApiTrackName) ServeHttp(w http.ResponseWriter, r *http.Request) {
// 	trackName := "Re forro"
// 	fetchTrackName, err := spotify.SearchTracksByName(h.accessToken, trackName)
// 	if err != nil {
// 		fmt.Println("couldnt fetch tracks", err)
// 		return
// 	}
// 	fmt.Printf("[fetchTrackName]Searched %v\n", trackName)
// 	w.Write(fmt.Appendf(nil, "fetchTrackName %#v", fetchTrackName))
// }
//
// type GetSpotifyApiAlbumName struct {
// 	accessToken string
// }
//
// func NewGetSpotifyApiAlbumName(accessToken string) *GetSpotifyApiAlbumName {
// 	return &GetSpotifyApiAlbumName{accessToken}
// }
// func (h *GetSpotifyApiAlbumName) ServeHttp(w http.ResponseWriter, r *http.Request) {
// 	albumName := "Papota"
// 	fetchAlbumName, err := spotify.SearchAlbumsByName(h.accessToken, albumName)
// 	if err != nil {
// 		fmt.Println("couldnt fetch albums", err)
// 		return
// 	}
// 	fmt.Printf("[fetchAlbumName] for %+v\n", albumName)
// 	w.Write(fmt.Appendf(nil, "fetchAlbumName %#v", fetchAlbumName))
// }
//
// type GetSpotifyApiArtistName struct {
// 	accessToken string
// }
//
// func NewGetSpotifyApiArtistName(accessToken string) *GetSpotifyApiArtistName {
// 	return &GetSpotifyApiArtistName{accessToken}
// }
// func (h *GetSpotifyApiArtistName) ServeHttp(w http.ResponseWriter, r *http.Request) {
// 	artistName := "Lady Gaga"
// 	fetchArtistName, err := spotify.SearchArtistsByName(h.accessToken, artistName)
// 	if err != nil {
// 		fmt.Println("couldnt fetch artists", err)
// 		return
// 	}
// 	fmt.Printf("[fetchArtistName] for %v\n", artistName)
// 	w.Write(fmt.Appendf(nil, "fetchArtistName %#v", fetchArtistName))
// }
//
// type GetSpotifyApiPlaylistName struct {
// 	accessToken string
// }
//
// func NewGetSpotifyApiPlaylistName(accessToken string) *GetSpotifyApiPlaylistName {
// 	return &GetSpotifyApiPlaylistName{accessToken}
// }
// func (h *GetSpotifyApiPlaylistName) ServeHttp(w http.ResponseWriter, r *http.Request) {
// 	playlistName := "Salsa"
// 	fetchPlaylistName, err := spotify.SearchPlaylistsByName(h.accessToken, playlistName)
// 	if err != nil {
// 		fmt.Println("couldnt fetch playlists", err)
// 		return
// 	}
// 	fmt.Printf("fetchPlaylistName for %v\n", playlistName)
// 	w.Write(fmt.Appendf(nil, "fetchPlaylistName %#v", fetchPlaylistName))
// }
