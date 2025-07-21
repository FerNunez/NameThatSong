package handlers

import (
	"net/http"

	"github.com/FerNunez/NameThatSong/web/templates"
)

type GetIndexHandler struct {
}

func NewGetIndexHandler() *GetIndexHandler {
	return &GetIndexHandler{}
}

func (h GetIndexHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {

	component := templates.IndexPage()
	layout := templates.Layout(component, "NameThatSong")
	layout.Render(r.Context(), w)

}
