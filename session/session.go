package session

import (
	"net/http"
	"sync"
	"time"
)

// Session holds per-user session data.
type Session struct {
	ID        string
	Values    map[string]any
	IsNew     bool
	modified  bool
	expiresAt time.Time
	mu        sync.RWMutex
}

// Get returns the value for the given key, or nil if not found.
func (s *Session) Get(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Values[key]
}

// GetString returns the value for the given key as a string.
// Returns empty string if not found or not a string.
func (s *Session) GetString(key string) string {
	v := s.Get(key)
	if v == nil {
		return ""
	}
	str, ok := v.(string)
	if !ok {
		return ""
	}
	return str
}

// Set sets a value in the session.
func (s *Session) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Values[key] = value
	s.modified = true
}

// Delete removes a value from the session.
func (s *Session) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Values, key)
	s.modified = true
}

// Clear removes all values from the session.
func (s *Session) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Values = make(map[string]any)
	s.modified = true
}

// Modified reports whether the session has been modified since it was loaded.
func (s *Session) Modified() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.modified
}

// Store is the interface for session persistence backends.
type Store interface {
	// Get retrieves the session associated with the request.
	// If no session exists, a new one is created with IsNew set to true.
	Get(r *http.Request) (*Session, error)

	// Save persists the session and writes the session cookie.
	Save(w http.ResponseWriter, r *http.Request, sess *Session) error

	// Destroy removes the session from the store and clears the cookie.
	Destroy(w http.ResponseWriter, r *http.Request, sess *Session) error
}
