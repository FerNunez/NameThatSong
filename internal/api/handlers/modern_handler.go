package handlers

import (
	"net/http"

	"github.com/FerNunez/NameThatSong/web/templates"
)

type ModernHandler struct{}

func NewModernHandler() *ModernHandler {
	return &ModernHandler{}
}

func (h *ModernHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Create the modern index page component
	indexComponent := templates.ModernIndexPage()
	
	// Wrap it in the modern layout
	component := templates.ModernLayout(indexComponent, "NameThatSong - Modern Interface")
	
	// Render the component
	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}