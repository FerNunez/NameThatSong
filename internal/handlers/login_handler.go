package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/auth"
	"github.com/FerNunez/NameThatSong/internal/store"
	"github.com/FerNunez/NameThatSong/internal/store/database"
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
	UserStore store.UserStore
	//dbQuery   *database.Queries
	// sessionStore      store.SessionStore
	// passwordhash      hash.PasswordHash
	// sessionCookieName string
}

func NewPostLoginHandler(dbQuery *database.Queries) *PostLoginHandler {
	return &PostLoginHandler{
		UserStore: store.NewSQLUserStore(dbQuery),
	}
}

func (h PostLoginHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {

	email := r.FormValue("email")
	password := r.FormValue("password")
	fmt.Printf("Received a email: %v and pass %v\n", email, password)
	dbUser, err := h.UserStore.GetEmail(r.Context(), email)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		c := templates.LoginError()
		c.Render(r.Context(), w)
		return
	}

	// Check Password
	if err := auth.CheckPasswordHash(password, dbUser.HashedPassword); err != nil {
		errmsg := fmt.Sprintln("Wrong password")
		fmt.Println(errmsg)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(errmsg))
		return
	}

	fmt.Println("Todo rest :", dbUser)

	// TODO: add here redirect to main page?
}
