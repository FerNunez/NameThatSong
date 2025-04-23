package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/database"
	"github.com/FerNunez/NameThatSong/internal/templates"
)

func (h GameHandler) GetLoginHandler(w http.ResponseWriter, r *http.Request) {

	t := templates.Login("Login")
	err := templates.Layout(t, "NameThanSong").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}

}

type PostLoginHandler struct {
	dbQuery *database.Queries
	// sessionStore      store.SessionStore
	// passwordhash      hash.PasswordHash
	// sessionCookieName string
}

func NewPostLoginHandler(dbQuery *database.Queries) *PostLoginHandler {
	return &PostLoginHandler{
		dbQuery,
	}
}

func (h PostLoginHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {

	username := r.FormValue("username")
	password := r.FormValue("password")
	fmt.Printf("Received a username: %v and pass %v\n", username, password)
	dbUser, err := h.dbQuery.GetUserByEmail(r.Context(), username)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		c := templates.LoginError()
		c.Render(r.Context(), w)
		return
	}
	fmt.Println("Todo rest :", dbUser)

	// TODO: add here redirect to main page?
}
