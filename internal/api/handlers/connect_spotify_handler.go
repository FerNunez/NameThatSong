package handlers

import (
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/api/middleware"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
	"github.com/FerNunez/NameThatSong/web/templates"
)

type ConnectSpotifyHandler struct{}

func NewConnectSpotifyHandler() *ConnectSpotifyHandler {
	return &ConnectSpotifyHandler{}
}

func (h *ConnectSpotifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger.Info(r.Context(), "connect spotify page requested")

	// Check if user is authenticated
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		logger.Warn(r.Context(), "connect spotify page requested without authenticated user")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// If user is already connected, redirect to home
	if user.SpotifyConnected {
		logger.Info(r.Context(), "user already connected to spotify, redirecting to home",
			logger.F("user_id", user.ID.String()))
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	logger.Info(r.Context(), "rendering connect spotify page",
		logger.F("user_id", user.ID.String()))

	// Render the connect spotify page with auth layout
	component := templates.ConnectSpotifyPage()
	layout := templates.AuthLayout(component, "Connect to Spotify - NameThatSong")
	if err := layout.Render(r.Context(), w); err != nil {
		logger.Error(r.Context(), "failed to render connect spotify page",
			logger.F("user_id", user.ID.String()),
			logger.F("error", err))
		http.Error(w, "failed to render page", http.StatusInternalServerError)
		return
	}
}
