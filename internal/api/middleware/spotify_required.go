package middleware

import (
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
)

func RequireSpotifyConnection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUser(r.Context())
		if !ok {
			logger.Warn(r.Context(), "spotify required middleware: no authenticated user")
			// For HTMX requests, redirect via header
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		if !user.SpotifyConnected {
			logger.Info(r.Context(), "spotify required middleware: user not connected to spotify",
				logger.F("user_id", user.ID.String()))

			// For HTMX requests, redirect via header
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/connect-spotify")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			http.Redirect(w, r, "/connect-spotify", http.StatusFound)
			return
		}

		logger.Debug(r.Context(), "spotify required middleware: user has spotify connection",
			logger.F("user_id", user.ID.String()))

		next.ServeHTTP(w, r)
	})
}
