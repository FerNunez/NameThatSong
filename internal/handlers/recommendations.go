package handlers

import (
	"goth/internal/templates"
	"net/http"
)

func IndexHttp(w http.ResponseWriter, r *http.Request) {

	component := templates.IndexPage()
	layout := templates.Layout(component, "Search")
	layout.Render(r.Context(), w)
	//layout := templates.Layout("Layout")
	//layout.Render(r.Context(), w)

}
