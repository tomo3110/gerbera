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
	ID        string
	CSRFToken string
	View      View
	tree      gerbera.Node
	lastSeen  time.Time
	mu        sync.Mutex
	infoCh    chan any // channel for HandleInfo messages
	stopTick  chan struct{}
}

type sessionStore struct {
	mu        sync.RWMutex
	sessions  map[string]*Session
	ttl       time.Duration
	onExpired func(id string)
}

func newSessionStore(ttl time.Duration, onExpired func(string)) *sessionStore {
	s := &sessionStore{
		sessions:  make(map[string]*Session),
		ttl:       ttl,
		onExpired: onExpired,
	}
	go s.gc()
	return s
}

func (s *sessionStore) create(view View) *Session {
	id := generateID()
	sess := &Session{
		ID:        id,
		CSRFToken: generateID(),
		View:      view,
		lastSeen:  time.Now(),
		infoCh:    make(chan any, 32),
		stopTick:  make(chan struct{}),
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
				if s.onExpired != nil {
					s.onExpired(id)
				}
			}
			sess.mu.Unlock()
		}
		s.mu.Unlock()
	}
}

// SendInfo sends a message to the session's info channel.
// The message will be delivered to HandleInfo if the View implements InfoReceiver.
func (s *Session) SendInfo(msg any) {
	select {
	case s.infoCh <- msg:
	default:
	}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
