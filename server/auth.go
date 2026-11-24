package server

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
)

func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
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

	// Constant-time comparison so response timing doesn't leak the code.
	supplied := r.Header.Get("X-Auth-Code")
	if subtle.ConstantTimeCompare([]byte(supplied), []byte(s.SecretCode)) == 1 {
		newToken, err := GenerateToken()
		if err != nil {
			return false, ""
		}
		s.Sessions.Add(newToken)
		return true, newToken
	}

	return false, ""
}
