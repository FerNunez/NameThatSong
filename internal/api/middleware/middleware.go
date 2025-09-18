package middleware

import (
	"context"
	"net/http"
	"strings"

	b64 "encoding/base64"

	"github.com/FerNunez/NameThatSong/internal/models"
	"github.com/FerNunez/NameThatSong/internal/pkg/logger"
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

type contextKey string

const UserKey contextKey = "user"

// Gets Cookie -> Validates session and gets user from UserService
func (m *AuthMiddleware) AddUserToCtxt(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		sessionCookie, err := r.Cookie(m.sessionCookieName)
		if err != nil {
			logger.Debug(r.Context(), "no session cookie found")
			next.ServeHTTP(w, r)
			return
		}

		decodedValue, err := b64.StdEncoding.DecodeString(sessionCookie.Value)
		if err != nil {
			logger.Warn(r.Context(), "could not decode session cookie, clearing invalid cookie")
			m.clearSessionCookie(w)
			next.ServeHTTP(w, r)
			return
		}

		splitValue := strings.Split(string(decodedValue), ":")
		if len(splitValue) != 2 {
			logger.Warn(r.Context(), "invalid session cookie format, clearing cookie")
			m.clearSessionCookie(w)
			next.ServeHTTP(w, r)
			return
		}

		sessionID := splitValue[0]
		expectedUserID := splitValue[1]

		logger.Debug(r.Context(), "validating session from cookie",
			logger.F("session_id", sessionID),
			logger.F("expected_user_id", expectedUserID))

		// Use UserService to validate session and get user
		user, err := m.userService.ValidateSession(r.Context(), sessionID)
		if err != nil {
			logger.Info(r.Context(), "session validation failed, clearing invalid session cookie",
				logger.F("session_id", sessionID),
				logger.F("error", err))
			m.clearSessionCookie(w)
			next.ServeHTTP(w, r)
			return
		}

		// Additional validation: ensure user ID matches what's in the cookie
		if user.ID.String() != expectedUserID {
			logger.Warn(r.Context(), "user ID mismatch between session and cookie, clearing cookie",
				logger.F("session_user_id", user.ID.String()),
				logger.F("cookie_user_id", expectedUserID))
			m.clearSessionCookie(w)
			next.ServeHTTP(w, r)
			return
		}

		logger.Debug(r.Context(), "session validation successful, user added to context",
			logger.F("user_id", user.ID.String()),
			logger.F("session_id", sessionID))

		ctx := context.WithValue(r.Context(), UserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// clearSessionCookie clears the session cookie when it's invalid
func (m *AuthMiddleware) clearSessionCookie(w http.ResponseWriter) {
	clearCookie := http.Cookie{
		Name:     m.sessionCookieName,
		Value:    "",
		MaxAge:   -1, // Delete immediately
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	}
	http.SetCookie(w, &clearCookie)
}

func GetUser(ctx context.Context) (*models.User, bool) {
	user := ctx.Value(UserKey)
	if user == nil {
		return nil, false
	}
	return user.(*models.User), true
}
