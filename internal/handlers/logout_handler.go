package handlers

import (
	"net/http"
	"time"
)

type PostLogoutHandler struct {
	sessionCookieName string
}

func NewPostLogoutHandler(sessionCookieName string) *PostLogoutHandler {
	return &PostLogoutHandler{sessionCookieName: sessionCookieName}
}

func (h *PostLogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	http.SetCookie(w, &http.Cookie{
		Name:    h.sessionCookieName,
		MaxAge:  -1,
		Expires: time.Now().Add(-100 * time.Hour),
		Path:    "/",
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
