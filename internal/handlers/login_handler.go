package handlers

import (
	b64 "encoding/base64"

	"fmt"
	"net/http"
	"time"

	"github.com/FerNunez/NameThatSong/internal/auth"
	"github.com/FerNunez/NameThatSong/internal/manager"
	"github.com/FerNunez/NameThatSong/internal/store"
	"github.com/FerNunez/NameThatSong/internal/store/database"
	"github.com/FerNunez/NameThatSong/internal/templates"
)

type GetLoginHandler struct{}

func NewGetLoginHandler() *GetLoginHandler {
	return &GetLoginHandler{}
}

func (h GetLoginHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {

	t := templates.Login("Login")
	err := templates.Layout(t, "NameThanSong").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}

}

type PostLoginHandler struct {
	UserStore         store.UserStore
	SessionStore      store.SessionStore
	SpotifyTokenStore store.SpotifyTokenStore
	SessionName       string
	GameManager       *manager.GameManager

	//dbQuery   *database.Queries
	// sessionStore      store.SessionStore
	// passwordhash      hash.PasswordHash
	// sessionCookieName string
}

func NewPostLoginHandler(dbQuery *database.Queries, sessionName string, gm *manager.GameManager) *PostLoginHandler {
	return &PostLoginHandler{
		UserStore:         store.NewSQLUserStore(dbQuery),
		SessionStore:      store.NewSQLSessionStore(dbQuery),
		SpotifyTokenStore: store.NewSQLSpotifyTokenStore(dbQuery),
		SessionName:       sessionName,
		GameManager:       gm,
	}
}

func (h PostLoginHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {

	email := r.FormValue("email")
	password := r.FormValue("password")

	dbUser, err := h.UserStore.GetByEmail(r.Context(), email)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		c := templates.LoginError()
		c.Render(r.Context(), w)
		return
	}

	// Check Password
	if err := auth.CheckPasswordHash(password, dbUser.HashedPassword); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		c := templates.LoginError()
		c.Render(r.Context(), w)
		return
	}

	// Login without game
	if _, err := h.GameManager.GetGame(r.Context()); err != nil {
		fmt.Println("recreating game for user: ", dbUser.ID.String())
		h.GameManager.CreateGame(dbUser.ID, h.SpotifyTokenStore)
	}

	ttl := time.Duration(24 * time.Hour)
	dbSession, err := h.SessionStore.Create(r.Context(), dbUser.ID, ttl)
	if err != nil {
		fmt.Printf("error creating session!: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		c := templates.LoginError()
		c.Render(r.Context(), w)
		return
	}

	cookieValue := b64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%s:%s", dbSession.ID, dbSession.UserID.String()))
	cookie := http.Cookie{
		Name:     h.SessionName,
		Value:    cookieValue,
		Expires:  dbSession.ExpiresAt,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	}
	http.SetCookie(w, &cookie)

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}
