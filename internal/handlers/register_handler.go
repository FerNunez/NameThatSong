package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/auth"
	"github.com/FerNunez/NameThatSong/internal/store"
	"github.com/FerNunez/NameThatSong/internal/store/database"
	"github.com/FerNunez/NameThatSong/internal/templates"
)

type PostRegisterHandler struct {
	UserStore store.UserStore
	//dbQuery *database.Queries
	// sessionStore      store.SessionStore
	// passwordhash      hash.PasswordHash
	// sessionCookieName string
}

func (h GameHandler) GetRegisterHandler(w http.ResponseWriter, r *http.Request) {
	t := templates.RegisterPage()
	err := templates.Layout(t, "NameThanSong").Render(r.Context(), w)

	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func NewPostRegisterHandler(dbQuery *database.Queries) *PostRegisterHandler {
	return &PostRegisterHandler{
		UserStore: store.NewSQLUserStore(dbQuery),
	}
}

func (h PostRegisterHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {

	email := r.FormValue("email")
	password := r.FormValue("password")
	fmt.Printf("Received a email: %v and pass %v\n", email, password)

	hashedPass, err := auth.HashPassword(password)

	dbUser, err := h.UserStore.Create(r.Context(), email, hashedPass)

	fmt.Println("im here with dbUser", dbUser.Email)

	if err != nil {

		fmt.Println("error creating user err:", err)
		w.WriteHeader(http.StatusBadRequest)
		c := templates.RegisterError()
		c.Render(r.Context(), w)
		return
	}

	c := templates.RegisterSuccess()
	err = c.Render(r.Context(), w)

	if err != nil {
		http.Error(w, "error rendering template", http.StatusInternalServerError)
		return
	}

	// TODO: add here redirect to main page?
}
