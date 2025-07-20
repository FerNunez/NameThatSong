package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	b64 "encoding/base64"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/services/user"
)

type AuthMiddleware struct {
	userService       user.UserService
	sessionCookieName string
}

func NewAuthMiddleware(userService user.UserService, sessionCookieName string) *AuthMiddleware {
	return &AuthMiddleware{
		userService:       userService,
		sessionCookieName: sessionCookieName,
	}
}

var UserKey string = "user"

// Gets Cookie -> Validates session and gets user from UserService
func (m *AuthMiddleware) AddUserToCtxt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sessionCookie, err := r.Cookie(m.sessionCookieName)
		if err != nil {
			fmt.Println("[middleware][AddUserToCtxt] Could not retrieve session from cookie")
			next.ServeHTTP(w, r)
			return
		}

		decodedValue, err := b64.StdEncoding.DecodeString(sessionCookie.Value)
		if err != nil {
			fmt.Println("[middleware][AddUserToCtxt] Could not decode session cookie value")
			next.ServeHTTP(w, r)
			return
		}
		
		splitValue := strings.Split(string(decodedValue), ":")
		if len(splitValue) != 2 {
			fmt.Println("[middleware][AddUserToCtxt] Invalid session cookie format")
			next.ServeHTTP(w, r)
			return
		}

		sessionID := splitValue[0]
		
		// Use UserService to validate session and get user
		user, err := m.userService.ValidateSession(r.Context(), sessionID)
		if err != nil {
			fmt.Println("[middleware][AddUserToCtxt] Session validation failed:", err)
			next.ServeHTTP(w, r)
			return
		}

		fmt.Println("[middleware] User added to context for user_id:", user.ID.String())
		ctx := context.WithValue(r.Context(), UserKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUser(ctx context.Context) (*models.User, bool) {
	user := ctx.Value(UserKey)
	if user == nil {
		return nil, false
	}
	return user.(*models.User), true
}
