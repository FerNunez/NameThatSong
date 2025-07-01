package handlers

import (
	b64 "encoding/base64"

	"fmt"
	"net/http"
	"time"

	"github.com/FerNunez/NameThatSong/internal/pkg/jwt"
	"github.com/FerNunez/NameThatSong/internal/pkg/utils"
	"github.com/FerNunez/NameThatSong/internal/repository"
	"github.com/FerNunez/NameThatSong/internal/repository/database"
	"github.com/FerNunez/NameThatSong/internal/services/game"
	"github.com/FerNunez/NameThatSong/web/templates"
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
	UserStore         repository.UserStore
	SessionStore      repository.SessionStore
	SpotifyTokenStore repository.SpotifyTokenStore
	SessionName       string
	GameManager       *game.GameManager

	//dbQuery   *database.Queries
	// sessionStore      repository.SessionStore
	// passwordhash      hash.PasswordHash
	// sessionCookieName string
}

func NewPostLoginHandler(dbQuery *database.Queries, tokenEncryptor *utils.TokenEncryptor, sessionName string, gm *game.GameManager) *PostLoginHandler {
	return &PostLoginHandler{
		UserStore:         repository.NewSQLUserStore(dbQuery),
		SessionStore:      repository.NewSQLSessionStore(dbQuery),
		SpotifyTokenStore: repository.NewSQLSpotifyTokenStore(dbQuery, tokenEncryptor),
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
	if err := jwt.CheckPasswordHash(password, dbUser.HashedPassword); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		c := templates.LoginError()
		c.Render(r.Context(), w)
		return
	}

	// Login without game
	if _, err := h.GameManager.GetGame(r.Context()); err != nil {
		fmt.Println("[PostLoginHandler] ServeHttp: Recreating game for user", dbUser.ID.String())
		// TODO: FIX THIS FOR TH ELOVE OF G
		dbSpotifyToken, err := h.SpotifyTokenStore.Get(r.Context(), dbUser.ID)
		if err != nil {
			fmt.Println("[PostLoginHandler] ServeHttp: Spotify token not found for", dbUser.ID.String())
			h.GameManager.CreateGame(dbUser.ID, h.SpotifyTokenStore)

		} else {
			fmt.Println("[PostLoginHandler] ServeHttp: Spotify token found for", dbUser.ID.String())
			h.GameManager.CreateGame(dbUser.ID, h.SpotifyTokenStore)
			g, ok := h.GameManager.Games[string(dbUser.ID.String())]
			if ok {
				fmt.Println("[PostLoginHandler] ServeHttp: retrieved access token from db for user", dbUser.ID.String())
				g.SpotifyApi.AccessToken = dbSpotifyToken.AccessToken
				g.SpotifyApi.RefreshToken = dbSpotifyToken.RefreshToken
				// TODO: FIX THIS
			}
		}

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
