package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/api/middleware"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
	"github.com/FerNunez/NameThatSong/internal/services/playlist"
	"github.com/FerNunez/NameThatSong/internal/services/spotify"
	"github.com/FerNunez/NameThatSong/internal/services/user"
)

type GetAuthHandler struct {
	spotifyService spotify.SpotifyService
}

func NewGetAuthHandler(ss spotify.SpotifyService) *GetAuthHandler {
	return &GetAuthHandler{ss}

}
func (h *GetAuthHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {
	logger.Info(r.Context(), "spotify auth request initiated")

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		logger.Warn(r.Context(), "spotify auth request without authenticated user")
		// TODO: add here a way to say "you need to log in"
		return
	}

	logger.Info(r.Context(), "generating spotify auth URL",
		logger.F("user_id", user.ID.String()))

	urlString, err := h.spotifyService.AuthRequestURL(user.ID.String())
	if err != nil {
		logger.Error(r.Context(), "failed to generate spotify auth URL",
			logger.F("user_id", user.ID.String()),
			logger.F("error", err))
		http.Error(w, fmt.Sprintf("error generating the auth request url: %v", err), http.StatusBadRequest)
		return
	}

	logger.Info(r.Context(), "redirecting to spotify auth",
		logger.F("user_id", user.ID.String()),
		logger.F("auth_url", urlString))

	// Redirect to Spotify
	w.Header().Set("HX-Redirect", urlString)
}

// //////////////////////////////////////
type GetAuthCallbackHandler struct {
	spotifyService  spotify.SpotifyService
	userService     user.UserService
	playlistService playlist.Service
}

func NewGetAuthCallbackHandler(ss spotify.SpotifyService, us user.UserService, ps playlist.Service) *GetAuthCallbackHandler {
	return &GetAuthCallbackHandler{
		spotifyService:  ss,
		userService:     us,
		playlistService: ps,
	}
}
func (h *GetAuthCallbackHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {
	logger.Info(r.Context(), "spotify auth callback received")

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		logger.Warn(r.Context(), "spotify auth callback without authenticated user")
		// TODO: add here a way to say "you need to log in"
		return
	}
	state := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	logger.Info(r.Context(), "processing spotify token exchange",
		logger.F("user_id", user.ID.String()),
		logger.F("has_code", code != ""),
		logger.F("has_state", state != ""))

	tr, err := h.spotifyService.TokenExchange(r.Context(), user.ID.String(), code, state)
	if err != nil {
		logger.Error(r.Context(), "failed to exchange spotify token",
			logger.F("user_id", user.ID.String()),
			logger.F("error", err))
		http.Error(w, "error exchanging spotify token", http.StatusBadRequest)
		return
	}

	logger.Info(r.Context(), "spotify token exchange successful",
		logger.F("user_id", user.ID.String()),
		logger.F("token_type", tr.TokenType),
		logger.F("expires_in", tr.ExpiresIn),
		logger.F("scope", tr.Scope))

	// Mark user as Spotify connected
	err = h.userService.MarkSpotifyConnected(r.Context(), user.ID)
	if err != nil {
		logger.Error(r.Context(), "failed to mark user as spotify connected",
			logger.F("user_id", user.ID.String()),
			logger.F("error", err))
		// Continue anyway since the token exchange was successful
	} else {
		logger.Info(r.Context(), "user marked as spotify connected",
			logger.F("user_id", user.ID.String()))
	}

	// Importing playlists from spotify
	go func() {
		newCtx := context.Background()
		spotifyPlaylists, err := h.playlistService.ImportUsersPlaylistsFromSpotify(newCtx, user.ID)
		if err != nil {
			logger.Info(newCtx, "couldn't import spotify playlists into local",
				logger.F("user_id", user.ID.String()))
		}
		logger.Info(r.Context(), "imported spotify playlists into local",
			logger.F("user_id", user.ID.String()),
			logger.F("numb imported", len(spotifyPlaylists)))
	}()

	http.Redirect(w, r, "/", http.StatusFound)
}
