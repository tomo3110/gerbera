package live

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/tomo3110/gerbera"
)

// Session holds a single LiveView connection's state.
type Session struct {
	ID       string
	View     View
	tree     *gerbera.Element
	lastSeen time.Time
	mu       sync.Mutex
}

type sessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	ttl      time.Duration
}

func newSessionStore(ttl time.Duration) *sessionStore {
	s := &sessionStore{
		sessions: make(map[string]*Session),
		ttl:      ttl,
	}
	go s.gc()
	return s
}

func (s *sessionStore) create(view View) *Session {
	id := generateID()
	sess := &Session{
		ID:       id,
		View:     view,
		lastSeen: time.Now(),
	}
	s.mu.Lock()
	s.sessions[id] = sess
	s.mu.Unlock()
	return sess
}

func (s *sessionStore) get(id string) *Session {
	s.mu.RLock()
	sess := s.sessions[id]
	s.mu.RUnlock()
	if sess != nil {
		sess.mu.Lock()
		sess.lastSeen = time.Now()
		sess.mu.Unlock()
	}
	return sess
}

func (s *sessionStore) remove(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

func (s *sessionStore) gc() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for id, sess := range s.sessions {
			sess.mu.Lock()
			if now.Sub(sess.lastSeen) > s.ttl {
				delete(s.sessions, id)
			}
			sess.mu.Unlock()
		}
		s.mu.Unlock()
	}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
