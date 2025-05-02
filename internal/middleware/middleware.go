package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	b64 "encoding/base64"

	"github.com/FerNunez/NameThatSong/internal/store"
)

type AuthMiddleware struct {
	userStore         store.UserStore
	sessionCookieName string
	count             int
}

func NewAuthMiddleware(userStore store.UserStore, sessionCookieName string) *AuthMiddleware {
	return &AuthMiddleware{
		userStore:         userStore,
		sessionCookieName: sessionCookieName,
		count:             0,
	}
}

var UserKey string = "user"

// Gets Cookie -> Gets user from Cookie
func (m *AuthMiddleware) AddUserToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("middleware called? from", r.URL)
		sessionCookie, err := r.Cookie(m.sessionCookieName)
		if err != nil {
			fmt.Println("error")
			next.ServeHTTP(w, r)
			return
		}

		decodedValue, err := b64.StdEncoding.DecodeString(sessionCookie.Value)
		if err != nil {
			fmt.Println("error here")
			next.ServeHTTP(w, r)
			return
		}

		splitValue := strings.Split(string(decodedValue), ":")
		if len(splitValue) != 2 {
			next.ServeHTTP(w, r)
			return
		}

		//sessionID := splitValue[0]
		//userID := splitValue[1]
		user, err := m.userStore.GetById(r.Context(), splitValue[1])
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		fmt.Println("Found ID in DB")
		ctx := context.WithValue(r.Context(), UserKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUser(ctx context.Context) (store.User, bool) {
	user := ctx.Value(UserKey)
	if user == nil {
		return store.User{}, false
	}
	return user.(store.User), true
}

// // TEMPORAL STUF
//
// func (m *AuthMiddleware) CreateTempUser(next http.Handler) http.Handler {
//
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//
// 		// get session from Cookie
// 		_, err := r.Cookie(m.sessionCookieName)
// 		if err == http.ErrNoCookie {
// 			// create a cookie session for "new user"
//
// 			fmt.Println("No cookie detected :O")
// 			userName := fmt.Sprintf("User%v", m.count)
//
// 			s := store.Session{
// 				ID:        m.count,
// 				SessionID: "",
// 				UserName:  userName,
// 			}
// 			sessWithId, _ := m.userStore.CreateSession(&s)
// 			m.count += 1
//
// 			cookieValue := b64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%s:%s", sessWithId.SessionID, sessWithId.UserName))
//
// 			expiration := time.Now().Add(365 * 24 * time.Hour)
// 			cookie := http.Cookie{
// 				Name:     m.sessionCookieName,
// 				Value:    cookieValue,
// 				Expires:  expiration,
// 				Path:     "/",
// 				HttpOnly: true,
// 				SameSite: http.SameSiteStrictMode,
// 			}
// 			http.SetCookie(w, &cookie)
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }
