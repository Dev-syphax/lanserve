package server

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
)

func GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (s *Server) Authenticate(r *http.Request) (bool, string) {
	if s.SecretCode == "" {
		return true, ""
	}

	if cookie, err := r.Cookie("lanserve_session"); err == nil {
		if s.Sessions.Has(cookie.Value) {
			return true, ""
		}
	}

	if r.Header.Get("X-Auth-Code") == s.SecretCode {
		newToken := GenerateToken()
		s.Sessions.Add(newToken)
		return true, newToken
	}

	return false, ""
}
