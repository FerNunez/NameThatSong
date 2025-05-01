package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/manager"
	"github.com/FerNunez/NameThatSong/internal/templates"
)

type GetIndexHandler struct {
	gm *manager.GameManager
}

func NewGetIndexHandler(gm *manager.GameManager) *GetIndexHandler {
	return &GetIndexHandler{gm}
}

func (h GetIndexHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {

	game, err := h.gm.GetGame(r.Context())
	if err != nil {
		fmt.Println("cant get games, ", err)
		component := templates.IndexPage(nil)
		layout := templates.Layout(component, "NameThatSong")
		layout.Render(r.Context(), w)
		return
	}

	component := templates.IndexPage(game)
	layout := templates.Layout(component, "NameThatSong")
	layout.Render(r.Context(), w)

}
