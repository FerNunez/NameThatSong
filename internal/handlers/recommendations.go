package handlers

import (
	"goth/internal/templates"
	"net/http"
)

func IndexHttp(w http.ResponseWriter, r *http.Request) {

	component := templates.IndexPage()
	layout := templates.Layout(component, "Search")
	layout.Render(r.Context(), w)
	//component.Render(r.Context(), w)
}
