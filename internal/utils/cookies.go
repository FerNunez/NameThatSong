package utils

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"
)

func ClearCookie(sessionName string) http.Cookie {
	clearCookie := http.Cookie{
		Name:     sessionName,
		Value:    "",
		MaxAge:   -1, // Delete immediately
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	}
	return clearCookie
}

func GenerateCookie(sessionID, userID, sessionName string, expiresAt time.Time) http.Cookie {
	cookieValue := base64.StdEncoding.EncodeToString(fmt.Appendf(nil, "%s:%s", sessionID, userID))
	cookie := http.Cookie{
		Name:     sessionName,
		Value:    cookieValue,
		Expires:  expiresAt,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	}
	return cookie
}
