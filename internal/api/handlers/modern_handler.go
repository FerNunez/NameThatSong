package handlers

import (
	"net/http"

	"github.com/FerNunez/NameThatSong/internal/api/middleware"
	"github.com/FerNunez/NameThatSong/web/templates"
)

type ModernHandler struct{}

func NewModernHandler() *ModernHandler {
	return &ModernHandler{}
}

func (h *ModernHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUser(r.Context())
	var templateUser *templates.User
	if ok && user != nil {
		templateUser = &templates.User{
			Name:      user.DisplayName,
			Email:     user.Email,
			AvatarURL: user.AvatarURL,
		}
	}
	
	// Create the modern index page component
	indexComponent := templates.ModernIndexPage()
	
	// Wrap it in the modern layout with user data
	component := templates.ModernLayoutWithUser(indexComponent, "NameThatSong - Modern Interface", templateUser)
	
	// Render the component
	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}