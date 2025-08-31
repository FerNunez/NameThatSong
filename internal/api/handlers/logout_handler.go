package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/FerNunez/NameThatSong/internal/services/user"
)

type PostLogoutHandler struct {
	UserService       user.UserService
	sessionCookieName string
}

func NewPostLogoutHandler(userService user.UserService, sessionCookieName string) *PostLogoutHandler {
	return &PostLogoutHandler{
		UserService:       userService,
		sessionCookieName: sessionCookieName,
	}
}

func (h *PostLogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Get session ID from cookie to properly logout
	cookie, err := r.Cookie(h.sessionCookieName)
	if err == nil && cookie.Value != "" {
		// Decode the session cookie to get session ID
		cookieData, err := base64.StdEncoding.DecodeString(cookie.Value)
		if err == nil {
			parts := strings.Split(string(cookieData), ":")
			if len(parts) == 2 {
				sessionID := parts[0]
				// Logout through service to properly revoke session
				if err := h.UserService.Logout(r.Context(), sessionID); err != nil {
					fmt.Println("error during logout:", err)
				}
			}
		}
	}

	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:    h.sessionCookieName,
		MaxAge:  -1,
		Expires: time.Now().Add(-100 * time.Hour),
		Path:    "/",
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
