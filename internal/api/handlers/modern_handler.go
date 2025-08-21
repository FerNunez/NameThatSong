package handlers

import (
	"net/http"

	"github.com/a-h/templ"
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
	var indexComponent templ.Component
	
	if ok && user != nil {
		templateUser = &templates.User{
			Name:      user.DisplayName,
			Email:     user.Email,
			AvatarURL: user.AvatarURL,
		}
		// Show the full modern interface for authenticated users
		indexComponent = templates.ModernIndexPage()
	} else {
		// Show welcome page for unauthenticated users
		indexComponent = templates.WelcomeIndexPage()
	}
	
	// Wrap it in the modern layout with user data
	component := templates.ModernLayoutWithUser(indexComponent, "NameThatSong - Modern Interface", templateUser)
	
	// Render the component
	err := component.Render(r.Context(), w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}