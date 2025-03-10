package handlers

import (
	"goth/internal/templates"
	"net/http"
)

var allArtists = []string{
	"Taylor Swift",
	"Ed Sheeran",
	"Adele",
	"Drake",
	"Beyoncé",
	"Rihanna",
	"The Weeknd",
	"Bruno Mars",
	"Katy Perry",
}

var albums = []templates.Album{
	{ID: 1, Title: "Thriller", Artist: "Michael Jackson", Year: 1982},
	{ID: 2, Title: "Rumours", Artist: "Fleetwood Mac", Year: 1977},
	{ID: 3, Title: "Back in Black", Artist: "AC/DC", Year: 1980},
	{ID: 4, Title: "The Dark Side of the Moon", Artist: "Pink Floyd", Year: 1973},
	{ID: 5, Title: "Born to Run", Artist: "Bruce Springsteen", Year: 1975},
	{ID: 6, Title: "The Dark Side of the Moon", Artist: "Pink Floyd", Year: 1973},
	{ID: 7, Title: "Born to Run", Artist: "Bruce Springsteen", Year: 1975},
}

func IndexHttp(w http.ResponseWriter, r *http.Request) {

	component := templates.IndexPage()
	layout := templates.Layout(component, "Search")
	layout.Render(r.Context(), w)
	//component.Render(r.Context(), w)
}

//func (api *ApiConfig) FilterRecomendationHttp(w http.ResponseWriter, r *http.Request) {
//	query := strings.ToLower(r.URL.Query().Get("search"))
//
//	//  TODO: get artist list from search
//	var results []string
//	if query != "" {
//		for _, item := range allArtists {
//			if strings.Contains(strings.ToLower(item), query) {
//				results = append(results, item)
//			}
//		}
//	}
//
//	// Return the search results component
//	component := templates.SearchResults(results)
//	component.Render(r.Context(), w)
//}

//func (api *ApiConfig) StartProcess(w http.ResponseWriter, r *http.Request) {
//
//	err := r.ParseForm()
//	if err != nil {
//		http.Error(w, "Error parsing form", http.StatusBadRequest)
//		return
//	}
//
//	// TODO: Get selected album IDs
//	selectedIDs := r.Form["selectedAlbums"]
//	fmt.Println("q: ", selectedIDs)
//
//	// 3. Convert to your album type
//	var selectedAlbums []templates.Album
//
//	output := ""
//	for _, id := range selectedIDs {
//		// Find matching album in your data store
//		idInt, _ := strconv.Atoi(id)
//		for _, album := range albums {
//			if album.ID == idInt {
//				selectedAlbums = append(selectedAlbums, album)
//				output += fmt.Sprintf("%v\n", album.Title)
//				break
//			}
//		}
//	}
//
//	// 4. Process selected albums (your logic here)
//	fmt.Printf("Processing albums: %+v\n", selectedAlbums)
//
//	musicPlayer := templates.MusicPlayer()
//	layout := templates.GuesserLayout(musicPlayer, "game")
//	layout.Render(r.Context(), w)
//
//}
