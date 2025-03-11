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
