package handlers

import (
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/api/middleware"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
	"github.com/FerNunez/NameThatSong/web/templates"
)

type GetIndexHandler struct {
}

func NewGetIndexHandler() *GetIndexHandler {
	return &GetIndexHandler{}
}

func (h GetIndexHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	user, ok := middleware.GetUser(r.Context())
	if ok && !user.SpotifyConnected {
		// User is logged in but not connected to Spotify
		logger.Info(r.Context(), "redirecting user to connect spotify",
			logger.F("user_id", user.ID.String()))
		http.Redirect(w, r, "/connect-spotify", http.StatusFound)
		return
	}

	component := templates.IndexPage()
	layout := templates.Layout(component, "NameThatSong")
	layout.Render(r.Context(), w)
}
