package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/services/game"
	"github.com/FerNunez/NameThatSong/web/templates"
)

type GetIndexHandler struct {
	gm *game.GameManager
}

func NewGetIndexHandler(gm *game.GameManager) *GetIndexHandler {
	return &GetIndexHandler{gm}
}

func (h GetIndexHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {

	game, err := h.gm.GetGame(r.Context())
	if err != nil {
		fmt.Println("[GetIndexHandler] Cant get game in context:", err)
		component := templates.IndexPage(nil)
		layout := templates.Layout(component, "NameThatSong")
		layout.Render(r.Context(), w)
		return
	}

	component := templates.IndexPage(game)
	layout := templates.Layout(component, "NameThatSong")
	layout.Render(r.Context(), w)

}
