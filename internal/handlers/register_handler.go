package handlers

import (
	"fmt"
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/database"
	"github.com/FerNunez/NameThatSong/internal/templates"
)

type PostRegisterHandler struct {
	dbQuery *database.Queries
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
		dbQuery,
	}
}

func (h PostRegisterHandler) ServeHttp(w http.ResponseWriter, r *http.Request) {

	username := r.FormValue("username")
	password := r.FormValue("password")
	fmt.Printf("Received a username: %v and pass %v\n", username, password)

	dbUser, err := h.dbQuery.CreateUser(r.Context(), database.CreateUserParams{
		Email:          username,
		HashedPassword: password,
	})

	fmt.Println("im here with dbUser", dbUser)

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
