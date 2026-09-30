package server

import "sync"

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]bool
}

func NewSessionStore() *SessionStore {
	return &SessionStore{sessions: make(map[string]bool)}
}

func (s *SessionStore) Add(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[token] = true
}

func (s *SessionStore) Has(token string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[token]
}
